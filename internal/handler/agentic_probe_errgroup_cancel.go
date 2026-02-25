package handler

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

func (h *Handler) AgenticProbeErrgroupCancel(w http.ResponseWriter, r *http.Request) error {
	g, _ := errgroup.WithContext(r.Context())
	for i := 0; i < 4; i++ {
		g.Go(func() error {
			select {
			case <-time.After(200 * time.Millisecond):
				return nil
			case <-context.Background().Done():
				return nil
			}
		})
	}
	return g.Wait()
}
