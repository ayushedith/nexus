package collab

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Server struct {
	upgrader websocket.Upgrader
	rooms    map[string]*Room
	mu       sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		rooms: make(map[string]*Room),
	}
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		http.Error(w, "room required", http.StatusBadRequest)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("upgrade websocket", "error", err)
		return
	}

	client := &Client{
		id:   generateClientID(),
		conn: conn,
		send: make(chan []byte, 256),
	}

	room := s.getOrCreateRoom(roomID)
	room.register <- client

	go client.writePump()
	go client.readPump(room)
}

func (s *Server) getOrCreateRoom(id string) *Room {
	s.mu.Lock()
	defer s.mu.Unlock()

	if room, ok := s.rooms[id]; ok {
		return room
	}

	room := NewRoom(id)
	s.rooms[id] = room
	go room.Run()

	return room
}

type Room struct {
	id         string
	clients    map[*Client]bool
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewRoom(id string) *Room {
	return &Room{
		id:         id,
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (r *Room) Run() {
	for {
		select {
		case client := <-r.register:
			r.mu.Lock()
			r.clients[client] = true
			r.mu.Unlock()

			r.broadcastPresence()

		case client := <-r.unregister:
			r.mu.Lock()
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.send)
			}
			r.mu.Unlock()

			r.broadcastPresence()

		case message := <-r.broadcast:
			r.mu.RLock()
			for client := range r.clients {
				if client.id != message.ClientID {
					data, _ := json.Marshal(message)
					select {
					case client.send <- data:
					default:
						close(client.send)
						delete(r.clients, client)
					}
				}
			}
			r.mu.RUnlock()
		}
	}
}

func (r *Room) broadcastPresence() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clientIDs := []string{}
	for client := range r.clients {
		clientIDs = append(clientIDs, client.id)
	}

	msg := &Message{
		Type: "presence",
		Data: map[string]interface{}{
			"clients": clientIDs,
		},
	}

	data, _ := json.Marshal(msg)
	for client := range r.clients {
		select {
		case client.send <- data:
		default:
		}
	}
}

type Client struct {
	id   string
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) readPump(room *Room) {
	defer func() {
		room.unregister <- c
		c.conn.Close()
	}()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		msg.ClientID = c.id
		room.broadcast <- &msg
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()

	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}

type Message struct {
	Type     string                 `json:"type"`
	ClientID string                 `json:"clientId,omitempty"`
	Data     map[string]interface{} `json:"data"`
}

func generateClientID() string {
	return fmt.Sprintf("client-%d", generateRandomID())
}

func generateRandomID() int64 {
	return 1000000
}
