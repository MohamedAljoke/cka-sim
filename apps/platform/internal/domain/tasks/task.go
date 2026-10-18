package tasks

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Domain string

const (
	Troubleshooting Domain = "troubleshooting"
	Architecture    Domain = "architecture"
	Networking      Domain = "networking"
	Workloads       Domain = "workloads"
	Storage         Domain = "storage"
)

// The CKA curriculum's share of the exam score, in percent.
var Weights = map[Domain]int{
	Troubleshooting: 30,
	Architecture:    25,
	Networking:      20,
	Workloads:       15,
	Storage:         10,
}

var Titles = map[Domain]string{
	Troubleshooting: "Troubleshooting",
	Architecture:    "Cluster Architecture, Installation and Configuration",
	Networking:      "Services and Networking",
	Workloads:       "Workloads and Scheduling",
	Storage:         "Storage",
}

func (d Domain) Known() bool {
	_, ok := Weights[d]
	return ok
}

type Order string

const Last Order = "last"

type Task struct {
	ID       string
	Title    string
	Domain   Domain
	Topics   []string
	Weight   int
	Host     string
	Order    Order
	Question string
	Solution string
	// Setup, Check, Solve and Reset are the bash the lab runs as root on Host, ready to run as they are. Check
	// prints the PASS or FAIL lines a grade is read from. Solve does what Solution explains. Reset, for a task
	// that breaks the cluster, undoes what its setup broke and does nothing on a healthy cluster.
	Setup string
	Check string
	Solve string
	Reset string
}

var kebab = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func (t Task) Validate() error {
	switch {
	case t.ID == "":
		return errors.New("id is required")
	case !kebab.MatchString(t.ID):
		return fmt.Errorf("id %q must be lowercase-kebab, like wl-scale", t.ID)
	case t.Title == "":
		return errors.New("title is required")
	case t.Weight <= 0:
		return errors.New("weight must be more than 0")
	case t.Host == "":
		return errors.New("host is required")
	case strings.TrimSpace(t.Question) == "":
		return errors.New("question is required")
	case strings.TrimSpace(t.Solution) == "":
		return errors.New("solution is required")
	case strings.TrimSpace(t.Setup) == "":
		return errors.New("setup is required")
	case strings.TrimSpace(t.Check) == "":
		return errors.New("check is required")
	case strings.TrimSpace(t.Solve) == "":
		return errors.New("solve is required")
	case t.Order != "" && t.Order != Last:
		return errors.New(`order must be empty or "last"`)
	}

	if !t.Domain.Known() {
		return fmt.Errorf("unknown domain %q", t.Domain)
	}

	for _, topic := range t.Topics {
		if !kebab.MatchString(topic) {
			return fmt.Errorf("topic %q must be lowercase-kebab, like rolling-update", topic)
		}
	}

	return nil
}
