package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/labstack/echo/v4"

	"todo-demo/internal/database/sqlc"
	mw "todo-demo/internal/middleware"
	"todo-demo/templates/components"
)

var (
	validStatuses   = []string{"todo", "in_progress", "done"}
	validPriorities = []string{"low", "medium", "high", "urgent"}
)

func (h *Handler) CreateTask(c echo.Context) error {
	userID := mw.GetUserID(c)

	title := c.FormValue("title")
	if title == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title is required")
	}

	description := c.FormValue("description")
	priority := c.FormValue("priority")
	if priority == "" {
		priority = "medium"
	}
	if !slices.Contains(validPriorities, priority) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid priority")
	}
	status := c.FormValue("status")
	if status == "" {
		status = "todo"
	}
	if !slices.Contains(validStatuses, status) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid status")
	}
	dueDate := c.FormValue("due_date")

	maxPos, err := h.queries.GetMaxPosition(c.Request().Context(), sqlc.GetMaxPositionParams{
		UserID: userID,
		Status: status,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get position")
	}

	var position float64
	switch v := maxPos.(type) {
	case float64:
		position = v + 1024
	case int64:
		position = float64(v) + 1024
	default:
		position = 1024
	}

	var dueDateNull sql.NullString
	if dueDate != "" {
		dueDateNull = sql.NullString{String: dueDate, Valid: true}
	}

	task, err := h.queries.CreateTask(c.Request().Context(), sqlc.CreateTaskParams{
		UserID:      userID,
		Title:       title,
		Description: description,
		Priority:    priority,
		Status:      status,
		Position:    position,
		DueDate:     dueDateNull,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create task")
	}

	_ = task
	c.Response().Header().Set("HX-Refresh", "true")
	return c.NoContent(http.StatusCreated)
}

func (h *Handler) EditTask(c echo.Context) error {
	userID := mw.GetUserID(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task id")
	}

	task, err := h.queries.GetTask(c.Request().Context(), sqlc.GetTaskParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "task not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get task")
	}

	return Render(c, http.StatusOK, components.EditForm(task))
}

func (h *Handler) UpdateTask(c echo.Context) error {
	userID := mw.GetUserID(c)
	ctx := c.Request().Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task id")
	}

	title := c.FormValue("title")
	if title == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title is required")
	}

	priority := c.FormValue("priority")
	if !slices.Contains(validPriorities, priority) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid priority")
	}
	status := c.FormValue("status")
	if !slices.Contains(validStatuses, status) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid status")
	}

	dueDate := c.FormValue("due_date")
	var dueDateNull sql.NullString
	if dueDate != "" {
		dueDateNull = sql.NullString{String: dueDate, Valid: true}
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to start transaction")
	}
	txQueries := h.queries.WithTx(tx)

	oldTask, err := txQueries.GetTask(ctx, sqlc.GetTaskParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "task not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get task")
	}

	statusChanged := oldTask.Status != status
	var nextStatusPosition float64
	if statusChanged {
		maxPos, err := txQueries.GetMaxPosition(ctx, sqlc.GetMaxPositionParams{
			UserID: userID,
			Status: status,
		})
		if err != nil {
			_ = tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get position for new status")
		}
		switch v := maxPos.(type) {
		case float64:
			nextStatusPosition = v + 1024
		case int64:
			nextStatusPosition = float64(v) + 1024
		default:
			nextStatusPosition = 1024
		}
	}

	task, err := txQueries.UpdateTask(ctx, sqlc.UpdateTaskParams{
		Title:       title,
		Description: c.FormValue("description"),
		Priority:    priority,
		Status:      status,
		DueDate:     dueDateNull,
		ID:          id,
		UserID:      userID,
	})
	if err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "task not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update task")
	}

	if statusChanged {
		_, err = txQueries.UpdateTaskPosition(ctx, sqlc.UpdateTaskPositionParams{
			Status:   status,
			Position: nextStatusPosition,
			ID:       id,
			UserID:   userID,
		})
		if err != nil {
			_ = tx.Rollback()
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to update task position")
		}
	}

	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to commit task update")
	}

	if statusChanged {
		c.Response().Header().Set("HX-Refresh", "true")
		return c.NoContent(http.StatusOK)
	}

	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Task updated", "type": "success"}}`)
	return Render(c, http.StatusOK, components.TaskCard(task))
}

func (h *Handler) DuplicateTask(c echo.Context) error {
	userID := mw.GetUserID(c)
	ctx := c.Request().Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task id")
	}

	task, err := h.queries.GetTask(ctx, sqlc.GetTaskParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "task not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load task")
	}

	maxPos, err := h.queries.GetMaxPosition(ctx, sqlc.GetMaxPositionParams{UserID: userID, Status: task.Status})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get position")
	}

	position := 1024.0
	switch v := maxPos.(type) {
	case float64:
		position = v + 1024
	case int64:
		position = float64(v) + 1024
	}

	copyTitle := task.Title + " (copy)"
	_, err = h.queries.CreateTask(ctx, sqlc.CreateTaskParams{
		UserID:      userID,
		Title:       copyTitle,
		Description: task.Description,
		Priority:    task.Priority,
		Status:      task.Status,
		Position:    position,
		DueDate:     task.DueDate,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to duplicate task")
	}

	c.Response().Header().Set("HX-Refresh", "true")
	c.Response().Header().Set("HX-Trigger", `{"showToast": {"message": "Task duplicated", "type": "success"}}`)
	return c.NoContent(http.StatusCreated)
}

func (h *Handler) DeleteTask(c echo.Context) error {
	userID := mw.GetUserID(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task id")
	}

	result, err := h.queries.DeleteTask(c.Request().Context(), sqlc.DeleteTaskParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete task")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete task")
	}
	if rows == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "task not found")
	}

	c.Response().Header().Set("HX-Trigger", fmt.Sprintf(`{"showToast": {"message": "Task deleted", "type": "success"}, "taskDeleted": {"id": %d}}`, id))
	return c.NoContent(http.StatusOK)
}
