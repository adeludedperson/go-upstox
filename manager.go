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

	// Log raw response for debugging
	fmt.Printf("Order Place Response - Status: %d, Body: %s\n", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var orderResp OrderResponse
	if err := json.Unmarshal(body, &orderResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Validate the API response status even if HTTP status is OK
	if orderResp.Status != "success" {
		errorMsg := "Order placement failed"
		if len(orderResp.Errors) > 0 {
			errorMsg = orderResp.Errors[0].Message
		}
		return nil, fmt.Errorf("API returned error status '%s': %s", orderResp.Status, errorMsg)
	}

	// Verify that we have order IDs
	if orderResp.Data == nil || len(orderResp.Data.OrderIDs) == 0 {
		return nil, fmt.Errorf("no order IDs returned in successful response")
	}

	// Wait briefly and get the actual order details to see the real status
	time.Sleep(500 * time.Millisecond)
	
	orderID := orderResp.Data.OrderIDs[0]
	orderDetails, err := m.GetOrderDetails(orderID)
	if err != nil {
		// If we can't get order details, return the original response
		fmt.Printf("Warning: Could not get order details for ID %s: %v\n", orderID, err)
		return &orderResp, nil
	}

	// Create a response with the actual order status
	detailedResponse := &OrderResponse{
		Status: "success",
		Data: &OrderResponseData{
			OrderIDs: orderResp.Data.OrderIDs,
		},
		Metadata: orderResp.Metadata,
	}

	// If order was rejected, add error details
	if orderDetails.Status == "rejected" {
		detailedResponse.Status = "error"
		detailedResponse.Errors = []OrderError{{
			ErrorCode: "ORDER_REJECTED",
			Message:   orderDetails.StatusMessage,
		}}
	}

	return detailedResponse, nil
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

func (m *Manager) NewWebSocketManager(instrumentKeys []string, onPriceUpdate func(string, float64, *int32)) (*WebSocketManager, error) {
	wsURL, err := m.getAuthorizedWebSocketURL()
	if err != nil {
		return nil, fmt.Errorf("failed to get authorized WebSocket URL: %w", err)
	}

	config := WebSocketConfig{
		InstrumentKeys: instrumentKeys,
		Token:          m.accessToken,
	}

	return NewWebSocketManager(wsURL, config, onPriceUpdate), nil
}

func (m *Manager) getAuthorizedWebSocketURL() (string, error) {
	authorizeURL := "https://api.upstox.com/v3/feed/market-data-feed/authorize"
	
	req, err := http.NewRequest("GET", authorizeURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+m.accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var authResp AuthorizeResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return "", err
	}

	if authResp.Status != "success" {
		return "", fmt.Errorf("authorization failed: %s", authResp.Status)
	}

	return authResp.Data.AuthorizedRedirectURI, nil
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

func (m *Manager) GetFundsAndMargin(segment ...string) (*FundsResponse, error) {
	url := "https://api.upstox.com/v2/user/get-funds-and-margin"
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if len(segment) > 0 {
		q := req.URL.Query()
		q.Add("segment", segment[0])
		req.URL.RawQuery = q.Encode()
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

	var fundsResp FundsResponse
	if err := json.Unmarshal(body, &fundsResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if fundsResp.Status != "success" {
		return nil, fmt.Errorf("API returned error status: %s", fundsResp.Status)
	}

	return &fundsResp, nil
}
