package services

import (
	"github.com/softonic/homing-pigeon/pkg/readers"
	"github.com/softonic/homing-pigeon/pkg/writers"
	"github.com/sarulabs/dingo"
	"github.com/softonic/homing-pigeon/pkg/writers/adapters"
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
		Name: "Writer",
		Build: &writers.Writer{},
		Params: dingo.Params{
			"WriteChannel": dingo.Service("WriteChannel"),
			"AckChannel": dingo.Service("AckChannel"),
			"WriteAdapter": dingo.Service("NopAdapter"),
		},
	},

	{
		Name: "ElasticsearchAdapter",
		Build: func() (adapters.WriteAdapter, error) {
			return &adapters.Elasticsearch{}, nil
		},
	},
	{
		Name: "NopAdapter",
		Build: func() (adapters.WriteAdapter, error) {
			return &adapters.Nop{}, nil
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