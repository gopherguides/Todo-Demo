package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"todo-demo/internal/middleware"
)

func TestClerkAuthRedirectsBrowserRequests(t *testing.T) {
	e := echo.New()
	e.Use(middleware.ClerkAuth())
	e.GET("/protected", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, rec.Code)
	}
	if rec.Header().Get("Location") != "/sign-in" {
		t.Fatalf("expected Location /sign-in, got %q", rec.Header().Get("Location"))
	}
}

func TestClerkAuthReturnsUnauthorizedForJSONRequests(t *testing.T) {
	e := echo.New()
	e.Use(middleware.ClerkAuth())
	e.GET("/protected", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if rec.Header().Get("HX-Redirect") != "" {
		t.Fatalf("expected no HX-Redirect header, got %q", rec.Header().Get("HX-Redirect"))
	}
}

func TestClerkAuthRedirectsHTMXRequests(t *testing.T) {
	e := echo.New()
	e.Use(middleware.ClerkAuth())
	e.GET("/protected", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if rec.Header().Get("HX-Redirect") != "/sign-in" {
		t.Fatalf("expected HX-Redirect /sign-in, got %q", rec.Header().Get("HX-Redirect"))
	}
}

func TestClerkAuthRedirectsHTMXTaskMutation(t *testing.T) {
	e := echo.New()
	e.Use(middleware.ClerkAuth())
	e.POST("/tasks", func(c echo.Context) error {
		return c.NoContent(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader("title=task"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Header.Set("HX-Request", "true")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if rec.Header().Get("HX-Redirect") != "/sign-in" {
		t.Fatalf("expected HX-Redirect /sign-in, got %q", rec.Header().Get("HX-Redirect"))
	}
}
