package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func JWTAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		auth := c.Request().Header.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "missing or invalid authorization header",
			})
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		secret := os.Getenv("JWT_SECRET")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid or expired token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid token claims",
			})
		}

		username, _ := claims["username"].(string)
		if username == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "username not found in token",
			})
		}

		// Set username in context for handlers
		c.Set("username", username)
		return next(c)
	}
}

