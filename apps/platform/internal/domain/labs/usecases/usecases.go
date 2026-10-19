package usecases

type UseCases struct {
	Get       *GetLab
	Start     *StartLab
	Stop      *StopLab
	OpenShell *OpenShell
	Inspect   *InspectCluster
}

func New(provider LabProvider, runner ScriptRunner) UseCases {
	return UseCases{
		Get:       NewGetLab(provider),
		Start:     NewStartLab(provider),
		Stop:      NewStopLab(provider),
		OpenShell: NewOpenShell(provider),
		Inspect:   NewInspectCluster(runner),
	}
}
