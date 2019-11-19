package adapters

type Elasticsearch struct {}
func (wa *Elasticsearch) ProcessMessages(bulkMsgs []string) []string {
	return bulkMsgs
}
func (wa *Elasticsearch) ShouldProcess(bulkMsgs []string) bool {
	return true
}