package main

import (
	"backend/idms/database"
	"backend/idms/handler"
	"backend/idms/middleware"
	"backend/idms/services"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error while loading the env file")
	}

	// Get environment variables
	endpoint := os.Getenv("ENDPOINT")
	accessKey := os.Getenv("ACCESS_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	tempStore := os.Getenv("TEMP_STORE")
	mainStore := os.Getenv("MAIN_STORE")
	useSSLStr := os.Getenv("USE_SSL")

	if endpoint == "" || accessKey == "" || secretKey == "" || tempStore == "" || mainStore == "" {
		log.Fatal("Missing required environment variables")
	}

	useSSL := useSSLStr == "true"
	fmt.Printf("Config loaded - Endpoint: %s,\n\n SSL: %t,\n\n Temp: %s,\n\n Main: %s\n", endpoint, useSSL, tempStore, mainStore)

	// Initialize PostgreSQL
	if err := database.InitPostgres(); err != nil {
		log.Fatal("Failed to initialize PostgreSQL:", err)
	}

	// Initialize MinIO service
	minioService, err := services.NewMinIOService(endpoint, accessKey, secretKey, tempStore, mainStore, useSSL)
	if err != nil {
		log.Fatal("Failed to initialize MinIO service:", err)
	}

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())
	e.Use(echomw.CORS())

	// Public routes (no auth required)
	api := e.Group("/api")
	api.POST("/signup", handler.Signup)
	api.POST("/login", handler.Login)

	// TESTING for the get request
	e.GET("/test", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "OK",
		})
	})

	// Protected routes (JWT required)
	protected := api.Group("", middleware.JWTAuth)

	filehandler := handler.FileHandler{
		Minio: minioService,
	}

	// all the request assigned to the file-handler
	protected.GET("/upload-url", filehandler.GetUploadUrl)
	protected.GET("/download-url", filehandler.GetDownloadUrl)
	protected.POST("/move-file", filehandler.PostMoveFile)
	protected.GET("/files", filehandler.GetFiles)
	protected.GET("/file-info", filehandler.GetFileInfo)
	protected.DELETE("/file", filehandler.DeleteFile)

	processhandler := handler.ProcessHandler{
		Minio: minioService,
	}
	// processing pipeline request assigned to the processing-handler
	protected.POST("/process-file", processhandler.PostProcessFile)

	client, err := mongo.Connect(options.Client().ApplyURI(os.Getenv("MONGO_URL")))
	if err != nil {
		log.Fatal(err)
	}

	queryhandler := handler.QueryHandler{
		Minio: minioService,
		Mongo: client,
	}
	// query request
	protected.POST("/query", queryhandler.QueryProcessing)

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
