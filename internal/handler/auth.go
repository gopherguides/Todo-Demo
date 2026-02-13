package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"todo-demo/internal/meta"
	"todo-demo/templates/pages"
)

func (h *Handler) SignIn(c echo.Context) error {
	page := meta.Page{
		Title:       "Sign In - Todo Kanban",
		Description: "Sign in to your Todo Kanban account.",
		Path:        "/sign-in",
	}
	return Render(c, http.StatusOK, pages.SignIn(page, h.site))
}

func (h *Handler) SignUp(c echo.Context) error {
	page := meta.Page{
		Title:       "Sign Up - Todo Kanban",
		Description: "Create your Todo Kanban account.",
		Path:        "/sign-up",
	}
	return Render(c, http.StatusOK, pages.SignUp(page, h.site))
}
