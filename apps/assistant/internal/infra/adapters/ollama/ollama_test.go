package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNomicDocumentsAndQueriesGetTheirTaskPrefix(t *testing.T) {
	var got []embedRequest
	url := serve(t, func(w http.ResponseWriter, req embedRequest) {
		got = append(got, req)
		answer(w, len(req.Input))
	})
	embedder := New(url, "nomic-embed-text")

	if _, err := embedder.EmbedDocuments(context.Background(), []string{"a page"}); err != nil {
		t.Fatal(err)
	}
	if _, err := embedder.EmbedQuery(context.Background(), "a question"); err != nil {
		t.Fatal(err)
	}

	if got[0].Model != "nomic-embed-text" || got[0].Input[0] != "search_document: a page" ||
		got[1].Input[0] != "search_query: a question" {
		t.Errorf("sent %+v", got)
	}
}

func TestOtherModelsGetTheTextAsIs(t *testing.T) {
	var got embedRequest
	url := serve(t, func(w http.ResponseWriter, req embedRequest) {
		got = req
		answer(w, 1)
	})

	if _, err := New(url, "embeddinggemma").EmbedQuery(context.Background(), "a question"); err != nil {
		t.Fatal(err)
	}

	if got.Input[0] != "a question" {
		t.Errorf("sent %q", got.Input)
	}
}

func TestOllamasErrorIsKept(t *testing.T) {
	url := serve(t, func(w http.ResponseWriter, _ embedRequest) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"model \"nomic-embed-text\" not found, try pulling it first"}`))
	})

	_, err := New(url, "nomic-embed-text").EmbedQuery(context.Background(), "q")

	if err == nil || !strings.Contains(err.Error(), "try pulling it first") {
		t.Errorf("err %v", err)
	}
}

func TestTooFewEmbeddingsIsAnError(t *testing.T) {
	url := serve(t, func(w http.ResponseWriter, _ embedRequest) { answer(w, 1) })

	_, err := New(url, "nomic-embed-text").EmbedDocuments(context.Background(), []string{"a", "b"})

	if err == nil {
		t.Error("no error for 1 embedding of 2 texts")
	}
}

func serve(t *testing.T, handle func(http.ResponseWriter, embedRequest)) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req embedRequest
		if r.URL.Path != "/api/embed" || json.NewDecoder(r.Body).Decode(&req) != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		handle(w, req)
	}))
	t.Cleanup(srv.Close)

	return srv.URL + "/"
}

func answer(w http.ResponseWriter, n int) {
	vectors := make([][]float32, n)
	for i := range vectors {
		vectors[i] = []float32{0.5, 0.5}
	}
	_ = json.NewEncoder(w).Encode(embedResponse{Embeddings: vectors})
}
