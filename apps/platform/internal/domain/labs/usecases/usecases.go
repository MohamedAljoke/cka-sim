package usecases

type UseCases struct {
	Get   *GetLab
	Start *StartLab
	Stop  *StopLab
}

func New(provider LabProvider) UseCases {
	return UseCases{Get: NewGetLab(provider), Start: NewStartLab(provider), Stop: NewStopLab(provider)}
}
