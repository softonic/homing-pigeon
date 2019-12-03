// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	messages "github.com/softonic/homing-pigeon/pkg/messages"
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// WriteAdapter is an autogenerated mock type for the WriteAdapter type
type WriteAdapter struct {
	mock.Mock
}

// GetTimeout provides a mock function with given fields:
func (_m *WriteAdapter) GetTimeout() time.Duration {
	ret := _m.Called()

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func() time.Duration); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	return r0
}

// ProcessMessages provides a mock function with given fields: msgs
func (_m *WriteAdapter) ProcessMessages(msgs []messages.Message) []messages.Ack {
	ret := _m.Called(msgs)

	var r0 []messages.Ack
	if rf, ok := ret.Get(0).(func([]messages.Message) []messages.Ack); ok {
		r0 = rf(msgs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]messages.Ack)
		}
	}

	return r0
}

// ShouldProcess provides a mock function with given fields: msgs
func (_m *WriteAdapter) ShouldProcess(msgs []messages.Message) bool {
	ret := _m.Called(msgs)

	var r0 bool
	if rf, ok := ret.Get(0).(func([]messages.Message) bool); ok {
		r0 = rf(msgs)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}
