package handler

import (
	"backend/idms/services"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"time"
	// "errors"
)

type FileHandler struct {
	Minio *services.MinIOService
}

func (h *FileHandler) GetUploadUrl(c echo.Context) error {
	filename := c.QueryParam("filename")
	prefix := c.QueryParam("username")
	// contentType := c.QueryParam("content_type")
	if filename == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "filename parameter is required",
		})
	}
	// link with the userprefix , so that each user gets a dedicated folder
	key := fmt.Sprintf("%s/%d_%s", prefix, time.Now().Unix(), filename)

	fmt.Print("\n", key, "\n")
	var uploadURL string
	var err error
	uploadURL, err = h.Minio.GenerateUploadURL(key)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"upload_url": uploadURL,
		"key":        key,
	})
}

func (h *FileHandler) GetDownloadUrl(c echo.Context) error {
	key := strings.TrimSpace(c.QueryParam("key"))
	bucket := strings.TrimSpace(c.QueryParam("bucket"))
	expiryStr := strings.TrimSpace(c.QueryParam("expiry"))

	if key == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "key parameter is required",
		})
	}

	var expiry time.Duration = 60 * time.Minute // default 1 hour
	if expiryStr != "" {
		if minutes, err := time.ParseDuration(expiryStr + "m"); err == nil {
			expiry = minutes
		}
	}

	downloadURL, err := h.Minio.GenerateDownloadURL(key, bucket, expiry)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"download_url": downloadURL,
		"expires_in":   expiry.String(),
	})
}

func (h *FileHandler) PostMoveFile(c echo.Context) error {
	key := c.FormValue("key")

	if key == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "key parameter is required",
		})
	}

	err := h.Minio.MoveFile(c.Request().Context(), key)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "File moved successfully",
	})
}

func (h *FileHandler) GetFiles(c echo.Context) error {
	bucket := c.QueryParam("bucket")
	prefix := c.QueryParam("prefix")
	// for testing
	if bucket == "" {
		bucket = "tempbucket"
	}
	objects, err := h.Minio.ListFiles(bucket, prefix)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// Convert to a simpler format for JSON response
	var files []map[string]interface{}
	for _, obj := range objects {
		files = append(files, map[string]interface{}{
			"key":           obj.Key,
			"size":          obj.Size,
			"last_modified": obj.LastModified,
			"etag":          obj.ETag,
			"content_type":  obj.ContentType,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"files": files,
	})
}

func (h *FileHandler) GetFileInfo(c echo.Context) error {
	key := c.QueryParam("key")
	bucket := c.QueryParam("bucket")

	if key == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "key parameter is required",
		})
	}
	objInfo, err := h.Minio.GetFileInfo(bucket, key)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"key":           objInfo.Key,
		"size":          objInfo.Size,
		"last_modified": objInfo.LastModified,
		"etag":          objInfo.ETag,
		"content_type":  objInfo.ContentType,
		"metadata":      objInfo.UserMetadata,
	})
}

func (h *FileHandler) DeleteFile(c echo.Context) error {
	key := c.QueryParam("key")
	bucket := c.QueryParam("bucket")

	if key == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "key parameter is required",
		})
	}

	err := h.Minio.DeleteFile(c.Request().Context(), bucket, key)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "File deleted successfully",
	})
}
