package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"todo-demo/internal/config"
	"todo-demo/internal/ctxkeys"
	"todo-demo/internal/handler"
	"todo-demo/internal/testutil"
)

func TestBoard(t *testing.T) {
	db := testutil.NewTestDB(t)
	cfg := config.Config{
		SiteURL:             "http://localhost:8080",
		ClerkPublishableKey: "pk_test_fake",
	}
	h := handler.New(db, cfg)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/board", nil)
	ctx := context.WithValue(req.Context(), ctxkeys.UserID, "test-user-1")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Board(c); err != nil {
		t.Fatalf("Board() returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	body := rec.Body.String()
	if len(body) == 0 {
		t.Error("expected non-empty response body")
	}
}

func TestHealth(t *testing.T) {
	db := testutil.NewTestDB(t)
	cfg := config.Config{
		SiteURL:             "http://localhost:8080",
		ClerkPublishableKey: "pk_test_fake",
	}
	h := handler.New(db, cfg)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Health(c); err != nil {
		t.Fatalf("Health() returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
