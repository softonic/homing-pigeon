package middleware

import (
	"github.com/softonic/homing-pigeon/proto"
)

type Middleware interface {
	proto.MiddlewareServer
	Next(req *proto.Data) (*proto.Data, error)
}

var (
	// see https://github.com/grpc/grpc/blob/master/doc/service_config.md to know more about service config
	defaultRetryPolicy = `{
		"methodConfig": [{
		  "name": [{"service": "proto.Middleware"}],
		  "retryPolicy": {
			  "MaxAttempts": 5,
			  "InitialBackoff": "1s",
			  "MaxBackoff": "9s",
			  "BackoffMultiplier": 3.0,
			  "RetryableStatusCodes": [ "UNAVAILABLE" ]
		  }
		}]}`
)
