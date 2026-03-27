package ws

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID       string
	TenantID string
	UserID   string
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *Hub
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client
	rooms   map[string]map[string]*Client
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
		rooms:   make(map[string]map[string]*Client),
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.ID] = client

	tenantRoom := "tenant:" + client.TenantID
	if h.rooms[tenantRoom] == nil {
		h.rooms[tenantRoom] = make(map[string]*Client)
	}
	h.rooms[tenantRoom][client.ID] = client
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.ID]; ok {
		delete(h.clients, client.ID)
		close(client.Send)

		tenantRoom := "tenant:" + client.TenantID
		if room, ok := h.rooms[tenantRoom]; ok {
			delete(room, client.ID)
			if len(room) == 0 {
				delete(h.rooms, tenantRoom)
			}
		}
	}
}

func (h *Hub) BroadcastToTenant(tenantID string, event string, payload interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room := "tenant:" + tenantID
	clients, ok := h.rooms[room]
	if !ok {
		return
	}

	message := Message{
		Type:    event,
		Payload: payload,
	}

	data, _ := json.Marshal(message)

	for _, client := range clients {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			delete(h.clients, client.ID)
			delete(clients, client.ID)
		}
	}
}

func (h *Hub) SendToUser(userID string, event string, payload interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	message := Message{
		Type:    event,
		Payload: payload,
	}

	data, _ := json.Marshal(message)

	for _, client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.clients, client.ID)
			}
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "subscribe":
			c.Hub.Register(c)
		case "unsubscribe":
			c.Hub.Unregister(c)
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Event struct {
	Type      string      `json:"type"`
	TenantID  string      `json:"tenant_id"`
	EntityID  string      `json:"entity_id"`
	EntityType string     `json:"entity_type"`
	Payload   interface{} `json:"payload"`
}
