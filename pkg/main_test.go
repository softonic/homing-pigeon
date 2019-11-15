package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type readAdapterMock struct {
	mock.Mock
}

type writeAdapterMock struct {
	mock.Mock
}

func TestInputNotValid(t *testing.T) {
	assert := assert.New(t)

	InitAdapter()
}
