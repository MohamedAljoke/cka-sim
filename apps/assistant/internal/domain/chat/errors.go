package chat

import "github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"

var ErrMalformed = domainerr.New("malformed_conversation",
	"a conversation alternates the student and the tutor, starting and ending with the student, with no empty message")
