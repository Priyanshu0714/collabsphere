package database

import (
	"backend/internal/model"
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

func ConnectDatabase() error {
	dsn := "postgres://bhupeshr45@localhost:5432/colab"

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return err
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return err
	}

	db = pool
	log.Println("Connected to PostgreSQL.")

	return nil
}

func InsertUser(u *model.User) error {
	log.Println("Inserting User")

	err := db.QueryRow(context.Background(), "Insert into users(name, email, password) values($1, $2, $3) RETURNING id", u.Name, u.Email, u.Password).Scan(&u.ID)
	if err != nil {
		return err
	}

	return nil
}

func FindUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := db.QueryRow(context.Background(), "select id, name, email, password from users where email = $1", email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		return nil, err
	}

	fmt.Println("User found")

	return &user, nil
}

func GetUserIDByEmail(email string) (int, error) {
	result, err := FindUserByEmail(email)
	if err != nil {
		return -1, err
	}

	return result.ID, nil
}

func FindUserByID(id int) (*model.User, error) {
	var user model.User
	err := db.QueryRow(context.Background(), "select id, name, email, password from users where id = $1", id).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		return nil, err
	}

	fmt.Println("User found")

	return &user, nil
}

func GetUserNameByID(userID int) (string, error) {
	result, err := FindUserByID(userID)
	if err != nil {
		return "", err
	}
	return result.Name, nil
}

func GetUserNameByEmail(email string) (string, error) {
	result, err := FindUserByEmail(email)
	if err != nil {
		return "", err
	}

	return result.Name, nil
}

func GetChatMembers(chatID int) ([]int, error) {
	rows, err := db.Query(context.Background(), "select member_id from chat_members where chat_id = $1", chatID)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()

	var members []int
	for rows.Next() {
		var member int
		rows.Scan(&member)
		members = append(members, member)
	}
	fmt.Println(members)

	return members, nil
}

func GetChatMemberNames(chatID int) ([]string, error) {
	rows, err := db.Query(context.Background(), "select name from users where id in (select member_id from chat_members where chat_id = $1)", chatID)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()

	var members []string
	for rows.Next() {
		var member string
		rows.Scan(&member)
		members = append(members, member)
	}
	fmt.Println(members)

	return members, nil
}

func GetUserChats(userID int) ([]int, error) {
	rows, err := db.Query(context.Background(), "select chat_id from chat_members where member_id = $1", userID)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer rows.Close()

	var chats []int
	for rows.Next() {
		var chat int
		rows.Scan(&chat)
		chats = append(chats, chat)
	}
	fmt.Println(chats)

	return chats, nil
}

func InsertMessage(msg *model.Message) error {
	query := "Insert into messages(chat_id, sender_id, sender_name, content) values($1, $2, $3, $4) RETURNING id"

	err := db.QueryRow(context.Background(), query, msg.ChatId, msg.SenderID, msg.SenderName, msg.Content).Scan(&msg.ID)
	if err != nil {
		return err
	}

	fmt.Println("Inserted a message with id ", msg.ID)

	return nil
}

func GetMessagesByChatID(chatID int) ([]model.Message, error) {
	query := "select chat_id, sender_id, sender_name, content from messages where chat_id = $1"
	rows, err := db.Query(context.Background(), query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var message model.Message
		err := rows.Scan(
			&message.ChatId,
			&message.SenderID,
			&message.SenderName,
			&message.Content,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, nil
}
