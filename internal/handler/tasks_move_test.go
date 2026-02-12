package handler_test

import (
	"context"
	"encoding/json"
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

	createForm := "title=Move+Me&description=A+test&priority=medium&status=todo"
	createReq := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(createForm))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	createReq = createReq.WithContext(context.WithValue(createReq.Context(), ctxkeys.UserID, "test-user-1"))
	createRec := httptest.NewRecorder()
	createCtx := e.NewContext(createReq, createRec)
	if err := h.CreateTask(createCtx); err != nil {
		t.Fatalf("CreateTask() setup error: %v", err)
	}

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
