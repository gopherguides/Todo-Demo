package handler

import (
	"database/sql"

	"github.com/labstack/echo/v4"

	"todo-demo/internal/config"
	"todo-demo/internal/database/sqlc"
	"todo-demo/internal/meta"
	mw "todo-demo/internal/middleware"
)

type Handler struct {
	db      *sql.DB
	queries *sqlc.Queries
	site    meta.SiteConfig
}

func New(db *sql.DB, cfg config.Config) *Handler {
	return &Handler{
		db:      db,
		queries: sqlc.New(db),
		site: meta.SiteConfig{
			URL:                 cfg.SiteURL,
			ClerkPublishableKey: cfg.ClerkPublishableKey,
		},
	}
}

func (h *Handler) Register(e *echo.Echo) {
	e.GET("/", h.Home)
	e.GET("/health", h.Health)

	e.GET("/sign-in", h.SignIn)
	e.GET("/sign-up", h.SignUp)

	protected := e.Group("", mw.ClerkAuth())
	protected.GET("/board", h.Board)
	protected.GET("/tasks/search", h.SearchTasks)
	protected.POST("/tasks", h.CreateTask)
	protected.GET("/tasks/:id/edit", h.EditTask)
	protected.PUT("/tasks/:id", h.UpdateTask)
	protected.POST("/tasks/:id/duplicate", h.DuplicateTask)
	protected.DELETE("/tasks/:id", h.DeleteTask)
	protected.PUT("/tasks/:id/move", h.MoveTask)
}
