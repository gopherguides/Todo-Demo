package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/labstack/echo/v4"

	"todo-demo/internal/ctxkeys"
)

func isHTMXRequest(c echo.Context) bool {
	return c.Request().Header.Get("HX-Request") == "true"
}

func wantsJSON(c echo.Context) bool {
	accept := c.Request().Header.Get("Accept")
	return strings.Contains(accept, "application/json")
}

func unauthenticatedResponse(c echo.Context, message string) error {
	if isHTMXRequest(c) {
		c.Response().Header().Set("HX-Redirect", "/sign-in")
		return c.NoContent(http.StatusUnauthorized)
	}
	if wantsJSON(c) {
		return echo.NewHTTPError(http.StatusUnauthorized, message)
	}
	return c.Redirect(http.StatusFound, "/sign-in")
}

func ClerkAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sessionToken, err := c.Cookie("__session")
			if err != nil {
				return unauthenticatedResponse(c, "authentication required")
			}

			claims, err := jwt.Verify(c.Request().Context(), &jwt.VerifyParams{
				Token: sessionToken.Value,
			})
			if err != nil {
				return unauthenticatedResponse(c, "invalid session")
			}

			userID := claims.Subject
			ctx := context.WithValue(c.Request().Context(), ctxkeys.UserID, userID)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

func GetUserID(c echo.Context) string {
	userID, _ := c.Request().Context().Value(ctxkeys.UserID).(string)
	return userID
}

func OptionalClerkAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sessionToken, err := c.Cookie("__session")
			if err != nil {
				return next(c)
			}

			claims, err := jwt.Verify(c.Request().Context(), &jwt.VerifyParams{
				Token: sessionToken.Value,
			})
			if err != nil {
				return next(c)
			}

			userID := claims.Subject
			ctx := context.WithValue(c.Request().Context(), ctxkeys.UserID, userID)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

func ClerkInit(secretKey string) {
	clerk.SetKey(secretKey)
}
