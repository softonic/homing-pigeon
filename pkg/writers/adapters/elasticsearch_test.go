package adapters

import (
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockEsClient struct {
	elasticsearch.Client
}

func TestAdapterProcessSingleMessage(t *testing.T) {
	client, err := MockElasticsearch()
	if err != nil {
		assert.Fail(t, "Elasticsearch client could not be mock", err)
	}

	esAdapter := Elasticsearch{
		FlushMaxSize:  0,
		FlushInterval: 0,
		Client:        client,
	}

	
}

func MockElasticsearch() (*elasticsearch.Client, error) {
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()

	handler = func(w http.ResponseWriter, r *http.Request) {
		resp := `{
			  "took" : 122,
			  "timed_out" : false,
			  "_shards" : {
			    "total" : 6,
			    "successful" : 5,
			    "failed" : 1,
			    "failures" : [ {
			      "shard" : 0,
			      "index" : ".kibana",
			      "node" : "jucBX9QkQIini9dLG9tZIw",
			      "reason" : {
				"type" : "search_parse_exception",
				"reason" : "No mapping found for [offset] in order to sort on"
			      }
			    } ]
			  },
			  "hits" : {
			    "total" : 10,
			    "max_score" : null,
			    "hits" : [ {
			      "_index" : "logstash-2016.07.25",
			      "_type" : "log",
			      "_id" : "AVYkNv542Gim_t2htKPU",
			      "_score" : null,
			      "_source" : {
				"message" : "Alice message 10",
				"@version" : "1",
				"@timestamp" : "2016-07-25T22:39:55.760Z",
				"source" : "/Users/yury/logs/alice.log",
				"offset" : 144,
				"type" : "log",
				"input_type" : "log",
				"count" : 1,
				"fields" : null,
				"beat" : {
				  "hostname" : "Yurys-MacBook-Pro.local",
				  "name" : "Yurys-MacBook-Pro.local"
				},
				"host" : "Yurys-MacBook-Pro.local",
				"tags" : [ "beats_input_codec_plain_applied" ],
				"app" : "alice"
			      },
			      "sort" : [ 144 ]
			    }, {
			      "_index" : "logstash-2016.07.25",
			      "_type" : "log",
			      "_id" : "AVYkNv542Gim_t2htKPT",
			      "_score" : null,
			      "_source" : {
				"message" : "Alice message 9",
				"@version" : "1",
				"@timestamp" : "2016-07-25T22:39:55.760Z",
				"source" : "/Users/yury/logs/alice.log",
				"offset" : 128,
				"input_type" : "log",
				"count" : 1,
				"beat" : {
				  "hostname" : "Yurys-MacBook-Pro.local",
				  "name" : "Yurys-MacBook-Pro.local"
				},
				"type" : "log",
				"fields" : null,
				"host" : "Yurys-MacBook-Pro.local",
				"tags" : [ "beats_input_codec_plain_applied" ],
				"app" : "alice"
			      },
			      "sort" : [ 128 ]
			    }, {
			      "_index" : "logstash-2016.07.25",
			      "_type" : "log",
			      "_id" : "AVYkNv542Gim_t2htKPR",
			      "_score" : null,
			      "_source" : {
				"message" : "Alice message 8",
				"@version" : "1",
				"@timestamp" : "2016-07-25T22:39:55.760Z",
				"type" : "log",
				"input_type" : "log",
				"source" : "/Users/yury/logs/alice.log",
				"count" : 1,
				"fields" : null,
				"beat" : {
				  "hostname" : "Yurys-MacBook-Pro.local",
				  "name" : "Yurys-MacBook-Pro.local"
				},
				"offset" : 112,
				"host" : "Yurys-MacBook-Pro.local",
				"tags" : [ "beats_input_codec_plain_applied" ],
				"app" : "alice"
			      },
			      "sort" : [ 112 ]
			    } ]
			  }
			}`

		w.Write([]byte(resp))
	}

	addresses := make([]string, 1)
	addresses[0] = ts.URL

	client, err := elasticsearch.NewClient(elasticsearch.Config{Addresses: addresses})
	if err != nil {
		return nil, err
	}
	return client, nil
}
