package adapters

type WriteAdapter interface {
	ProcessMessages(msgs []string) []string
	ShouldProcess(msgs []string) bool
}