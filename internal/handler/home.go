package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"todo-demo/internal/meta"
	mw "todo-demo/internal/middleware"
	"todo-demo/templates/pages"
)

func (h *Handler) Home(c echo.Context) error {
	userID := mw.GetUserID(c)
	if userID != "" {
		return c.Redirect(http.StatusFound, "/board")
	}

	page := meta.Page{
		Title:       "Todo Kanban - Organize Your Tasks",
		Description: "A modern kanban-style task manager to organize your work.",
		Path:        "/",
	}
	return Render(c, http.StatusOK, pages.Home(page, h.site))
}

func (h *Handler) Health(c echo.Context) error {
	if err := h.db.Ping(); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "unhealthy"})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
