package tasks

type SetupState string

const (
	SetupPreparing SetupState = "preparing"
	SetupReady     SetupState = "ready"
	SetupFailed    SetupState = "failed"
)

type Current struct {
	Task  Task
	Setup SetupState
}
