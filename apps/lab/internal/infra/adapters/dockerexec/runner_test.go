package dockerexec

import (
	"strings"
	"testing"
)

func TestALongOutputKeepsItsEnd(t *testing.T) {
	out := &lastBytes{limit: 10}

	out.Write([]byte("0123456789"))
	out.Write([]byte("abc"))

	if got := string(out.buf); got != "3456789abc" {
		t.Errorf("kept %q", got)
	}
}

func TestAShortOutputIsKeptWhole(t *testing.T) {
	out := &lastBytes{limit: outputLimit}

	out.Write([]byte(strings.Repeat("x", 100)))

	if len(out.buf) != 100 {
		t.Errorf("kept %d bytes", len(out.buf))
	}
}
