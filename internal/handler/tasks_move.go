package handler

import (
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
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to calculate position")
	}

	err = h.queries.UpdateTaskPosition(c.Request().Context(), sqlc.UpdateTaskPositionParams{
		Status:   req.Status,
		Position: position,
		ID:       id,
		UserID:   userID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to move task")
	}

	return c.JSON(http.StatusOK, map[string]any{"status": "ok", "position": position})
}

func (h *Handler) calculatePosition(c echo.Context, userID string, req moveRequest) (float64, error) {
	ctx := c.Request().Context()

	if req.AfterID == 0 && req.BeforeID == 0 {
		maxPos, err := h.queries.GetMaxPosition(ctx, sqlc.GetMaxPositionParams{
			UserID: userID,
			Status: req.Status,
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

	if req.AfterID != 0 && req.BeforeID == 0 {
		afterPos, err := h.queries.GetTaskPosition(ctx, sqlc.GetTaskPositionParams{
			ID:     req.AfterID,
			UserID: userID,
		})
		if err != nil {
			return 0, err
		}
		return afterPos + 1024, nil
	}

	if req.AfterID == 0 && req.BeforeID != 0 {
		beforePos, err := h.queries.GetTaskPosition(ctx, sqlc.GetTaskPositionParams{
			ID:     req.BeforeID,
			UserID: userID,
		})
		if err != nil {
			return 0, err
		}
		return beforePos / 2, nil
	}

	afterPos, err := h.queries.GetTaskPosition(ctx, sqlc.GetTaskPositionParams{
		ID:     req.AfterID,
		UserID: userID,
	})
	if err != nil {
		return 0, err
	}

	beforePos, err := h.queries.GetTaskPosition(ctx, sqlc.GetTaskPositionParams{
		ID:     req.BeforeID,
		UserID: userID,
	})
	if err != nil {
		return 0, err
	}

	return (afterPos + beforePos) / 2, nil
}
