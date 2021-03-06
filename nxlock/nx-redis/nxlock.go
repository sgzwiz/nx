package nx_redis

import (
	"context"
	"errors"
	"time"

	"sync"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/redis"
	"github.com/jukylin/nx/nxlock/pkg"
)

type Client struct {
	*redis.Client

	logger log.Logger

	// 续租的Key
	keepAliveKey map[string]chan struct{}

	mutex *sync.RWMutex
}

type ClientOption func(*Client)

// 分布式锁 redis > 2.6 解决方案
func NewClient(options ...ClientOption) pkg.NxlockSolution {
	rc := &Client{}

	for _, option := range options {
		option(rc)
	}

	rc.keepAliveKey = make(map[string]chan struct{})

	rc.mutex = &sync.RWMutex{}
	return rc
}

func WithLogger(logger log.Logger) ClientOption {
	return func(e *Client) {
		e.logger = logger
	}
}

func WithClient(client *redis.Client) ClientOption {
	return func(e *Client) {
		e.Client = client
	}
}

func (rc *Client) Lock(ctx context.Context, key string, ttl int64) error {
	err := rc.set(ctx, key, "1", ttl)
	if err != nil {
		rc.logger.Debugc(ctx, err.Error())
		return err
	}

	keepAliveChan := make(chan struct{}, 1)
	rc.mutex.Lock()
	rc.keepAliveKey[key] = keepAliveChan
	rc.mutex.Unlock()
	go rc.keepAlive(ctx, key, ttl, keepAliveChan)

	return nil
}

func (rc *Client) set(ctx context.Context, key, val string, ttl int64) error {
	conn := rc.GetCtxRedisConn()
	defer conn.Close()

	ok, err := redis.String(conn.Do(ctx, "set", key, val, "nx", "ex", ttl))
	if err != nil {
		return err
	}

	if ok != "OK" {
		return errors.New(pkg.ErrRedisLockFailure)
	}

	return err
}

func (rc *Client) Release(ctx context.Context, key string) error {
	err := rc.expire(ctx, key, -1)
	if err != nil {
		return err
	}

	// 释放续租协程
	rc.mutex.Lock()
	close(rc.keepAliveKey[key])
	delete(rc.keepAliveKey, key)
	rc.mutex.Unlock()

	return nil
}

// 续租
func (rc *Client) keepAlive(ctx context.Context, key string, ttl int64, keepAliveChan chan struct{}) {
	ticker := time.NewTicker(time.Duration(ttl/3) * time.Second)

	for {
		select {
		case <-ticker.C:
			err := rc.expire(ctx, key, ttl)
			if err != nil {
				rc.logger.Errorc(ctx, "Nxredis keepAlive %s : %s", key, err.Error())
			}
		case _, ok := <-keepAliveChan:
			if !ok {
				rc.logger.Infoc(ctx, "Nxredis keepAlive 关闭续租 %s", key)
				return
			}
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (rc *Client) expire(ctx context.Context, key string, ttl int64) error {
	conn := rc.GetCtxRedisConn()
	defer conn.Close()

	_, err := redis.Bool(conn.Do(ctx, "expire", key, ttl))
	if err != nil {
		return err
	}

	return nil
}

func (rc *Client) Close() error {
	for key, _ := range rc.keepAliveKey {
		close(rc.keepAliveKey[key])
		delete(rc.keepAliveKey, key)
	}
	return rc.Client.Close()
}
