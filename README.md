# Go-Upstox

A comprehensive Go package for Upstox broker integration with real-time market data streaming and order management.

> **Note**: This package was created for personal use in my trading project. Feel free to fork, modify, or use it however you want

## Features

### Market Data (WebSocket)
- **Real-time Streaming**: Market data using Upstox Market Data Feed V3
- **Subscription Management**: Dynamic subscription/unsubscription without reconnecting
- **Multiple Data Modes**: Support for LTPC, Full, Option Greeks, and Full D30 modes
- **Typed Callbacks**: Clean, typed Go structs for market data
- **Automatic Protobuf Handling**: Internal protobuf encoding/decoding
- **Connection Management**: Automatic authorization flow and connection handling

### Order Management (REST API)
- **Order Placement**: Market orders with auto-slicing support
- **Position Management**: Real-time position tracking and closure
- **Order Tracking**: Complete order book and individual order details
- **Convenience Methods**: Simple buy/sell order placement
- **Error Handling**: Comprehensive error responses from API
- **Emergency Exit**: Close all positions with a single call

## Installation

```bash
go get github.com/adeludedperson/go-upstox
```

## Usage

### Order Management

#### Basic Order Placement

```go
package main

import (
    "fmt"
    "log"

    "github.com/adeludedperson/go-upstox"
)

func main() {
    // Create manager instance
    manager := upstox.NewManager("your-client-id", "your-client-secret", "your-access-token")

    // Place a buy order for 1 share of SBIN
    instrumentToken := "NSE_EQ|INE062A01020"
    buyResp, err := manager.PlaceBuyOrder(instrumentToken, 1)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Buy order placed! Order ID: %s\n", buyResp.Data.OrderIDs[0])

    // Place a sell order for 1 share
    sellResp, err := manager.PlaceSellOrder(instrumentToken, 1)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Sell order placed! Order ID: %s\n", sellResp.Data.OrderIDs[0])

    // Alternative: Use PlaceMarketOrder directly
    resp, err := manager.PlaceMarketOrder(instrumentToken, 2, "BUY")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Market order placed! Order ID: %s\n", resp.Data.OrderIDs[0])
}
```

#### Position Management

```go
// Get current positions
positions, err := manager.GetPositions()
if err != nil {
    log.Fatal(err)
}

for _, pos := range positions {
    fmt.Printf("Symbol: %s, Qty: %d, P&L: %.2f\n", 
        pos.TradingSymbol, pos.Quantity, pos.PNL)
}

// Close a specific position
closeResp, err := manager.ClosePosition("NSE_EQ|INE062A01020")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Position closed! Order ID: %s\n", closeResp.Data.OrderIDs[0])

// Emergency: Close all positions
responses, err := manager.CloseAllPositions()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("All positions closed! %d orders placed\n", len(responses))
```

#### Order Tracking

```go
// Get order book (all orders for the day)
orders, err := manager.GetOrderBook()
if err != nil {
    log.Fatal(err)
}

for _, order := range orders {
    fmt.Printf("Order ID: %s, Symbol: %s, Status: %s\n", 
        order.OrderID, order.TradingSymbol, order.Status)
}

// Get specific order details
orderDetail, err := manager.GetOrderDetails("231019025057849")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Order Status: %s, Filled: %d/%d\n", 
    orderDetail.Status, orderDetail.FilledQuantity, orderDetail.Quantity)
```

### WebSocket Market Data

#### Basic Setup

```go
package main

import (
    "fmt"
    "log"
    "time"

    "github.com/adeludedperson/go-upstox"
)

func main() {
    // Create manager instance
    manager := upstox.NewManager("your-client-id", "your-client-secret", "your-access-token")

    // Create WebSocket instance
    ws := manager.NewWebSocket()

    // Set up market info callback
    ws.OnMarketInfo(func(info upstox.MarketInfoMessage) {
        fmt.Printf("Market Info - Current Time: %d\n", info.CurrentTS)
        for segment, status := range info.MarketInfo.SegmentStatus {
            fmt.Printf("  %s: %s\n", segment, status)
        }
    })

    // Set up live feed callback
    ws.OnLiveFeed(func(feed upstox.LiveFeedMessage) {
        fmt.Printf("Live Feed - Current Time: %d\n", feed.CurrentTS)
        for symbol, data := range feed.Feeds {
            if data.LTPC != nil {
                fmt.Printf("  Symbol: %s, LTP: %.2f, LTQ: %d, CP: %.2f\n",
                    symbol, data.LTPC.LTP, data.LTPC.LTQ, data.LTPC.CP)
            }
        }
    })

    // Connect to WebSocket
    if err := ws.Connect(); err != nil {
        log.Fatal("Failed to connect:", err)
    }
    defer ws.Close()

    // Subscribe to instruments (default LTPC mode)
    err := ws.Subscribe("NSE_EQ|INE062A01020", "NSE_EQ|INE467B01029")
    if err != nil {
        log.Fatal("Failed to subscribe:", err)
    }

    // Subscribe with specific mode
    err = ws.SubscribeWithMode("full", "NSE_INDEX|Nifty Bank")
    if err != nil {
        log.Fatal("Failed to subscribe with mode:", err)
    }

    // Keep the program running
    time.Sleep(30 * time.Second)
}
```

### Advanced Usage

#### Different Subscription Modes

```go
// LTPC mode (default) - Latest Trading Price and Close Price
ws.Subscribe("NSE_EQ|INE062A01020")

// Full mode - Includes LTPC, market depth, option greeks, OHLC data
ws.SubscribeWithMode("full", "NSE_EQ|INE062A01020")

// Option Greeks mode - LTPC with option greeks data
ws.SubscribeWithMode("option_greeks", "NSE_FO|45450")

// Full D30 mode - Full data with 30 levels of market depth (requires Upstox Plus)
ws.SubscribeWithMode("full_d30", "NSE_EQ|INE062A01020")
```

#### Dynamic Subscription Management

```go
// Change subscription mode for existing instruments
err := ws.ChangeMode("full", "NSE_EQ|INE062A01020")
if err != nil {
    log.Printf("Failed to change mode: %v", err)
}

// Unsubscribe from instruments
err = ws.Unsubscribe("NSE_EQ|INE467B01029")
if err != nil {
    log.Printf("Failed to unsubscribe: %v", err)
}
```

#### Handling Different Feed Types

```go
ws.OnLiveFeed(func(feed upstox.LiveFeedMessage) {
    for symbol, data := range feed.Feeds {
        switch {
        case data.LTPC != nil:
            // Handle LTPC data
            fmt.Printf("LTPC - Symbol: %s, LTP: %.2f\n", symbol, data.LTPC.LTP)

        case data.FullFeed != nil:
            // Handle full feed data
            if data.FullFeed.MarketFF != nil {
                market := data.FullFeed.MarketFF
                if market.LTPC != nil {
                    fmt.Printf("Full Feed - Symbol: %s, LTP: %.2f\n", symbol, market.LTPC.LTP)
                }
                if market.OptionGreeks != nil {
                    fmt.Printf("Greeks - Delta: %.4f, Gamma: %.4f\n",
                        market.OptionGreeks.Delta, market.OptionGreeks.Gamma)
                }
                fmt.Printf("Market Data - ATP: %.2f, VTT: %d, OI: %.2f\n",
                    market.ATP, market.VTT, market.OI)
            }

        case data.FirstLevelWithGreeks != nil:
            // Handle option greeks data
            greeks := data.FirstLevelWithGreeks
            if greeks.LTPC != nil && greeks.OptionGreeks != nil {
                fmt.Printf("Option - Symbol: %s, LTP: %.2f, Delta: %.4f\n",
                    symbol, greeks.LTPC.LTP, greeks.OptionGreeks.Delta)
            }
        }
    }
})
```

## API Reference

### Manager

```go
type Manager struct {
    // Contains client credentials, access token, and HTTP client
}

func NewManager(clientID, clientSecret, accessToken string) *Manager
func (m *Manager) NewWebSocket() *WebSocket

// Order Management
func (m *Manager) PlaceMarketOrder(instrumentToken string, quantity int, side string) (*OrderResponse, error)
func (m *Manager) PlaceBuyOrder(instrumentToken string, quantity int) (*OrderResponse, error)
func (m *Manager) PlaceSellOrder(instrumentToken string, quantity int) (*OrderResponse, error)

// Position Management
func (m *Manager) GetPositions() ([]Position, error)
func (m *Manager) ClosePosition(instrumentToken string) (*OrderResponse, error)
func (m *Manager) CloseAllPositions() ([]OrderResponse, error)

// Order Tracking
func (m *Manager) GetOrderBook() ([]Order, error)
func (m *Manager) GetOrderDetails(orderID string) (*Order, error)
```

### WebSocket

```go
type WebSocket struct {
    // WebSocket connection and subscription management
}

func (ws *WebSocket) Subscribe(instrumentKeys ...string) error
func (ws *WebSocket) SubscribeWithMode(mode string, instrumentKeys ...string) error
func (ws *WebSocket) Unsubscribe(instrumentKeys ...string) error
func (ws *WebSocket) ChangeMode(mode string, instrumentKeys ...string) error
func (ws *WebSocket) OnMarketInfo(callback MarketInfoCallback) error
func (ws *WebSocket) OnLiveFeed(callback LiveFeedCallback) error
func (ws *WebSocket) Connect() error
func (ws *WebSocket) Close() error
```

### Data Types

#### Subscription Modes

- `"ltpc"` - Latest Trading Price and Close Price
- `"full"` - Complete market data with depth and option greeks
- `"option_greeks"` - LTPC with option greeks
- `"full_d30"` - Full data with 30 levels (Upstox Plus required)

#### Market Data Structures

```go
type LTPCData struct {
    LTP float64 `json:"ltp"` // Last Traded Price
    LTT int64   `json:"ltt"` // Last Traded Time
    LTQ int64   `json:"ltq"` // Last Traded Quantity
    CP  float64 `json:"cp"`  // Close Price
}

type OptionGreeks struct {
    Delta float64 `json:"delta"`
    Theta float64 `json:"theta"`
    Gamma float64 `json:"gamma"`
    Vega  float64 `json:"vega"`
    Rho   float64 `json:"rho"`
}

type MarketFullFeed struct {
    LTPC         *LTPCData     `json:"ltpc,omitempty"`
    MarketLevel  []Quote       `json:"marketLevel,omitempty"`   // Market depth
    OptionGreeks *OptionGreeks `json:"optionGreeks,omitempty"`
    MarketOHLC   []OHLC        `json:"marketOHLC,omitempty"`
    ATP          float64       `json:"atp,omitempty"`          // Average Traded Price
    VTT          int64         `json:"vtt,omitempty"`          // Volume Traded Today
    OI           float64       `json:"oi,omitempty"`           // Open Interest
    IV           float64       `json:"iv,omitempty"`           // Implied Volatility
    TBQ          float64       `json:"tbq,omitempty"`          // Total Buy Quantity
    TSQ          float64       `json:"tsq,omitempty"`          // Total Sell Quantity
}

// Order and Position Types
type OrderResponse struct {
    Status   string             `json:"status"`
    Data     *OrderResponseData `json:"data,omitempty"`
    Metadata *OrderMetadata     `json:"metadata,omitempty"`
    Errors   []OrderError       `json:"errors,omitempty"`
    Summary  *OrderSummary      `json:"summary,omitempty"`
}

type Position struct {
    Exchange        string  `json:"exchange"`
    InstrumentToken string  `json:"instrument_token"`
    TradingSymbol   string  `json:"trading_symbol"`
    Quantity        int     `json:"quantity"`
    PNL             float64 `json:"pnl"`
    LastPrice       float64 `json:"last_price"`
    AveragePrice    float64 `json:"average_price"`
    Unrealised      float64 `json:"unrealised"`
    Realised        float64 `json:"realised"`
    // ... and many more fields
}

type Order struct {
    OrderID           string  `json:"order_id"`
    InstrumentToken   string  `json:"instrument_token"`
    TradingSymbol     string  `json:"trading_symbol"`
    TransactionType   string  `json:"transaction_type"`
    OrderType         string  `json:"order_type"`
    Quantity          int     `json:"quantity"`
    FilledQuantity    int     `json:"filled_quantity"`
    PendingQuantity   int     `json:"pending_quantity"`
    Status            string  `json:"status"`
    Price             float64 `json:"price"`
    AveragePrice      float64 `json:"average_price"`
    OrderTimestamp    string  `json:"order_timestamp"`
    // ... and many more fields
}
```

## Subscription Limits

### Normal Connection

- **Connections**: 2 per user
- **LTPC**: 5000 instrument keys (individual) / 2000 (combined)
- **Option Greeks**: 3000 instrument keys (individual) / 2000 (combined)
- **Full**: 2000 instrument keys (individual) / 1500 (combined)

### Upstox Plus

- **Connections**: 5 per user
- **Full D30**: 50 instrument keys (individual) / 1500 (combined)

## Error Handling

```go
// Connection errors
if err := ws.Connect(); err != nil {
    log.Printf("Connection failed: %v", err)
    // Handle reconnection logic
}

// Subscription errors
if err := ws.Subscribe("INVALID_KEY"); err != nil {
    log.Printf("Subscription failed: %v", err)
}

// Monitor connection status
ws.OnLiveFeed(func(feed upstox.LiveFeedMessage) {
    if len(feed.Feeds) == 0 {
        log.Println("No data received - connection may be lost")
    }
})
```

## Examples

See the [examples](examples/) directory for more detailed usage examples:

### WebSocket Examples
- [`websocket/basic_ltpc.go`](examples/websocket/basic_ltpc.go) - Basic LTPC data streaming
- [`websocket/full_market_data.go`](examples/websocket/full_market_data.go) - Complete market data with depth
- [`websocket/option_greeks.go`](examples/websocket/option_greeks.go) - Option chain data with greeks
- [`websocket/dynamic_subscriptions.go`](examples/websocket/dynamic_subscriptions.go) - Managing subscriptions dynamically

### Order Management Examples
- [`orders/basic_orders.go`](examples/orders/basic_orders.go) - Basic order placement (buy/sell)
- [`orders/position_management.go`](examples/orders/position_management.go) - Position tracking and closure
- [`orders/order_tracking.go`](examples/orders/order_tracking.go) - Order book and order details

## Key Features & Notes

### Order Management
- **Product Type**: Defaults to "I" (Intraday) for all market orders
- **Auto-Slicing**: Enabled by default to handle large orders
- **Market Orders**: Uses price = 0 and order_type = "MARKET"
- **Error Handling**: Comprehensive API error responses
- **Position Closure**: Automatically determines opposite side for closing

### API Endpoints Used
- **Place Order**: `POST https://api-hft.upstox.com/v3/order/place`
- **Get Positions**: `GET https://api.upstox.com/v2/portfolio/short-term-positions`
- **Exit All Positions**: `POST https://api.upstox.com/v2/order/positions/exit`
- **Get Order Book**: `GET https://api.upstox.com/v2/order/retrieve-all`
- **Get Order Details**: `GET https://api.upstox.com/v2/order/details`

## Contributing

This is a personal project, but contributions are welcome! Feel free to submit a Pull Request, open issues, or just fork it and do whatever
