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
	},
	{
		Name: "ElasticsearchWriter",
		Build: &writers.ElasticsearchWriter{},
	},
}