// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Channel is an autogenerated mock type for the Channel type
type Channel struct {
	mock.Mock
}

// Ack provides a mock function with given fields: tag, multiple
func (_m *Channel) Ack(tag uint64, multiple bool) error {
	ret := _m.Called(tag, multiple)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint64, bool) error); ok {
		r0 = rf(tag, multiple)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Close provides a mock function with given fields:
func (_m *Channel) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Nack provides a mock function with given fields: tag, multiple, requeue
func (_m *Channel) Nack(tag uint64, multiple bool, requeue bool) error {
	ret := _m.Called(tag, multiple, requeue)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint64, bool, bool) error); ok {
		r0 = rf(tag, multiple, requeue)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
