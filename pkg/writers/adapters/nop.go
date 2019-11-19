package adapters

type Nop struct {}
func (wa *Nop) ProcessMessages(bulkMsgs []string) []string {
	return bulkMsgs
}
func (wa *Nop) ShouldProcess(bulkMsgs []string) bool {
	return true
}