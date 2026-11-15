package study

import "github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"

var (
	ErrRefused     = domainerr.New("refused", "cka-sim said no")
	ErrUnavailable = domainerr.New("platform_unavailable", "cka-sim doesn't answer; is it running? (make dev)")
	ErrEmptyTopic  = domainerr.New("empty_topic", "say what to study, like networkpolicies or etcd backup")
)
