package api

import (
	"context"
	"net/http"
	"ossyne/internal/db"
	"ossyne/internal/models"
	"strings"
	"github.com/labstack/echo/v4"
)

type contextKey string

const userContextKey contextKey = "user"
func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization header is required"})
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Authorization header format must be Bearer {token}"})
		}
		token := parts[1]
		var user models.User
		result := db.DB.Where("github_access_token = ?", token).First(&user)
		if result.Error != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or expired token"})
		}
		ctx := context.WithValue(c.Request().Context(), userContextKey, &user)
		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}