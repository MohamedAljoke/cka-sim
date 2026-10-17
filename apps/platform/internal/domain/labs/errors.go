package labs

import "github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"

var (
	ErrBusy        = domainerr.New("lab_busy", "the lab is still starting or stopping; wait for it to finish")
	ErrUnavailable = domainerr.New("lab_unavailable", "the lab doesn't answer; is it running? (make lab)")
	ErrNotReady    = domainerr.New("lab_not_ready", "the lab isn't ready; start it and wait for it first")
)
