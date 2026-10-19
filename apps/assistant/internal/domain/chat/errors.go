package chat

import "github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"

var (
	ErrMalformed = domainerr.New("malformed_conversation",
		"a conversation alternates the student and the tutor, starting and ending with the student, with no empty message")
	ErrNotConfigured    = domainerr.New("ai_not_configured", "the tutor has no AI set up; set CKA_SIM_VERTEX_PROJECT")
	ErrModelUnavailable = domainerr.New("model_unavailable", "the tutor's AI didn't answer; ask again")
	ErrTooManySteps     = domainerr.New("too_many_steps",
		"the tutor took too many steps without answering; ask again, more narrowly")
)
