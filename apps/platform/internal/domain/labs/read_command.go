package labs

import (
	"fmt"
	"regexp"
	"strings"
)

// ControlPlane is where a read command runs: root's kubeconfig there reaches the whole cluster.
const ControlPlane = "cka-sim-control-plane"

const maxArgs = 24

var (
	readVerbs = map[string]bool{
		"get": true, "describe": true, "logs": true, "events": true, "top": true, "explain": true,
		"api-resources": true, "api-versions": true, "version": true,
	}
	flagsWithValue = map[string]bool{
		"-n": true, "--namespace": true, "-l": true, "--selector": true, "--field-selector": true,
		"-o": true, "--output": true, "-c": true, "--container": true, "--tail": true, "--since": true,
		"--sort-by": true, "--for": true, "--types": true, "--as": true, "--api-group": true,
	}
	switches = map[string]bool{
		"-A": true, "--all-namespaces": true, "--show-labels": true, "-p": true, "--previous": true,
		"--recursive": true, "--containers": true, "--no-headers": true, "--all-containers": true,
		"--list": true, "--show-kind": true, "--timestamps": true, "--namespaced": true,
	}
	outputs = map[string]bool{"wide": true, "yaml": true, "json": true, "name": true}

	plainArg  = regexp.MustCompile(`^[A-Za-z0-9*][A-Za-z0-9._:/@*-]{0,252}$`)
	flagValue = regexp.MustCompile(`^[^\x00-\x1f\x7f-][^\x00-\x1f\x7f]{0,252}$`)
)

// ReadCommand is a kubectl command that only reads the cluster: one of a few verbs, with flags that change what
// is shown but never where kubectl connects or what it runs.
type ReadCommand struct {
	args []string
}

func NewReadCommand(args []string) (ReadCommand, error) {
	if len(args) == 0 || len(args) > maxArgs {
		return ReadCommand{}, fmt.Errorf("%w: give a verb and at most %d arguments", ErrNotReadOnly, maxArgs-1)
	}

	rest, err := afterVerb(args)
	if err != nil {
		return ReadCommand{}, err
	}

	for i := 0; i < len(rest); i++ {
		if !strings.HasPrefix(rest[i], "-") {
			if !plainArg.MatchString(rest[i]) {
				return ReadCommand{}, refused(rest[i])
			}

			continue
		}

		used, err := checkFlag(rest[i:])
		if err != nil {
			return ReadCommand{}, err
		}

		i += used - 1
	}

	return ReadCommand{args: args}, nil
}

// Script runs the command with every argument quoted, so nothing in one reaches the shell.
func (c ReadCommand) Script() string {
	quoted := make([]string, len(c.args))
	for i, arg := range c.args {
		quoted[i] = "'" + strings.ReplaceAll(arg, "'", `'\''`) + "'"
	}

	return "timeout 15 kubectl " + strings.Join(quoted, " ") + "\n"
}

func afterVerb(args []string) ([]string, error) {
	switch {
	case readVerbs[args[0]]:
		return args[1:], nil
	case args[0] == "auth" && len(args) > 1 && args[1] == "can-i":
		return args[2:], nil
	}

	return nil, refused(args[0])
}

// checkFlag checks the flag that starts args and answers how many arguments it used: one, or two when its value
// is the next argument.
func checkFlag(args []string) (int, error) {
	flag, value, hasValue := strings.Cut(args[0], "=")

	switch {
	case switches[flag] && (!hasValue || value == "true" || value == "false"):
		return 1, nil
	case flagsWithValue[flag] && hasValue:
		return 1, checkValue(flag, value)
	case flagsWithValue[flag] && len(args) > 1:
		return 2, checkValue(flag, args[1])
	}

	return 0, refused(args[0])
}

func checkValue(flag, value string) error {
	switch {
	case !flagValue.MatchString(value):
		return refused(flag + "=" + value)
	case (flag == "-o" || flag == "--output") && !outputs[value]:
		return fmt.Errorf("%w: -o takes wide, yaml, json or name", ErrNotReadOnly)
	}

	return nil
}

func refused(arg string) error {
	return fmt.Errorf("%w: %q isn't allowed", ErrNotReadOnly, arg)
}
