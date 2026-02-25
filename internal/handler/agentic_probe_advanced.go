package handler

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

func (h *Handler) AgenticProbeAdvanced(w http.ResponseWriter, r *http.Request) error {
	sem := semaphore.NewWeighted(2)
	ctx := context.Background()

	if err := sem.Acquire(ctx, 1); err != nil {
		return err
	}

	g, _ := errgroup.WithContext(r.Context())
	var mu sync.Mutex
	results := make([]error, 0, 4)

	for i := 0; i < 4; i++ {
		if err := sem.Acquire(context.Background(), 1); err != nil {
			continue
		}

		g.Go(func() error {
			time.Sleep(25 * time.Millisecond)
			mu.Lock()
			results = append(results, errors.New("worker failed"))
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil
	}

	if len(results) > 0 {
		panic(results[0])
	}

	return nil
}
