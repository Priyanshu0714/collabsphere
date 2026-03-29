package main

import (
	"backend/idms/handler"
	"backend/idms/services"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error while loading the env file")
	}

	// Get environment variables
	endpoint := os.Getenv("ENDPOINT")           // e.g., "localhost:9000"
	accessKey := os.Getenv("ACCESS_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	tempStore := os.Getenv("TEMP_STORE")
	mainStore := os.Getenv("MAIN_STORE")
	useSSLStr := os.Getenv("USE_SSL")           // "true" or "false"

	if endpoint == "" || accessKey == "" || secretKey == "" || tempStore == "" || mainStore == "" {
		log.Fatal("Missing required environment variables")
	}

	useSSL := useSSLStr == "true"
	fmt.Printf("Config loaded - Endpoint: %s,\n\n SSL: %t,\n\n Temp: %s,\n\n Main: %s\n", endpoint, useSSL, tempStore, mainStore)

	// Initialize MinIO service
	minioService, err := services.NewMinIOService(endpoint, accessKey, secretKey, tempStore, mainStore, useSSL)
	if err != nil {
		log.Fatal("Failed to initialize MinIO service:", err)
	}

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())


	// Routes
	api := e.Group("/api")

	// TESTING for the get request
	e.GET("/test",func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":"OK",
		});
	})

	filehandler :=handler.FileHandler{
		Minio:minioService,
	}

	// all the request assigined to the file-handler
	api.GET("/upload-url", filehandler.GetUploadUrl)
	api.GET("/download-url", filehandler.GetDownloadUrl)
	api.POST("/move-file", filehandler.PostMoveFile)
	api.GET("/files", filehandler.GetFiles)
	api.GET("/file-info", filehandler.GetFileInfo)
	api.DELETE("/file", filehandler.DeleteFile)

	processhandler:=handler.ProcessHandler{
		Minio:minioService,
	}
	// processing pileline request assigned to the processing-handler
	api.POST("/process-file",processhandler.PostProcessFile)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s\n", port)
	if err := e.Start(":" + port); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}