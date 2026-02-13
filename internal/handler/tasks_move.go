package handler

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"todo-demo/internal/database/sqlc"
	mw "todo-demo/internal/middleware"
)

type moveRequest struct {
	Status   string `json:"status"`
	AfterID  int64  `json:"after_id"`
	BeforeID int64  `json:"before_id"`
}

func (h *Handler) MoveTask(c echo.Context) error {
	userID := mw.GetUserID(c)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid task id")
	}

	var req moveRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	validStatuses := map[string]bool{"todo": true, "in_progress": true, "done": true}
	if !validStatuses[req.Status] {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid status")
	}

	position, err := h.calculatePosition(c, userID, req)
	if err != nil {
		if isStaleNeighborError(err) {
			position, err = h.nextPosition(c, userID, req.Status)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to calculate position")
			}
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to calculate position")
		}
	}

	result, err := h.queries.UpdateTaskPosition(c.Request().Context(), sqlc.UpdateTaskPositionParams{
		Status:   req.Status,
		Position: position,
		ID:       id,
		UserID:   userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to move task")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to move task")
	}
	if rows == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "task not found")
	}

	return c.JSON(http.StatusOK, map[string]any{"status": "ok", "position": position})
}

func (h *Handler) calculatePosition(c echo.Context, userID string, req moveRequest) (float64, error) {
	if req.AfterID == 0 && req.BeforeID == 0 {
		return h.nextPosition(c, userID, req.Status)
	}

	ctx := c.Request().Context()

	if req.AfterID != 0 && req.BeforeID == 0 {
		afterPos, err := h.positionForStatus(ctx, userID, req.Status, req.AfterID)
		if err != nil {
			return 0, err
		}
		return afterPos + 1024, nil
	}

	if req.AfterID == 0 && req.BeforeID != 0 {
		beforePos, err := h.positionForStatus(ctx, userID, req.Status, req.BeforeID)
		if err != nil {
			return 0, err
		}
		return beforePos / 2, nil
	}

	afterPos, err := h.positionForStatus(ctx, userID, req.Status, req.AfterID)
	if err != nil {
		return 0, err
	}

	beforePos, err := h.positionForStatus(ctx, userID, req.Status, req.BeforeID)
	if err != nil {
		return 0, err
	}

	return (afterPos + beforePos) / 2, nil
}

func (h *Handler) nextPosition(c echo.Context, userID string, status string) (float64, error) {
	maxPos, err := h.queries.GetMaxPosition(c.Request().Context(), sqlc.GetMaxPositionParams{
		UserID: userID,
		Status: status,
	})
	if err != nil {
		return 0, err
	}
	switch v := maxPos.(type) {
	case float64:
		return v + 1024, nil
	case int64:
		return float64(v) + 1024, nil
	default:
		return 1024, nil
	}
}

func isStaleNeighborError(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func (h *Handler) positionForStatus(ctx context.Context, userID string, status string, id int64) (float64, error) {
	task, err := h.queries.GetTask(ctx, sqlc.GetTaskParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return 0, err
	}
	if task.Status != status {
		return 0, sql.ErrNoRows
	}
	return task.Position, nil
}
