package handler

import (
	"backend/internal/database"
	"backend/internal/model"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func RegisterPage(c echo.Context) error {
	data := map[string]any{}
	return c.Render(http.StatusOK, "register.html", data)
}

func Register(c echo.Context) error {
	u := new(model.User)
	if err := c.Bind(u); err != nil {
		return err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash password")
	}

	u.Password = string(hashed)

	fmt.Println("Registering User")
	if err := database.InsertUser(u); err != nil {
		return c.String(http.StatusInternalServerError, "DB error: "+err.Error())
	}

	// fmt.Println(u.ID, u.Name, u.Email, u.Password)
	fmt.Println("User Registered")
	// return c.Redirect(http.StatusSeeOther, "/")
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Register successful",
	})
}
