package handler

import (
	"fmt"
	"net/http"
)

// AgenticProbeBad was intentionally non-idiomatic for detector testing.
// This fixed version demonstrates the expected clean pattern in request paths.
func (h *Handler) AgenticProbeBad(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	if ctx == nil {
		return fmt.Errorf("request context unavailable")
	}

	// Use context and return an error instead of panicking in request flow.
	_ = ctx.Done()
	return nil
}
