package usecases

type UseCases struct {
	Reply *Reply
}

// New takes a nil model when no AI is set up.
func New(model Model, tools Toolbox) UseCases {
	return UseCases{Reply: NewReply(model, tools)}
}
