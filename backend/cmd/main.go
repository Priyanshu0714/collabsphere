package main

import (
	"backend/internal/database"
	"backend/internal/handler"
	"html/template"
	"io"
	"log"
	"net/http"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data any, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	t := &Template{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}

	e.Renderer = t

	if err := database.ConnectDatabase(); err != nil {
		log.Fatal("failed to connect db:", err)
	}
	// defer database.DisconnectDatabase()

	hub := handler.NewHub()
	go hub.Run()

	e.GET("/register", handler.RegisterPage)
	e.POST("/register", handler.Register)
	// e.GET("/login", handler.LoginPage)
	e.POST("/login", handler.Login)

	r := e.Group("")
	r.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey:  handler.JWTSecret,
		TokenLookup: "cookie:token",
	}))

	r.GET("/", handler.ChatHandler)
	r.GET("/ws", handler.WSHandler(hub))
	r.GET("/logout", handler.Logout)
	r.GET("/chats/:chat_id/messages", handler.GetChatMessages)

	e.Logger.Fatal(e.Start(":8000"))
}
