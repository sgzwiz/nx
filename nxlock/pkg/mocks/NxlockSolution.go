// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// NxlockSolution is an autogenerated mock type for the NxlockSolution type
type NxlockSolution struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *NxlockSolution) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Lock provides a mock function with given fields: ctx, key, ttl
func (_m *NxlockSolution) Lock(ctx context.Context, key string, ttl int64) error {
	ret := _m.Called(ctx, key, ttl)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, int64) error); ok {
		r0 = rf(ctx, key, ttl)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Release provides a mock function with given fields: ctx, key
func (_m *NxlockSolution) Release(ctx context.Context, key string) error {
	ret := _m.Called(ctx, key)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
