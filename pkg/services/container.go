package services

import (
	"github.com/softonic/homing-pigeon/pkg/readers"
	"github.com/softonic/homing-pigeon/pkg/writers"
	"github.com/sarulabs/dingo"
)

var Container = []dingo.Def{
	{
		Name: "AmqpReader",
		Build: &readers.AmqpReader{},
		Params: dingo.Params{
			"WriteChannel": dingo.Service("WriteChannel"),
			"AckChannel": dingo.Service("AckChannel"),
		},
	},
	{
		Name: "ElasticsearchWriter",
		Build: &writers.ElasticsearchWriter{},
		Params: dingo.Params{
			"WriteChannel": dingo.Service("WriteChannel"),
			"AckChannel": dingo.Service("AckChannel"),
		},
	},
	{
		Name: "WriteChannel",
		Build: func() (*chan string, error) {
			c := make(chan string)
			return &c, nil
		},
	},
	{
		Name: "AckChannel",
		Build: func() (*chan string, error) {
			c := make(chan string)
			return &c, nil
		},
	},
}