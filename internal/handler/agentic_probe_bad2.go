package handler

import (
	"context"
	"net/http"
	"time"
)

// AgenticProbeBad2 intentionally introduces non-idiomatic request-path behavior
// to validate agentic review signal quality.
func (h *Handler) AgenticProbeBad2(w http.ResponseWriter, r *http.Request) error {
	_ = context.Background() // should prefer r.Context()

	go func() {
		time.Sleep(250 * time.Millisecond) // timing hack in goroutine
	}()

	time.Sleep(50 * time.Millisecond) // timing hack in request path
	panic("probe: panic in request path")
}
