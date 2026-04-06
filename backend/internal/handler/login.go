package handler

import (
	"backend/internal/database"
	"backend/internal/model"
	"fmt"
	"net/http"

	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt/v5"
)

var JWTSecret = []byte("supersecretkey")

func LoginPage(c echo.Context) error {
	data := map[string]any{}
	return c.Render(http.StatusOK, "login.html", data)
}

func Login(c echo.Context) error {
	u := new(model.User)
	if err := c.Bind(u); err != nil {
		return err
	}

	user, err := database.FindUserByEmail(u.Email)
	fmt.Println("User ID: ", user.ID)
	fmt.Printf("User email type %T: \n", u.Email)
	if err != nil {
		fmt.Println("DB Error: ", err)
		return echo.ErrInternalServerError
	}

	if user == nil {
		fmt.Println("Email not found")
		return echo.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password)); err != nil {
		fmt.Println("Incorrect Password")
		return echo.ErrUnauthorized
	}

	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	claims := jwt.MapClaims{
		"userID": user.ID,
		"exp":    expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString(JWTSecret)
	if err != nil {
		return err
	}

	cookie := new(http.Cookie)
	cookie.Name = "token"
	cookie.Value = t
	cookie.Expires = expiresAt
	cookie.HttpOnly = true // Not accessible via JS
	cookie.Secure = false  // set true in production with HTTPS
	c.SetCookie(cookie)

	fmt.Println("Logged In")

	// return c.Redirect(http.StatusSeeOther, "/")
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Login successful",
	})
}

func Logout(c echo.Context) error {
	cookie := new(http.Cookie)
	cookie.Name = "token"
	cookie.Value = ""
	cookie.Expires = time.Unix(0, 0)
	cookie.HttpOnly = true
	c.SetCookie(cookie)

	fmt.Println("Logged Out")

	return c.Redirect(http.StatusSeeOther, "/login")
}
