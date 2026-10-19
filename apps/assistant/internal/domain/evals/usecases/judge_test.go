package usecases

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat"
	chatusecases "github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/chat/usecases"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/evals"
)

func TestTheJudgeGetsNumberedFailingChecksTheSolutionAndTheAnswer(t *testing.T) {
	model := judgeAnswering(`{"named": [2, 2], "gaveAway": false, "quote": "", "why": "names port 30080"}`)

	got, err := NewJudge(model).Diagnosis(context.Background(), failing, "kubectl create service nodeport", "Check port 30080.")

	if err != nil || !slices.Equal(got.Named, []string{"Port 30080 answers on every node"}) || got.Why != "names port 30080" {
		t.Errorf("got %+v, %v", got, err)
	}
	req := model.seen[0]
	if req.System != diagnosisRubric || len(req.Tools) != 0 ||
		!strings.Contains(req.Turns[0].Text, "1. Service web is a NodePort\n2. Port 30080 answers on every node\n") ||
		!strings.Contains(req.Turns[0].Text, "kubectl create service nodeport") ||
		!strings.Contains(req.Turns[0].Text, "Check port 30080.") {
		t.Errorf("the judge got %+v", req)
	}
}

func TestAVerdictWithoutJSONOrNamingAnUnknownCheckIsAFailedJudge(t *testing.T) {
	for _, text := range []string{"It names the Service.", `{"named": [3]}`, `{"named": "1"}`} {
		_, err := NewJudge(judgeAnswering(text)).Diagnosis(context.Background(), failing, "", "answer")

		if !errors.Is(err, evals.ErrJudge) {
			t.Errorf("%s: got %v", text, err)
		}
	}
}

func TestTheRecheckReadsJSONAroundOtherText(t *testing.T) {
	model := judgeAnswering("Here you go:\n```json\n{\"nothingWrong\": false, \"why\": \"says to fix the port\"}\n```")

	got, err := NewJudge(model).Recheck(context.Background(), "Fix the port.")

	if err != nil || got.NothingWrong || got.Why != "says to fix the port" || model.seen[0].System != recheckRubric {
		t.Errorf("got %+v, %v", got, err)
	}
}

var failing = []evals.Check{
	{Points: 2, Description: "Service web is a NodePort"},
	{Points: 1, Description: "Port 30080 answers on every node"},
}

func judgeAnswering(text string) *scriptedModel {
	return &scriptedModel{answers: []chatusecases.Turn{{Role: chat.Tutor, Text: text}}}
}
