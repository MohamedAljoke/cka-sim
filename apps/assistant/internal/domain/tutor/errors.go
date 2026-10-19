package tutor

import "github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"

var (
	ErrRefused     = domainerr.New("refused", "cka-sim said no")
	ErrUnavailable = domainerr.New("platform_unavailable", "cka-sim doesn't answer; is it running? (make dev)")
)
