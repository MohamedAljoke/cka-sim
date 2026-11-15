package usecases

import "context"

type SaveNote struct {
	platform Platform
}

func NewSaveNote(platform Platform) *SaveNote {
	return &SaveNote{platform: platform}
}

func (uc *SaveNote) Execute(ctx context.Context, student, taskID, text string) error {
	return uc.platform.SaveNote(ctx, student, taskID, text)
}
