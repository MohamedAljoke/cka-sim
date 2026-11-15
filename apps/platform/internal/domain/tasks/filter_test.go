package tasks

import (
	"errors"
	"strings"
	"testing"
)

func TestFilterMatches(t *testing.T) {
	netpol := Task{ID: "nw-netpol", Title: "Lock down egress", Domain: Networking, Topics: []string{"networkpolicy", "egress"}}
	pvc := Task{ID: "st-pvc", Domain: Storage, Topics: []string{"pvc"}}
	bare := Task{ID: "nw-bare", Domain: Networking}

	tests := []struct {
		name   string
		filter Filter
		task   Task
		want   bool
	}{
		{"an empty filter keeps everything", Filter{}, bare, true},
		{"one of the domains", Filter{Domains: []Domain{Storage, Networking}}, netpol, true},
		{"another domain", Filter{Domains: []Domain{Networking}}, pvc, false},
		{"one of the topics", Filter{Topics: []string{"pvc", "egress"}}, netpol, true},
		{"none of the topics", Filter{Topics: []string{"pvc"}}, netpol, false},
		{"a topic filter drops a task without topics", Filter{Topics: []string{"networkpolicy"}}, bare, false},
		{"domain and topic both match", Filter{Domains: []Domain{Networking}, Topics: []string{"networkpolicy"}}, netpol, true},
		{"the domain matches and the topic does not", Filter{Domains: []Domain{Networking}, Topics: []string{"pvc"}}, netpol, false},
		{"a search finds a word of the title", Filter{Search: "egress"}, netpol, true},
		{"a search finds part of the id", Filter{Search: "netpol"}, netpol, true},
		{"a search ignores case", Filter{Search: "LOCK"}, netpol, true},
		{"every word of a search must be found", Filter{Search: "lock ingress"}, netpol, false},
		{"words of a search may come from the id and the title", Filter{Search: "nw- down"}, netpol, true},
		{"a blank search keeps everything", Filter{Search: "  "}, bare, true},
		{"a search drops a task it does not find", Filter{Search: "egress"}, pvc, false},
		{"search and domain both match", Filter{Domains: []Domain{Networking}, Search: "lock"}, netpol, true},
		{"the search matches and the domain does not", Filter{Domains: []Domain{Storage}, Search: "lock"}, netpol, false},
		{"one of the ids", Filter{IDs: []string{"st-pvc", "nw-netpol"}}, netpol, true},
		{"none of the ids", Filter{IDs: []string{"st-pvc"}}, netpol, false},
		{"the id matches and the domain does not", Filter{Domains: []Domain{Storage}, IDs: []string{"nw-netpol"}}, netpol, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.Matches(tt.task); got != tt.want {
				t.Errorf("Matches = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAFilterIsEmptyOnlyWhenItSelectsNothing(t *testing.T) {
	for _, f := range []Filter{{Domains: []Domain{Storage}}, {Topics: []string{"pvc"}}, {Search: "pvc"}, {IDs: []string{"st-pvc"}}} {
		if f.Empty() {
			t.Errorf("%+v is empty", f)
		}
	}
	if !(Filter{Search: "  "}).Empty() {
		t.Error("a blank search selects nothing")
	}
}

func TestSelectKeepsWhatTheFilterMatches(t *testing.T) {
	all := []Task{{ID: "nw-netpol", Domain: Networking}, {ID: "st-pvc", Domain: Storage}, {ID: "nw-dns", Domain: Networking}}

	got, err := Filter{Domains: []Domain{Networking}, IDs: []string{"nw-dns", "st-pvc"}}.Select(all)

	if err != nil || len(got) != 1 || got[0].ID != "nw-dns" {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestSelectRefusesAnIDThatNamesNoTask(t *testing.T) {
	_, err := Filter{IDs: []string{"st-pvc", "no-such-task"}}.Select([]Task{{ID: "st-pvc"}})

	if !errors.Is(err, ErrNotFound) || !strings.Contains(err.Error(), `"no-such-task"`) {
		t.Errorf("got %v", err)
	}
}

func TestAFilterOfKnownDomainsAndKebabTopicsIsValid(t *testing.T) {
	f := Filter{Domains: []Domain{Networking, Storage}, Topics: []string{"networkpolicy", "rolling-update"}}

	if err := f.Validate(); err != nil {
		t.Errorf("got %v, want no error", err)
	}
}

func TestAFilterRejectsAnUnknownDomain(t *testing.T) {
	err := Filter{Domains: []Domain{Networking, "security"}}.Validate()

	if !errors.Is(err, ErrUnknownDomain) {
		t.Errorf("got %v, want ErrUnknownDomain", err)
	}
}

func TestAFilterRejectsAMalformedTopic(t *testing.T) {
	for _, topic := range []string{"", "Network Policy", "net_pol"} {
		err := Filter{Topics: []string{topic}}.Validate()

		if !errors.Is(err, ErrMalformedTopic) {
			t.Errorf("%q: got %v, want ErrMalformedTopic", topic, err)
		}
	}
}
