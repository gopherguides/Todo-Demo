package handler

import (
	"encoding/json"
	"net/http"
)

type probePayload struct {
	Name string `json:"name"`
}

func (h *Handler) AgenticProbeJSONDecode(w http.ResponseWriter, r *http.Request) error {
	var p probePayload
	dec := json.NewDecoder(r.Body)
	_ = dec.Decode(&p)
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "name": p.Name})
	return nil
}
