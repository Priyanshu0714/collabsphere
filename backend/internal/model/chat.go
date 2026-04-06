package model

// import "go.mongodb.org/mongo-driver/v2/bson"

type Chat struct {
	ID      int   `json:"id,omitempty" bson:"_id"`
	IsGroup bool  `json:"is_group" bson:"is_group"`
	// Members []int `json:"members" bson:"members"`
}
