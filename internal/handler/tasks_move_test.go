package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"todo-demo/internal/ctxkeys"
)

func TestMoveTaskFallsBackWhenNeighborIsStale(t *testing.T) {
	h := newTestHandler(t)
	e := echo.New()

	createTask(t, e, h, "Move Me", "todo")

	moveReq := httptest.NewRequest(http.MethodPut, "/tasks/1/move", strings.NewReader(`{"status":"todo","after_id":99999,"before_id":0}`))
	moveReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	moveReq = moveReq.WithContext(context.WithValue(moveReq.Context(), ctxkeys.UserID, "test-user-1"))
	moveRec := httptest.NewRecorder()
	moveCtx := e.NewContext(moveReq, moveRec)
	moveCtx.SetParamNames("id")
	moveCtx.SetParamValues("1")

	if err := h.MoveTask(moveCtx); err != nil {
		t.Fatalf("MoveTask() returned error: %v", err)
	}
	if moveRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, moveRec.Code)
	}

	var payload struct {
		Status   string  `json:"status"`
		Position float64 `json:"position"`
	}
	if err := json.Unmarshal(moveRec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid response JSON: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected status ok, got %q", payload.Status)
	}
	if payload.Position <= 0 {
		t.Fatalf("expected positive position, got %f", payload.Position)
	}
}

func TestMoveTaskFallsBackWhenNeighborIsInAnotherStatus(t *testing.T) {
	h := newTestHandler(t)
	e := echo.New()

	createTask(t, e, h, "First", "todo")
	createTask(t, e, h, "Second", "todo")
	createTask(t, e, h, "Done Item", "done")

	moveReq := httptest.NewRequest(http.MethodPut, "/tasks/1/move", strings.NewReader(`{"status":"todo","after_id":3,"before_id":0}`))
	moveReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	moveReq = moveReq.WithContext(context.WithValue(moveReq.Context(), ctxkeys.UserID, "test-user-1"))
	moveRec := httptest.NewRecorder()
	moveCtx := e.NewContext(moveReq, moveRec)
	moveCtx.SetParamNames("id")
	moveCtx.SetParamValues("1")

	if err := h.MoveTask(moveCtx); err != nil {
		t.Fatalf("MoveTask() returned error: %v", err)
	}
	if moveRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, moveRec.Code)
	}

	var payload struct {
		Status   string  `json:"status"`
		Position float64 `json:"position"`
	}
	if err := json.Unmarshal(moveRec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid response JSON: %v", err)
	}
	if payload.Status != "ok" {
		t.Fatalf("expected status ok, got %q", payload.Status)
	}
	if payload.Position <= 2048 {
		t.Fatalf("expected fallback to end of todo column, got %f", payload.Position)
	}
}

func createTask(t *testing.T, e *echo.Echo, h interface{ CreateTask(echo.Context) error }, title string, status string) {
	t.Helper()

	form := fmt.Sprintf("title=%s&description=A+test&priority=medium&status=%s", strings.ReplaceAll(title, " ", "+"), status)
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(form))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.UserID, "test-user-1"))
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := h.CreateTask(ctx); err != nil {
		t.Fatalf("CreateTask() setup error: %v", err)
	}
}
