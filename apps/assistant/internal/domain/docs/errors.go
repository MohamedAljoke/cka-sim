package docs

import "github.com/MohamedAljoke/cka-sim/apps/go-shared/domainerr"

var (
	ErrQuery       = domainerr.New("docs_query", "the query must be 1 to 500 characters")
	ErrUnavailable = domainerr.New("docs_unavailable", "the docs search can't read the question; is Ollama running?")
)
