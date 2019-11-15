package services

import (
	"bitbucket.org/softonic-development/homming-pidgeon/pkg/readers"
	"bitbucket.org/softonic-development/homming-pidgeon/pkg/writers"
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
		Params: dingo.Params{
			"FieldA": "test",
			"FieldB": "value",
		},
	},
}