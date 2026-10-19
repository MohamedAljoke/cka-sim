package labs

import (
	"errors"
	"strings"
	"testing"
)

func TestReadCommandsAreAllowed(t *testing.T) {
	for _, args := range [][]string{
		{"get", "pods", "-A"},
		{"get", "deploy", "web", "-n", "shop", "-o", "yaml"},
		{"get", "pods", "--namespace=shop", "-l", "app=web,tier!=db", "--show-labels"},
		{"get", "pods", "-l", "env in (prod, staging)"},
		{"describe", "pod/web-7d9f", "-n", "shop"},
		{"logs", "web-7d9f", "-c", "app", "--previous", "--tail=50"},
		{"events", "-n", "shop", "--for", "pod/web-7d9f"},
		{"top", "pods", "-A", "--containers"},
		{"explain", "pod.spec.containers", "--recursive"},
		{"api-resources", "--namespaced=true"},
		{"auth", "can-i", "list", "secrets", "--as", "system:serviceaccount:shop:web", "-n", "shop"},
		{"auth", "can-i", "*", "*"},
		{"get", "pods", "--sort-by", ".metadata.creationTimestamp"},
	} {
		if _, err := NewReadCommand(args); err != nil {
			t.Errorf("%q: %v", args, err)
		}
	}
}

func TestCommandsThatChangeOrEscapeAreRefused(t *testing.T) {
	for _, args := range [][]string{
		nil,
		{"delete", "pod", "web"},
		{"exec", "web", "--", "sh"},
		{"apply", "-f", "x.yaml"},
		{"auth", "reconcile"},
		{"get", "pods", "--kubeconfig=/root/other"},
		{"get", "pods", "--server", "https://evil"},
		{"get", "pods", "--token", "x"},
		{"get", "--raw", "/api"},
		{"get", "pods", "-w"},
		{"logs", "web", "-f"},
		{"get", "pods", "-o", "go-template={{.}}"},
		{"get", "pods", "-ojson"},
		{"get", "pods;", "rm", "-rf", "/"},
		{"get", "$(reboot)"},
		{"get", "pods", "-n", "--kubeconfig=/x"},
		{"get", "pods", "-n"},
		{"get", "pods", "--show-labels=maybe"},
		{"get", "pods", "-l", "a\nb"},
		{"get", strings.Repeat("a", 254)},
		{"get", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x"},
	} {
		if _, err := NewReadCommand(args); !errors.Is(err, ErrNotReadOnly) {
			t.Errorf("%q: got %v, want ErrNotReadOnly", args, err)
		}
	}
}

func TestARefusalNamesWhatWasRefused(t *testing.T) {
	_, err := NewReadCommand([]string{"get", "pods", "--kubeconfig=/x"})

	if err == nil || !strings.Contains(err.Error(), `"--kubeconfig=/x"`) {
		t.Errorf("got %v", err)
	}
}

func TestTheScriptQuotesEveryArgumentUnderATimeout(t *testing.T) {
	command, err := NewReadCommand([]string{"get", "pods", "-l", "app in (web)", "--sort-by", ".metadata.name"})
	if err != nil {
		t.Fatal(err)
	}

	want := "timeout 15 kubectl 'get' 'pods' '-l' 'app in (web)' '--sort-by' '.metadata.name'\n"
	if got := command.Script(); got != want {
		t.Errorf("Script() = %q, want %q", got, want)
	}
}

func TestAQuoteInAValueStaysInsideItsArgument(t *testing.T) {
	command, err := NewReadCommand([]string{"get", "pods", "-l", "a='; reboot; '"})
	if err != nil {
		t.Fatal(err)
	}

	want := `timeout 15 kubectl 'get' 'pods' '-l' 'a='\''; reboot; '\'''` + "\n"
	if got := command.Script(); got != want {
		t.Errorf("Script() = %q, want %q", got, want)
	}
}
