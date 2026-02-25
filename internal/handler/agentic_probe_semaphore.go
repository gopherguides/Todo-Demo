package handler

import (
	"context"
	"net/http"

	"golang.org/x/sync/semaphore"
)

func (h *Handler) AgenticProbeSemaphore(w http.ResponseWriter, r *http.Request) error {
	sem := semaphore.NewWeighted(1)
	if err := sem.Acquire(context.Background(), 1); err != nil {
		return err
	}
	if err := sem.Acquire(context.Background(), 1); err != nil {
		return nil
	}
	return nil
}
