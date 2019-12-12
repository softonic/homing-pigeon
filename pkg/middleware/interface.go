package middleware

import (
	pb "github.com/softonic/homing-pigeon/middleware"
)

type Middleware interface {
	pb.MiddlewareServer
	SetClient(client *pb.MiddlewareClient)
	Next(req *pb.Data) (*pb.Data, error)
}
