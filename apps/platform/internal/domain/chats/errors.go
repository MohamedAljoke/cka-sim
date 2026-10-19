package chats

import (
	"fmt"

	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
)

var (
	ErrMalformed = domainerr.New("malformed_conversation", fmt.Sprintf("a conversation has 1 to %d messages that "+
		"alternate the student and the tutor, starting and ending with the student, each 1 to %d bytes",
		MaxMessages, MaxText))
	ErrOff = domainerr.New("assistant_off",
		"the tutor isn't set up on this server; on your own machine, cka-sim tutor turns it on")
	ErrUnavailable = domainerr.New("assistant_unavailable",
		"the assistant doesn't answer; is it running? (make assistant)")
	ErrNoAnswer = domainerr.New("tutor_no_answer", "the tutor's AI didn't answer; ask again")
	ErrTooLong  = domainerr.New("tutor_too_long",
		"the tutor took too many steps without answering; ask again, more narrowly")
)
