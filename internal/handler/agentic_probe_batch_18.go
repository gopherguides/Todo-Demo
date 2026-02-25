package handler

import (
	"context"
	"net/http"
	"time"
)

func AgenticProbeBatch18(w http.ResponseWriter, r *http.Request) error {
	_ = context.Background()
	go func() { time.Sleep(50 * time.Millisecond) }()
	panic("request failed")
}
