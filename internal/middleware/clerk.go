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

func isAPIRequest(c echo.Context) bool {
	if c.Request().Header.Get("HX-Request") == "true" {
		return true
	}
	accept := c.Request().Header.Get("Accept")
	return strings.Contains(accept, "application/json")
}

func ClerkAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sessionToken, err := c.Cookie("__session")
			if err != nil {
				if isAPIRequest(c) {
					return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
				}
				return c.Redirect(http.StatusFound, "/sign-in")
			}

			claims, err := jwt.Verify(c.Request().Context(), &jwt.VerifyParams{
				Token: sessionToken.Value,
			})
			if err != nil {
				if isAPIRequest(c) {
					return echo.NewHTTPError(http.StatusUnauthorized, "invalid session")
				}
				return c.Redirect(http.StatusFound, "/sign-in")
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
