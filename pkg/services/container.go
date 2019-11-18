package services

import (
	"bitbucket.org/softonic-development/homing-pigeon/pkg/readers"
	"bitbucket.org/softonic-development/homing-pigeon/pkg/writers"
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