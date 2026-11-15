package usecases

type UseCases struct {
	Start  *StartExam
	Result *GetResult
}

func New(platform Platform) UseCases {
	return UseCases{Start: NewStartExam(platform), Result: NewGetResult(platform)}
}
