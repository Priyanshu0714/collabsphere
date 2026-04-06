package database
//
// import (
// 	"backend/internal/model"
// 	"context"
//
// 	"fmt"
//
// 	"go.mongodb.org/mongo-driver/v2/bson"
// 	"go.mongodb.org/mongo-driver/v2/mongo"
// 	"go.mongodb.org/mongo-driver/v2/mongo/options"
// )
//
// var MongoClient *mongo.Client
//
// // const connectionString = "mongodb+srv://bhupesh178355_db_user:Pacg6J4icXQKdFYu@cluster0.9accbv1.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"
// const connectionString = "mongodb://localhost:27017";
// const db = "collab"
//
// func ConnectDatabase() error {
// 	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
// 	clientOptions := options.Client().ApplyURI(connectionString).SetServerAPIOptions(serverAPI)
// 	client, err := mongo.Connect(clientOptions)
// 	if err != nil {
// 		return err
// 	}
//
// 	// Send a ping to confirm a successful connection
// 	if err := client.Ping(context.TODO(), nil); err != nil {
// 		return err
// 	}
// 	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
//
// 	MongoClient = client
// 	return nil
// }
//
// func DisconnectDatabase() error {
// 	if err := MongoClient.Disconnect(context.TODO()); err != nil {
// 		return err
// 	}
//
// 	fmt.Println("Database Disconnected")
//
// 	return nil
// }
//
// func InsertUser(u *model.User) error {
// 	fmt.Println("Inserting User")
// 	u.ID = bson.NewObjectID()
//
// 	collection := MongoClient.Database(db).Collection("users")
// 	inserted, err := collection.InsertOne(context.TODO(), u)
// 	if err != nil {
// 		fmt.Println("DB ERROR: ", err)
// 		return err
// 	}
// 	fmt.Println("Inserted a user with id ", inserted.InsertedID)
// 	return nil
// }
//
// func FindUserByEmail(email string) (*model.User, error) {
// 	filter := bson.M{"email": email}
// 	collection := MongoClient.Database(db).Collection("users")
//
// 	var user model.User
// 	err := collection.FindOne(context.TODO(), filter).Decode(&user)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return &user, nil
// }
//
// func GetUserIDByEmail(email string) (bson.ObjectID, error) {
// 	result, err := FindUserByEmail(email)
// 	if err != nil {
// 		return bson.NilObjectID, err
// 	}
//
// 	return result.ID, nil
// }
//
// func FindUserByID(userID bson.ObjectID) (*model.User, error) {
// 	filter := bson.M{"_id": userID}
// 	collection := MongoClient.Database(db).Collection("users")
//
// 	var user model.User
// 	err := collection.FindOne(context.TODO(), filter).Decode(&user)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return &user, nil
// }
//
// func GetUserNameByID(userID bson.ObjectID) (string, error) {
// 	result, err := FindUserByID(userID)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	return result.Name, nil
// }
//
// func GetUserNameByEmail(email string) (string, error) {
// 	result, err := FindUserByEmail(email)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	return result.Name, nil
// }
//
// func InsertChat() error {
// 	var email1 string = "bhupesh178355@gmail.com"
// 	var email2 string = "r@gmail.com"
// 	id1, err := GetUserIDByEmail(email1)
// 	id2, err := GetUserIDByEmail(email2)
// 	if id1 == bson.NilObjectID {
// 		fmt.Println("User1 not found")
// 	}
// 	fmt.Println("ID1: ", id1)
// 	if id2 == bson.NilObjectID {
// 		fmt.Println("User2 not found")
// 		return nil
// 	}
// 	fmt.Println("ID2: ", id2)
//
// 	chat := model.Chat{
// 		ID:      bson.NewObjectID(),
// 		IsGroup: false,
// 		Members: []bson.ObjectID{id1, id2},
// 	}
// 	collection := MongoClient.Database(db).Collection("chats")
// 	inserted, err := collection.InsertOne(context.TODO(), chat)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println("Chat ID: ", inserted.InsertedID)
//
// 	return nil
// }
//
// func GetChatMembers(chatID bson.ObjectID) ([]bson.ObjectID, error) {
// 	var result model.Chat
// 	collection := MongoClient.Database(db).Collection("chats")
// 	err := collection.FindOne(context.TODO(), bson.M{"_id": chatID}).Decode(&result)
//
// 	if err != nil {
// 		fmt.Println(err)
// 		return []bson.ObjectID{}, err
// 	}
//
// 	fmt.Println(result.Members)
//
// 	return result.Members, nil
// }
//
// func GetUserChats(userID bson.ObjectID) ([]bson.ObjectID, error) {
// 	filter := bson.M{"members": userID}
// 	opts := options.Find().SetProjection(bson.M{"_id": 1})
// 	collection := MongoClient.Database(db).Collection("chats")
// 	cursor, err := collection.Find(context.TODO(), filter, opts)
// 	if err != nil {
// 		return make([]bson.ObjectID, 0), err
// 	}
//
// 	type chatID struct {
// 		ID bson.ObjectID `bson:"_id"`
// 	}
//
// 	var chats []chatID
// 	if err = cursor.All(context.TODO(), &chats); err != nil {
// 		fmt.Println(err)
// 		return make([]bson.ObjectID, 0), err
// 	}
//
// 	var result []bson.ObjectID
// 	for _, chat := range chats {
// 		result = append(result, chat.ID)
// 	}
//
// 	fmt.Println(chats)
// 	return result, nil
// }
//
// func GetChatMembersNames(chatID bson.ObjectID) ([]string, error) {
// 	members, err := GetChatMembers(chatID)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var result []string
//
// 	for _, member := range members {
// 		name, err := GetUserNameByID(member)
// 		if err != nil {
// 			return nil, err
// 		}
// 		result = append(result, name)
// 	}
//
// 	return result, nil
// }
//
// func InsertMessage(msg *model.Message) error {
// 	collection := MongoClient.Database(db).Collection("messages")
// 	inserted, err := collection.InsertOne(context.TODO(), msg)
// 	if err != nil {
// 		fmt.Println("DB ERROR: ", err)
// 		return err
// 	}
// 	fmt.Println("Inserted a message with id ", inserted.InsertedID)
// 	return nil
//
// }
//
// // func getChatName(chatID bson.ObjectID) (string, error) {
// // 	filter := bson.M{"_id": chatID}
// // 	collection := MongoClient.Database(db).Collection("chats")
// //
// // 	var result model.Chat
// //
// // 	err := collection.FindOne(context.TODO(), filter).Decode(&result)
// // 	if err != nil {
// // 		return "", err
// // 	}
// //
// // 	return result.
// // }
