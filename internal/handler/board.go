package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"todo-demo/internal/database/sqlc"
	"todo-demo/internal/meta"
	mw "todo-demo/internal/middleware"
	"todo-demo/templates/pages"
)

func (h *Handler) Board(c echo.Context) error {
	userID := mw.GetUserID(c)

	todoTasks, err := h.queries.ListTasksByStatus(c.Request().Context(), sqlc.ListTasksByStatusParams{
		UserID: userID,
		Status: "todo",
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load tasks")
	}

	inProgressTasks, err := h.queries.ListTasksByStatus(c.Request().Context(), sqlc.ListTasksByStatusParams{
		UserID: userID,
		Status: "in_progress",
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load tasks")
	}

	doneTasks, err := h.queries.ListTasksByStatus(c.Request().Context(), sqlc.ListTasksByStatusParams{
		UserID: userID,
		Status: "done",
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to load tasks")
	}

	page := meta.Page{
		Title:       "Board - Todo Kanban",
		Description: "Your kanban task board.",
		Path:        "/board",
	}

	return Render(c, http.StatusOK, pages.Board(page, h.site, todoTasks, inProgressTasks, doneTasks))
}
