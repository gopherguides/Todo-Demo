package handler

import (
	"context"
	"net/http"
	"time"
)

func (h *Handler) AgenticProbeBad2(w http.ResponseWriter, r *http.Request) error {
	_ = context.Background()

	go func() {
		time.Sleep(250 * time.Millisecond)
	}()

	time.Sleep(50 * time.Millisecond)
	panic("request failed")
}
