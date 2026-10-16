package chat

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/llm"
	"github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"
)

const etcdPage = "https://kubernetes.io/docs/tasks/administer-cluster/configure-upgrade-etcd/"

func TestHintAsksForTheSchemaAndReturnsTheParsedHint(t *testing.T) {
	model := &fakeModel{replies: []llm.Message{
		{Role: llm.RoleModel, Calls: []llm.Call{{Name: "search_docs", Args: json.RawMessage(`{"query":"etcd snapshot"}`)}}},
		{Role: llm.RoleModel, Text: `{"diagnosis":"No snapshot was saved.","next_step":"Save one with etcdctl.","doc_url":"` + etcdPage + `#restoring"}`},
	}}
	var kinds []EventKind

	turn, err := newChat(model, 10).Hint(context.Background(), "a", etcdDocs(), nil, func(e Event) { kinds = append(kinds, e.Kind) })
	if err != nil {
		t.Fatal(err)
	}

	want := Hint{Diagnosis: "No snapshot was saved.", NextStep: "Save one with etcdctl.", DocURL: etcdPage + "#restoring"}
	if turn.Hint == nil || *turn.Hint != want {
		t.Errorf("hint = %+v", turn.Hint)
	}
	if turn.Text != "**Diagnosis:** No snapshot was saved.\n\n**Next step:** Save one with etcdctl.\n\n"+etcdPage+"#restoring" {
		t.Errorf("text = %q", turn.Text)
	}
	req := model.requests[0]
	if string(req.Schema) != hintSchema || req.Thinking != llm.ThinkingLow || req.Temperature == nil || req.System != "the reviewer rules"+hintRules {
		t.Errorf("request = schema %s, thinking %q, temperature %v, system %q", req.Schema, req.Thinking, req.Temperature, req.System)
	}
	if slices.Contains(kinds, EventText) || !slices.Contains(kinds, EventTool) || !slices.Contains(kinds, EventSources) {
		t.Errorf("events = %v; want tools and sources but no JSON text", kinds)
	}
}

func TestHintSendsABadAnswerBackOnceWithItsProblems(t *testing.T) {
	model := &fakeModel{replies: []llm.Message{
		{Role: llm.RoleModel, Text: `{"diagnosis":"Wrong image.","next_step":"Fix the tag.","doc_url":"https://kubernetes.io/docs/concepts/containers/images/"}`},
		{Role: llm.RoleModel, Text: `{"diagnosis":"Wrong image.","next_step":"Fix the tag.","doc_url":""}`},
	}}

	turn, err := newChat(model, 10).Hint(context.Background(), "a", &fakeTools{}, nil, ignore)
	if err != nil {
		t.Fatal(err)
	}

	if turn.Hint == nil || turn.Hint.DocURL != "" || len(model.requests) != 2 {
		t.Fatalf("hint = %+v after %d requests", turn.Hint, len(model.requests))
	}
	repair := model.requests[1].Messages
	sent, asked := repair[len(repair)-2], repair[len(repair)-1]
	if sent.Role != llm.RoleModel || !strings.Contains(sent.Text, "images/") {
		t.Errorf("the bad answer isn't in the history: %+v", sent)
	}
	if asked.Role != llm.RoleUser || !strings.Contains(asked.Text, "isn't a page search_docs returned") {
		t.Errorf("repair message = %q", asked.Text)
	}
}

func TestHintGivesUpWhenTheRepairFailsToo(t *testing.T) {
	model := &fakeModel{replies: []llm.Message{
		{Role: llm.RoleModel, Text: "Your image tag has a typo."},
		{Role: llm.RoleModel, Text: `{"diagnosis":"Typo.","next_step":"Fix it.","doc_url":"","confidence":"high"}`},
	}}

	_, err := newChat(model, 10).Hint(context.Background(), "a", &fakeTools{}, nil, ignore)

	if !errors.Is(err, domainerr.Unavailable("assistant_hint", "")) || len(model.requests) != 2 {
		t.Errorf("err = %v after %d requests", err, len(model.requests))
	}
}

func TestHintRepairSharesTheStepBudget(t *testing.T) {
	loop := llm.Message{Role: llm.RoleModel, Calls: []llm.Call{{Name: "current_task"}}}
	model := &fakeModel{replies: []llm.Message{loop, loop, {Role: llm.RoleModel, Text: "not json"}, loop, loop}}

	_, err := newChat(model, 10).Hint(context.Background(), "a", &fakeTools{}, nil, ignore)

	if !errors.Is(err, domainerr.Unavailable("assistant_steps", "")) || len(model.requests) != 3 {
		t.Errorf("err = %v after %d requests", err, len(model.requests))
	}
}

func TestHintCountsAgainstTheDailyLimit(t *testing.T) {
	chat := newChat(&fakeModel{replies: []llm.Message{{Role: llm.RoleModel, Text: "ok"}}}, 1)

	_, _ = chat.Reply(context.Background(), "a", &fakeTools{}, nil, ignore)
	_, err := chat.Hint(context.Background(), "a", &fakeTools{}, nil, ignore)

	if !errors.Is(err, domainerr.TooMany("assistant_limit", "")) {
		t.Errorf("err = %v, want the daily limit", err)
	}
}

func TestParseHintFindsEachProblem(t *testing.T) {
	searched := []Source{{Title: "Operating etcd clusters", URL: etcdPage + "#built-in-snapshot"}}
	for answer, want := range map[string]string{
		`{"diagnosis":"d","next_step":"n","doc_url":"` + etcdPage + `"}`: "",
		`{"diagnosis":"d","next_step":"n","doc_url":""}`:                 "",
		`the tag is wrong`: "isn't JSON matching the schema",
		`{"diagnosis":"d","next_step":"n","doc_url":"","why":"x"}`:                                         "isn't JSON matching the schema",
		`{"diagnosis":" ","next_step":"n","doc_url":""}`:                                                   "diagnosis is empty",
		`{"diagnosis":"d","next_step":"` + strings.Repeat("x", hintFieldMax+1) + `","doc_url":""}`:         "next_step is longer than",
		`{"diagnosis":"d","next_step":"Apply:\n` + "```" + `yaml\nkind: Pod\n` + "```" + `","doc_url":""}`: "next_step has a code block",
		`{"diagnosis":"d","next_step":"n","doc_url":"https://kubernetes.io/docs/home/"}`:                   "isn't a page search_docs returned",
	} {
		_, problems := parseHint(answer, searched)

		if got := strings.Join(problems, "; "); want == "" && got != "" || !strings.Contains(got, want) {
			t.Errorf("parseHint(%.60q) problems = %q, want %q", answer, got, want)
		}
	}
}

func etcdDocs() *fakeTools {
	return &fakeTools{out: map[string]string{"search_docs": `{"results":[` +
		`{"title":"Operating etcd clusters","url":"` + etcdPage + `#built-in-snapshot","text":"snapshot save"}]}`}}
}
