package handler

import (
	"backend/internal/database"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func GetChatMessages(c echo.Context) error {
	chatIDParam := c.Param("chat_id")
	chatID, err := strconv.Atoi(chatIDParam)

	messages, err := database.GetMessagesByChatID(chatID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch messages"})
	}

	return c.JSON(http.StatusOK, messages)
}
