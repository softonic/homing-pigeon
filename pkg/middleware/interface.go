package middleware

import (
	"github.com/softonic/homing-pigeon/proto"
)

type Middleware interface {
	proto.MiddlewareServer
	Next(req *proto.Data) (*proto.Data, error)
}
