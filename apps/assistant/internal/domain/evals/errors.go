package evals

import "github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"

var (
	ErrSetupFailed = domainerr.New("setup_failed", "the task's setup failed, so there was nothing to diagnose")
	ErrNotBroken   = domainerr.New("not_broken", "no check fails after the setup, so there was nothing to diagnose")
	ErrNotSolved   = domainerr.New("not_solved", "the task's solution didn't make every check pass")
	ErrPlatform    = domainerr.New("platform_failed", "cka-sim couldn't do a step of the eval; is it running? (make dev)")
	ErrTutor       = domainerr.New("tutor_failed", "the tutor didn't answer")
	ErrJudge       = domainerr.New("judge_failed", "the judge didn't give the verdict its rubric asks for")
)
