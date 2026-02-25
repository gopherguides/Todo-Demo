package handler

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"

	"todo-demo/internal/database/sqlc"
	mw "todo-demo/internal/middleware"
	"todo-demo/templates/components"
)

func (h *Handler) SearchTasks(c echo.Context) error {
	userID := mw.GetUserID(c)

	q := c.QueryParam("q")
	if q == "" {
		return c.NoContent(http.StatusOK)
	}

	tasks, err := h.queries.SearchTasks(c.Request().Context(), sqlc.SearchTasksParams{
		UserID: userID,
		Query:  sql.NullString{String: q, Valid: true},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to search tasks")
	}

	return Render(c, http.StatusOK, components.SearchResults(tasks, q))
}
