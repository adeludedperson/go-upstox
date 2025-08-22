package upstox

import (
	"errors"
)

type Manager struct {
	clientID     string
	clientSecret string
	accessToken  string
}

func NewManager(clientID, clientSecret, accessToken string) *Manager {
	return &Manager{
		clientID:     clientID,
		clientSecret: clientSecret,
		accessToken:  accessToken,
	}
}

func (m *Manager) PlaceOrder(side OrderSide, quantity int, instrumentKey string) error {
	return errors.New("PlaceOrder method is not implemented yet")
}

func (m *Manager) NewWebSocket() *WebSocket {
	return &WebSocket{
		manager:       m,
		subscriptions: make(map[string]InstrumentSubscription),
		done:          make(chan struct{}),
	}
}

func (m *Manager) GetAccessToken() string {
	return m.accessToken
}

func (m *Manager) GetClientID() string {
	return m.clientID
}

func (m *Manager) GetClientSecret() string {
	return m.clientSecret
}