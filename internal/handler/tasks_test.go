package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"todo-demo/internal/config"
	"todo-demo/internal/ctxkeys"
	"todo-demo/internal/handler"
	"todo-demo/internal/testutil"
)

func newTestHandler(t *testing.T) *handler.Handler {
	t.Helper()
	db := testutil.NewTestDB(t)
	cfg := config.Config{
		SiteURL:             "http://localhost:8080",
		ClerkPublishableKey: "pk_test_fake",
	}
	return handler.New(db, cfg)
}

func TestCreateTask(t *testing.T) {
	h := newTestHandler(t)

	e := echo.New()
	form := "title=Test+Task&description=A+test&priority=high&status=todo"
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(form))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	ctx := context.WithValue(req.Context(), ctxkeys.UserID, "test-user-1")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.CreateTask(c); err != nil {
		t.Fatalf("CreateTask() returned error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	if rec.Header().Get("HX-Refresh") != "true" {
		t.Error("expected HX-Refresh header")
	}
}

func TestCreateTaskMissingTitle(t *testing.T) {
	h := newTestHandler(t)

	e := echo.New()
	form := "description=A+test&priority=high"
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(form))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	ctx := context.WithValue(req.Context(), ctxkeys.UserID, "test-user-1")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateTask(c)
	if err == nil {
		t.Fatal("expected error for missing title")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, httpErr.Code)
	}
}

func TestDeleteTask(t *testing.T) {
	h := newTestHandler(t)

	e := echo.New()
	form := "title=Delete+Me&status=todo"
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(form))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	ctx := context.WithValue(req.Context(), ctxkeys.UserID, "test-user-1")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.CreateTask(c); err != nil {
		t.Fatalf("CreateTask() setup error: %v", err)
	}

	req2 := httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
	ctx2 := context.WithValue(req2.Context(), ctxkeys.UserID, "test-user-1")
	req2 = req2.WithContext(ctx2)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	c2.SetParamNames("id")
	c2.SetParamValues("1")

	if err := h.DeleteTask(c2); err != nil {
		t.Fatalf("DeleteTask() returned error: %v", err)
	}
	if rec2.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec2.Code)
	}
}
