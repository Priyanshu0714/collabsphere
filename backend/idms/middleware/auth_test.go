package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func TestJWTAuthMiddleware(t *testing.T) {
	secret := "testsecret123"
	os.Setenv("JWT_SECRET", secret)

	e := echo.New()

	handler := JWTAuth(func(c echo.Context) error {
		username, _ := c.Get("username").(string)
		return c.String(http.StatusOK, "welcome "+username)
	})

	// Test 1: Missing auth header
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = handler(c)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing header, got %d", rec.Code)
	}

	// Test 2: Valid token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "priyanshu",
		"exp":      time.Now().Add(time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req2.Header.Set("Authorization", "Bearer "+tokenStr)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	_ = handler(c2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid token, got %d", rec2.Code)
	}
	if rec2.Body.String() != "welcome priyanshu" {
		t.Fatalf("unexpected body: %s", rec2.Body.String())
	}

	// Test 3: Invalid token secret
	badToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "hacker",
		"exp":      time.Now().Add(time.Hour).Unix(),
	})
	badTokenStr, _ := badToken.SignedString([]byte("wrongsecret"))
	req3 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req3.Header.Set("Authorization", "Bearer "+badTokenStr)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)
	_ = handler(c3)
	if rec3.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid signature, got %d", rec3.Code)
	}
}

