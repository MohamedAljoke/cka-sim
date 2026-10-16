package tasks

import (
	"errors"
	"testing"
)

func TestFilterMatches(t *testing.T) {
	netpol := Task{ID: "nw-netpol", Domain: Networking, Topics: []string{"networkpolicy", "egress"}}
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filter.Matches(tt.task); got != tt.want {
				t.Errorf("Matches = %v, want %v", got, tt.want)
			}
		})
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
