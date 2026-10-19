package usecases

type UseCases struct {
	Reply *Reply
}

func New(tasks Tasks) UseCases {
	return UseCases{Reply: NewReply(tasks)}
}
