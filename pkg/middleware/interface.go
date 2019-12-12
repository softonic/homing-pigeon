package middleware

import (
	"github.com/softonic/homing-pigeon/middleware"
)

type Middleware interface {
	middleware.MiddlewareServer
	Next(req *middleware.Data) (*middleware.Data, error)
}
