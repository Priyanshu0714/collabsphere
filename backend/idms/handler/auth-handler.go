package handler

import (
	"backend/idms/database"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Signup(c echo.Context) error {
	var req AuthRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "username and password are required"})
	}
	if len(req.Username) < 3 || len(req.Password) < 6 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "username min 3 chars, password min 6 chars"})
	}

	if err := database.CreateUser(req.Username, req.Password); err != nil {
		// Unique username
		return c.JSON(http.StatusConflict, map[string]string{"error": "username already exists"})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "user created"})
}

func Login(c echo.Context) error {
	var req AuthRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "username and password are required"})
	}

	ok, err := database.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid username or password"})
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": req.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	secret := os.Getenv("JWT_SECRET")
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token":    tokenStr,
		"username": req.Username,
	})
}

