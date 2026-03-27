package ws

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"github.com/oscar/oscar/pkg/crypto"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type HubHandler struct {
	hub           *Hub
	tokenManager  *crypto.TokenManager
}

func NewHubHandler(hub *Hub, tokenManager *crypto.TokenManager) *HubHandler {
	return &HubHandler{
		hub:          hub,
		tokenManager: tokenManager,
	}
}

func (h *HubHandler) HandleWebSocket(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return echo.ErrUnauthorized
	}

	payload, err := h.tokenManager.ValidateToken(token)
	if err != nil {
		return echo.ErrUnauthorized
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	clientID := uuid.New().String()
	client := &Client{
		ID:       clientID,
		TenantID: payload.TenantID,
		UserID:   payload.UserID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      h.hub,
	}

	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()

	return nil
}

func (h *HubHandler) BroadcastToTenant(tenantID string, event string, payload interface{}) {
	h.hub.BroadcastToTenant(tenantID, event, payload)
}

func (h *HubHandler) SendToUser(userID string, event string, payload interface{}) {
	h.hub.SendToUser(userID, event, event, payload)
}

func (h *HubHandler) NotifyNewNotification(tenantID, userID, title, body string) {
	h.hub.SendToUser(userID, "notification.new", map[string]interface{}{
		"title": title,
		"body":  body,
		"type":  "notification",
	})
}

func (h *HubHandler) NotifyDealMoved(tenantID string, dealID, fromStage, toStage string) {
	h.hub.BroadcastToTenant(tenantID, "deal.moved", map[string]interface{}{
		"deal_id":     dealID,
		"from_stage":  fromStage,
		"to_stage":    toStage,
	})
}

func (h *HubHandler) NotifyLeadAssigned(tenantID string, personID, ownerID string) {
	h.hub.SendToUser(ownerID, "lead.assigned", map[string]interface{}{
		"person_id": personID,
	})
}

func (h *HubHandler) NotifyAutomationFired(tenantID string, automationID, entityID string) {
	h.hub.BroadcastToTenant(tenantID, "automation.fired", map[string]interface{}{
		"automation_id": automationID,
		"entity_id":    entityID,
	})
}

func init() {
	upgrader.ReadDeadline = 60 * time.Second
	upgrader.WriteDeadline = 60 * time.Second
	upgrader.PongHandler = func(appData string) error {
		return nil
	}
}
