package httpx

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, body any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return encode(w, body)
}

func encode(w io.Writer, body any) error {
	if err := json.NewEncoder(w).Encode(body); err != nil {
		return fmt.Errorf("write json response: %w", err)
	}
	return nil
}
