package model

// import "go.mongodb.org/mongo-driver/v2/bson"

type User struct {
	ID       int    `json:"id,omitempty" bson:"_id"`
	Name     string `json:"name" bson:"name"`
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
}
