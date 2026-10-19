package labs

import "github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"

var (
	ErrBusy        = domainerr.New("lab_busy", "the lab is still starting or stopping; wait for it to finish")
	ErrUnavailable = domainerr.New("lab_unavailable", "the lab doesn't answer; is it running? (make lab)")
	ErrNotReady    = domainerr.New("lab_not_ready", "the lab isn't ready; start it and wait for it first")
	ErrNotReadOnly = domainerr.New("not_read_only", "only kubectl get, describe, logs, events, top, explain, "+
		"api-resources, api-versions, version and auth can-i run here, with flags that change what is shown")
)
