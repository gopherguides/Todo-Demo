package handler_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"

	"todo-demo/internal/config"
	"todo-demo/internal/ctxkeys"
	"todo-demo/internal/handler"
)

func TestUpdateTaskRollsBackWhenPositionUpdateFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error: %v", err)
	}
	defer func() { _ = db.Close() }()

	h := handler.New(db, config.Config{
		SiteURL:             "http://localhost:8080",
		ClerkPublishableKey: "pk_test_fake",
	})

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`-- name: GetTask :one
SELECT id, user_id, title, description, priority, status, position, due_date, created_at, updated_at FROM tasks
WHERE id = ? AND user_id = ?
`)).
		WithArgs(int64(1), "test-user-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "title", "description", "priority", "status", "position", "due_date", "created_at", "updated_at",
		}).AddRow(
			int64(1), "test-user-1", "Old title", "old", "medium", "todo", float64(99999), nil, "created", "updated",
		))

	mock.ExpectQuery(regexp.QuoteMeta(`-- name: GetMaxPosition :one
SELECT COALESCE(MAX(position), 0) AS max_position
FROM tasks
WHERE user_id = ? AND status = ?
`)).
		WithArgs("test-user-1", "done").
		WillReturnRows(sqlmock.NewRows([]string{"max_position"}).AddRow(float64(1024)))

	mock.ExpectQuery(regexp.QuoteMeta(`-- name: UpdateTask :one
UPDATE tasks
SET title = ?, description = ?, priority = ?, status = ?, due_date = ?, updated_at = datetime('now')
WHERE id = ? AND user_id = ?
RETURNING id, user_id, title, description, priority, status, position, due_date, created_at, updated_at
`)).
		WithArgs("Updated", "updated", "high", "done", sql.NullString{}, int64(1), "test-user-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "title", "description", "priority", "status", "position", "due_date", "created_at", "updated_at",
		}).AddRow(
			int64(1), "test-user-1", "Updated", "updated", "high", "done", float64(99999), nil, "created", "updated",
		))

	mock.ExpectExec(regexp.QuoteMeta(`-- name: UpdateTaskPosition :execresult
UPDATE tasks
SET status = ?, position = ?, updated_at = datetime('now')
WHERE id = ? AND user_id = ?
`)).
		WithArgs("done", float64(2048), int64(1), "test-user-1").
		WillReturnError(errors.New("position update failed"))

	mock.ExpectRollback()

	e := echo.New()
	form := "title=Updated&description=updated&priority=high&status=done"
	req := httptest.NewRequest(http.MethodPut, "/tasks/1", strings.NewReader(form))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req = req.WithContext(context.WithValue(req.Context(), ctxkeys.UserID, "test-user-1"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	err = h.UpdateTask(c)
	if err == nil {
		t.Fatal("expected UpdateTask() error")
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, httpErr.Code)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
