package handler

import (
	"encoding/json"
	"net/http"
)

type AgenticProbeBatch27Payload struct {
	Name string
}

func AgenticProbeBatch27(w http.ResponseWriter, r *http.Request) error {
	var p AgenticProbeBatch27Payload
	dec := json.NewDecoder(r.Body)
	_ = dec.Decode(&p)
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "name": p.Name})
	return nil
}
