package main

import (
	"github.com/sarulabs/dingo"
	"github.com/softonic/homing-pigeon/pkg/generatedServices/dic"
	"k8s.io/klog"
)

func main() {
	klog.InitFlags(nil)

	container, err := dic.NewContainer(dingo.App)
	if err != nil {
		panic(err)
	}
	reader := container.GetReader()
	middleware := container.GetMiddleware()
	ackManager := container.GetAckManager()
	writer := container.GetWriter()

	go reader.Start()
	go middleware.Start()
	go ackManager.Start()
	writer.Start()
}
