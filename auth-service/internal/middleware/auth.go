// internal/middleware/auth.go
package middleware

import (
	"auth-service/internal/service"
	"strings"

	"github.com/labstack/echo/v4"
)

func AuthMiddleware(authService *service.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.ErrUnauthorized
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return echo.ErrUnauthorized
			}

			token := parts[1]
			_, err := authService.ValidateToken(c.Request().Context(), token)
			if err != nil {
				return echo.ErrUnauthorized
			}

			return next(c)
		}
	}
}
