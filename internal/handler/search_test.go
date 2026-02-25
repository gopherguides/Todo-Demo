package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"todo-demo/internal/ctxkeys"
)

func seedTask(t *testing.T, h interface{ CreateTask(echo.Context) error }, e *echo.Echo, title, desc, priority, status string) {
	t.Helper()
	form := "title=" + title + "&description=" + desc + "&priority=" + priority + "&status=" + status
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(form))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.UserID, "test-user-1"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.CreateTask(c); err != nil {
		t.Fatalf("seed task %q: %v", title, err)
	}
}

func TestSearchTasksMatchesTitle(t *testing.T) {
	h := newTestHandler(t)
	e := echo.New()

	seedTask(t, h, e, "Buy+groceries", "milk+and+eggs", "medium", "todo")
	seedTask(t, h, e, "Fix+login+bug", "auth+broken", "high", "in_progress")
	seedTask(t, h, e, "Deploy+v2", "release+notes", "urgent", "todo")

	req := httptest.NewRequest(http.MethodGet, "/tasks/search?q=login", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.UserID, "test-user-1"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.SearchTasks(c); err != nil {
		t.Fatalf("SearchTasks() error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "login") {
		t.Error("expected search results to contain matching task")
	}
	if strings.Contains(body, "groceries") {
		t.Error("expected search results NOT to contain non-matching task")
	}
}

func TestSearchTasksMatchesDescription(t *testing.T) {
	h := newTestHandler(t)
	e := echo.New()

	seedTask(t, h, e, "Server+task", "upgrade+database+cluster", "medium", "todo")
	seedTask(t, h, e, "Client+task", "fix+button+style", "low", "todo")

	req := httptest.NewRequest(http.MethodGet, "/tasks/search?q=database", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.UserID, "test-user-1"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.SearchTasks(c); err != nil {
		t.Fatalf("SearchTasks() error: %v", err)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Server") {
		t.Error("expected search to match on description content")
	}
	if strings.Contains(body, "Client") {
		t.Error("expected non-matching task excluded from results")
	}
}

func TestSearchTasksEmptyQuery(t *testing.T) {
	h := newTestHandler(t)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/tasks/search?q=", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.UserID, "test-user-1"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.SearchTasks(c); err != nil {
		t.Fatalf("SearchTasks() error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Error("expected empty body for empty query")
	}
}

func TestSearchTasksNoResults(t *testing.T) {
	h := newTestHandler(t)
	e := echo.New()

	seedTask(t, h, e, "Real+task", "something+here", "medium", "todo")

	req := httptest.NewRequest(http.MethodGet, "/tasks/search?q=nonexistent", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.UserID, "test-user-1"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.SearchTasks(c); err != nil {
		t.Fatalf("SearchTasks() error: %v", err)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "No tasks matching") {
		t.Error("expected 'no tasks matching' message for zero results")
	}
}

func TestSearchTasksIsolatedPerUser(t *testing.T) {
	h := newTestHandler(t)
	e := echo.New()

	seedTask(t, h, e, "Shared+keyword", "test+data", "medium", "todo")

	req := httptest.NewRequest(http.MethodGet, "/tasks/search?q=keyword", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.UserID, "other-user"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.SearchTasks(c); err != nil {
		t.Fatalf("SearchTasks() error: %v", err)
	}
	body := rec.Body.String()
	if strings.Contains(body, "Shared") {
		t.Error("expected tasks to be isolated per user")
	}
}
