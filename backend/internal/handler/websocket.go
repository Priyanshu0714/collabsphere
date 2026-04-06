package handler

import (
	"backend/internal/database"
	"backend/internal/model"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// --- Constants for Stability ---
const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024
)

// --- Structs ---

type Client struct {
	UserID   int
	UserName string
	Conn     *websocket.Conn
	Send     chan model.Message
	Hub      *Hub
}

// BroadcastRequest wraps the message and the specific users who should receive it.
type BroadcastRequest struct {
	Message      model.Message
	RecipientIDs []int
}

type Hub struct {
	// Clients map: UserID -> Set of Clients (handling multiple tabs/devices per user)
	Clients    map[int]map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan BroadcastRequest
}

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins (for dev). In prod, restrict this.
	},
}

// --- Hub Logic ---

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[int]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan BroadcastRequest),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// Initialize map for user if not exists
			if h.Clients[client.UserID] == nil {
				h.Clients[client.UserID] = make(map[*Client]bool)
			}
			h.Clients[client.UserID][client] = true
			fmt.Printf("User %s (ID: %d) connected.\n", client.UserName, client.UserID)

		case client := <-h.Unregister:
			if conns, ok := h.Clients[client.UserID]; ok {
				if _, ok := conns[client]; ok {
					delete(conns, client)
					close(client.Send)
				}
				if len(conns) == 0 {
					delete(h.Clients, client.UserID)
				}
			}
			fmt.Printf("User %s (ID: %d) disconnected.\n", client.UserName, client.UserID)

		case req := <-h.Broadcast:
			// Route the message ONLY to the specific members of the chat
			for _, recipientID := range req.RecipientIDs {
				if conns, ok := h.Clients[recipientID]; ok {
					for client := range conns {
						select {
						case client.Send <- req.Message:
						default:
							// Client buffer is full, force disconnect to prevent blocking
							close(client.Send)
							delete(conns, client)
						}
					}
				}
			}
		}
	}
}

// --- Handler Logic ---

// WSHandler handles the initial connection upgrade
// GET /ws
func WSHandler(h *Hub) echo.HandlerFunc {
	return func(c echo.Context) error {
		// 1. Authenticate (JWT)
		user := c.Get("user")
		if user == nil {
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

		// Safer float64 conversion for JWT numbers
		var userID int
		if idFloat, ok := claims["userID"].(float64); ok {
			userID = int(idFloat)
		} else {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid userID in token")
		}

		// 2. Fetch User Details (Do this outside the Hub loop!)
		userName, _ := database.GetUserNameByID(userID)

		// 3. Upgrade Connection
		conn, err := Upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}

		client := &Client{
			UserID:   userID,
			UserName: userName,
			Conn:     conn,
			Send:     make(chan model.Message, 256),
			Hub:      h,
		}

		// 4. Register
		client.Hub.Register <- client

		// 5. Start pumps
		go client.writePump()
		go client.readPump()

		return nil
	}
}

// --- Client Pumps ---

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg model.Message
		// Read JSON. Note: The Client MUST send 'chat_id' in the JSON payload.
		if err := c.Conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error: %v", err)
			}
			break
		}

		// Sanitize/Fill Sender Info
		msg.SenderID = c.UserID
		msg.SenderName = c.UserName
		// msg.Timestamp = time.Now() // Uncomment if you want server-side time

		// --- ROUTING LOGIC ---
		
		// 1. Get all members of the target chat
		members, err := database.GetChatMembers(msg.ChatId)
		if err != nil {
			fmt.Printf("Error fetching members for chat %d: %v\n", msg.ChatId, err)
			continue
		}

		// 2. Validate: Is the sender actually in this chat?
		if !slices.Contains(members, c.UserID) {
			fmt.Printf("Security alert: User %d tried to post to chat %d without membership\n", c.UserID, msg.ChatId)
			continue
		}

		// 3. Persist to Database
		if err := database.InsertMessage(&msg); err != nil {
			fmt.Printf("Error saving message: %v\n", err)
			// Optional: send error back to client
			continue
		}

		// 4. Send to Hub (Hub distributes to online members only)
		c.Hub.Broadcast <- BroadcastRequest{
			Message:      msg,
			RecipientIDs: members,
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The Hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Write the JSON message
			if err := c.Conn.WriteJSON(msg); err != nil {
				return
			}

		case <-ticker.C:
			// Send Ping
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
