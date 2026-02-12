package middleware

import (
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func Common() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		echomw.Recover(),
		echomw.RequestID(),
		echomw.Secure(),
		echomw.GzipWithConfig(echomw.GzipConfig{
			Level: 5,
		}),
	}
}
