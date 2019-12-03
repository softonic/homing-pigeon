package elasticsearch

type ElasticSearchBulkResponse struct {
	Errors bool          `json:"errors"`
	Items  []interface{} `json:"items"`
}

type ElasticsearchBody struct {
	Meta interface{} `json:"meta"`
	Data interface{} `json:"data"`
}
