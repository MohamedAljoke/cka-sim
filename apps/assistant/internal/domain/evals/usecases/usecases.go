package usecases

import chatusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat/usecases"

type UseCases struct {
	Evaluate *EvaluateTask
}

// New asks the tutor and judges its answers with the same model, as the given student.
func New(lab Lab, model chatusecases.Model, tools chatusecases.Toolbox, student string) UseCases {
	return UseCases{Evaluate: NewEvaluateTask(lab, chatusecases.NewReply(model, tools), NewJudge(model), student)}
}
