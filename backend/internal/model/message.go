package model

// import "go.mongodb.org/mongo-driver/v2/bson"

type Message struct {
	ID         int    `json:"id,omitempty" bson:"_id"`
	ChatId     int    `json:"chat_id" bson:"chat_id"`
	SenderID   int    `json:"sender_id" bson:"sender_id"`
	SenderName string `json:"sender_name" bson:"sender_name"`
	Content    string `json:"content" bson:"content"`
}
