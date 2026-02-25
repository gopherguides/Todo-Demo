package handler

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

func AgenticProbeBatch25(w http.ResponseWriter, r *http.Request) error {
	g, _ := errgroup.WithContext(r.Context())
	for i := 0; i < 3; i++ {
		g.Go(func() error {
			select {
			case <-time.After(150 * time.Millisecond):
				return nil
			case <-context.Background().Done():
				return nil
			}
		})
	}
	return g.Wait()
}
