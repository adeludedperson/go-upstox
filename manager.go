package upstox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Manager struct {
	clientID     string
	clientSecret string
	accessToken  string
	httpClient   *http.Client
}

func NewManager(clientID, clientSecret, accessToken string) *Manager {
	return &Manager{
		clientID:     clientID,
		clientSecret: clientSecret,
		accessToken:  accessToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (m *Manager) PlaceMarketOrder(instrumentToken string, quantity int, side string) (*OrderResponse, error) {
	orderReq := OrderRequest{
		Quantity:          quantity,
		Product:           string(ProductIntraday),
		Validity:          string(ValidityDay),
		Price:             0,
		InstrumentToken:   instrumentToken,
		OrderType:         string(OrderTypeMarket),
		TransactionType:   side,
		DisclosedQuantity: 0,
		TriggerPrice:      0,
		IsAMO:             false,
		Slice:             true,
	}

	return m.placeOrder(orderReq)
}

func (m *Manager) PlaceBuyOrder(instrumentToken string, quantity int) (*OrderResponse, error) {
	return m.PlaceMarketOrder(instrumentToken, quantity, string(OrderSideBuy))
}

func (m *Manager) PlaceSellOrder(instrumentToken string, quantity int) (*OrderResponse, error) {
	return m.PlaceMarketOrder(instrumentToken, quantity, string(OrderSideSell))
}

func (m *Manager) placeOrder(orderReq OrderRequest) (*OrderResponse, error) {
	url := "https://api-hft.upstox.com/v3/order/place"

	reqBody, err := json.Marshal(orderReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal order request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+m.accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var orderResp OrderResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &orderResp, nil
}

func (m *Manager) GetPositions() ([]Position, error) {
	url := "https://api.upstox.com/v2/portfolio/short-term-positions"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+m.accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var posResp PositionResponse
	if err := json.Unmarshal(body, &posResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return posResp.Data, nil
}

func (m *Manager) ClosePosition(instrumentToken string) (*OrderResponse, error) {
	positions, err := m.GetPositions()
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %w", err)
	}

	var targetPosition *Position
	for _, pos := range positions {
		if pos.InstrumentToken == instrumentToken && pos.Quantity != 0 {
			targetPosition = &pos
			break
		}
	}

	if targetPosition == nil {
		return nil, fmt.Errorf("no position found for instrument token: %s", instrumentToken)
	}

	var side string
	quantity := targetPosition.Quantity
	if quantity > 0 {
		side = string(OrderSideSell)
	} else {
		side = string(OrderSideBuy)
		quantity = -quantity
	}

	return m.PlaceMarketOrder(instrumentToken, quantity, side)
}

func (m *Manager) CloseAllPositions() ([]OrderResponse, error) {
	url := "https://api.upstox.com/v2/order/positions/exit"

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+m.accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var exitResp OrderResponse
	if err := json.Unmarshal(body, &exitResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var responses []OrderResponse
	responses = append(responses, exitResp)
	return responses, nil
}

func (m *Manager) GetOrderBook() ([]Order, error) {
	url := "https://api.upstox.com/v2/order/retrieve-all"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+m.accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var orderBookResp OrderBookResponse
	if err := json.Unmarshal(body, &orderBookResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return orderBookResp.Data, nil
}

func (m *Manager) GetOrderDetails(orderID string) (*Order, error) {
	url := fmt.Sprintf("https://api.upstox.com/v2/order/details?order_id=%s", orderID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+m.accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var orderDetailResp OrderDetailResponse
	if err := json.Unmarshal(body, &orderDetailResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &orderDetailResp.Data, nil
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
