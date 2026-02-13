package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"todo-demo/internal/config"
	"todo-demo/internal/ctxkeys"
	"todo-demo/internal/database/sqlc"
	"todo-demo/internal/testutil"
)

func TestUpdateTaskStatusUsesDestinationMaxPosition(t *testing.T) {
	db := testutil.NewTestDB(t)
	cfg := config.Config{
		SiteURL:             "http://localhost:8080",
		ClerkPublishableKey: "pk_test_fake",
	}
	h := New(db, cfg)
	e := echo.New()

	createTaskForStatusTest(t, e, h, "Todo A", "todo")
	createTaskForStatusTest(t, e, h, "Todo B", "todo")
	createTaskForStatusTest(t, e, h, "Done A", "done")

	updateForm := "title=Todo+B&description=test&priority=medium&status=done&due_date="
	updateReq := httptest.NewRequest(http.MethodPut, "/tasks/2", strings.NewReader(updateForm))
	updateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	updateReq = updateReq.WithContext(context.WithValue(updateReq.Context(), ctxkeys.UserID, "test-user-1"))
	updateRec := httptest.NewRecorder()
	updateCtx := e.NewContext(updateReq, updateRec)
	updateCtx.SetParamNames("id")
	updateCtx.SetParamValues("2")

	if err := h.UpdateTask(updateCtx); err != nil {
		t.Fatalf("UpdateTask() returned error: %v", err)
	}
	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, updateRec.Code)
	}

	task, err := h.queries.GetTask(context.Background(), sqlc.GetTaskParams{
		ID:     2,
		UserID: "test-user-1",
	})
	if err != nil {
		t.Fatalf("GetTask() returned error: %v", err)
	}
	if task.Status != "done" {
		t.Fatalf("expected task status done, got %s", task.Status)
	}
	if task.Position != 2048 {
		t.Fatalf("expected destination-based position 2048, got %f", task.Position)
	}
}

func createTaskForStatusTest(t *testing.T, e *echo.Echo, h *Handler, title string, status string) {
	t.Helper()

	form := "title=" + strings.ReplaceAll(title, " ", "+") + "&description=test&priority=medium&status=" + status
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(form))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.UserID, "test-user-1"))
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := h.CreateTask(ctx); err != nil {
		t.Fatalf("CreateTask() setup error: %v", err)
	}
}
