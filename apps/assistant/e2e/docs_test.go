package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/domain/docs"
	"github.com/MohamedAljoke/cka-sim/apps/assistant/internal/infra/adapters/files"
)

func TestWithoutADocsIndexThereIsNoSearchDocs(t *testing.T) {
	res, err := connectTo(t, mcpURL).ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}

	if slices.ContainsFunc(res.Tools, func(tool *mcp.Tool) bool { return tool.Name == "search_docs" }) {
		t.Error("search_docs without an index")
	}
}

func TestSearchDocsAnswersFromTheIndexWithItsLinks(t *testing.T) {
	index := writeIndex(t, "fake-embed")
	assistant, url, err := start("http://127.0.0.1:1", assistantToken, "CKA_SIM_DOCS_INDEX="+index,
		"CKA_SIM_EMBED_MODEL=fake-embed", "CKA_SIM_OLLAMA_URL="+serveOllama(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { stop(assistant) })

	res, err := connectTo(t, url+"/mcp").CallTool(context.Background(), &mcp.CallToolParams{
		Name: "search_docs", Arguments: map[string]any{"query": "drain a node"},
	})

	if err != nil || res.IsError || !strings.Contains(text(res), "https://kubernetes.io/docs/tasks/drain/") {
		t.Errorf("got %v, %s", err, text(res))
	}
}

func TestAnIndexFromAnotherModelStopsTheAssistantAtStart(t *testing.T) {
	cmd := exec.Command(assistantBin)
	cmd.Env = []string{"CKA_SIM_ASSISTANT_ADDR=127.0.0.1:0", "CKA_SIM_PLATFORM_URL=http://127.0.0.1:1",
		"CKA_SIM_ASSISTANT_TOKEN=" + assistantToken, "CKA_SIM_DOCS_INDEX=" + writeIndex(t, "gemini-embedding-001"),
		"CKA_SIM_EMBED_MODEL=nomic-embed-text"}

	out, err := cmd.CombinedOutput()

	if err == nil || !strings.Contains(string(out), "embedded with gemini-embedding-001") {
		t.Errorf("got %v: %s", err, out)
	}
}

func writeIndex(t *testing.T, model string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "docs.gob.gz")
	index := docs.Index{Model: model, Pages: []docs.IndexedPage{{Path: "docs/tasks/drain.md", Hash: "h",
		Chunks: []docs.IndexedChunk{{
			Chunk:  docs.Chunk{Title: "Drain", URL: "https://kubernetes.io/docs/tasks/drain/", Text: "Drain a node."},
			Vector: []float32{1, 0},
		}},
	}}}
	if err := files.SaveDocs(path, index); err != nil {
		t.Fatal(err)
	}
	return path
}

func serveOllama(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"embeddings": [][]float32{{1, 0}}})
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}
