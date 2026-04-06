package handler

import (
	"backend/internal/database"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	// "go.mongodb.org/mongo-driver/v2/bson"
)

func ChatHandler(c echo.Context) error {
	user := c.Get("user")
	if user == nil {
		// return c.JSON(http.StatusUnauthorized, map[string]bool{"authorized": false})
		return echo.NewHTTPError(http.StatusUnauthorized, "missing or invalid JWT")
	}

	token, ok := user.(*jwt.Token)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token type")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid claims")
	}

	// email := claims["email"].(string)
	// userIDHex := claims["userID"].(string)
	// userID, err := bson.ObjectIDFromHex(userIDHex)
	userID := int(claims["userID"].(float64))
	// if err != nil {
	// 	return err
	// }

	name, err := database.GetUserNameByID(userID)
	if err != nil {
		return err
	}

	// id, err := database.GetUserID(email)
	chats, err := database.GetUserChats(userID)
	chatMemberNames := make(map[int][]string)
	for _, chatID := range chats {
		members, err := database.GetChatMembers(chatID)
		if err != nil {
			return err
		}

		for _, memberID := range members {
			name, err := database.GetUserNameByID(memberID)
			if err != nil {
				return err
			}
			chatMemberNames[chatID] = append(chatMemberNames[chatID], name)
		}

	}
	return c.JSON(http.StatusOK, map[string]any{"username": name, "chats": chats, "chatsNames": chatMemberNames})
}
