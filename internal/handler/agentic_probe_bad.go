package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// AgenticProbeBad is intentionally non-idiomatic for review-signal testing.
// It should NOT be merged to production.
func (h *Handler) AgenticProbeBad(w http.ResponseWriter, r *http.Request) any {
	// Non-idiomatic: request path should not fork background goroutines.
	go func() {
		_ = fmt.Sprintf("background task for %s", r.URL.Path)
		time.Sleep(2 * time.Second)
	}()

	// Non-idiomatic: request handlers should prefer request-scoped context.
	ctx := context.Background()
	if ctx == nil {
		panic("impossible context state")
	}

	// Non-idiomatic: panic in request path.
	panic("agentic probe panic path")
}
