package usecases

type UseCases struct {
	Open *OpenClaude
}

func New(cli CLI, endpoints Endpoints) UseCases {
	return UseCases{Open: NewOpenClaude(cli, endpoints)}
}
