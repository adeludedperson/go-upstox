package upstox

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	pb "github.com/adeludedperson/go-upstox/pb"
)

type WebSocketManager struct {
	ws                   *websocket.Conn
	url                  string
	config               WebSocketConfig
	onPriceUpdate        func(symbol string, price float64, ltq *int32)
	reconnectAttempts    int
	maxReconnectAttempts int
	reconnectDelay       time.Duration
	isConnecting         bool
	shouldReconnect      bool
	mu                   sync.RWMutex
	ctx                  context.Context
	cancel               context.CancelFunc
}

type WebSocketConfig struct {
	InstrumentKeys []string
	Token          string
}

type SubscriptionMessage struct {
	GUID   string                  `json:"guid"`
	Method string                  `json:"method"`
	Data   SubscriptionMessageData `json:"data"`
}

type SubscriptionMessageData struct {
	Mode           string   `json:"mode"`
	InstrumentKeys []string `json:"instrumentKeys"`
}

func NewWebSocketManager(url string, config WebSocketConfig, onPriceUpdate func(string, float64, *int32)) *WebSocketManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &WebSocketManager{
		url:                  url,
		config:               config,
		onPriceUpdate:        onPriceUpdate,
		maxReconnectAttempts: 3,
		reconnectDelay:       time.Second,
		shouldReconnect:      true,
		ctx:                  ctx,
		cancel:               cancel,
	}
}

func (wsm *WebSocketManager) connect() error {
	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	if wsm.isConnecting || wsm.ws != nil {
		return nil
	}

	wsm.isConnecting = true

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, resp, err := dialer.Dial(wsm.url, nil)
	if err != nil {
		wsm.isConnecting = false
		if resp != nil {
			log.Printf("WebSocket handshake failed with status: %s", resp.Status)
		}
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	wsm.ws = conn
	wsm.reconnectAttempts = 0
	wsm.reconnectDelay = time.Second
	wsm.isConnecting = false

	go wsm.handleMessages()

	// Only subscribe if we have instrument keys
	if len(wsm.config.InstrumentKeys) > 0 {
		return wsm.subscribe()
	}

	return nil
}

func (wsm *WebSocketManager) subscribe() error {
	guid, err := generateGUID()
	if err != nil {
		return fmt.Errorf("failed to generate GUID: %w", err)
	}

	subscribeMsg := SubscriptionMessage{
		GUID:   guid,
		Method: "sub",
		Data: SubscriptionMessageData{
			Mode:           "ltpc",
			InstrumentKeys: wsm.config.InstrumentKeys,
		},
	}

	msgBytes, err := json.Marshal(subscribeMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription message: %w", err)
	}

	// Per Upstox V3 docs: "The WebSocket request message should be sent in binary format"
	return wsm.ws.WriteMessage(websocket.BinaryMessage, msgBytes)
}

func (wsm *WebSocketManager) handleMessages() {
	defer func() {
		wsm.mu.Lock()
		wsm.ws = nil
		wsm.mu.Unlock()
	}()

	for {
		select {
		case <-wsm.ctx.Done():
			return
		default:
			messageType, data, err := wsm.ws.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				wsm.handleDisconnect()
				return
			}

			if messageType == websocket.BinaryMessage {
				wsm.processMessage(data)
			} else if messageType == websocket.TextMessage {
				log.Printf("Unexpected text message: %s", string(data))
			}
		}
	}
}

func (wsm *WebSocketManager) processMessage(data []byte) {

	var feedResponse pb.FeedResponse
	if err := proto.Unmarshal(data, &feedResponse); err != nil {
		log.Printf("Failed to unmarshal protobuf message: %v", err)
		return
	}

	// log.Printf("Processed feed response with %d symbols", len(feedResponse.Feeds))
	// log.Printf("Feed Response: %+v", feedResponse)

	if feedResponse.Type != pb.Type_live_feed && feedResponse.Type != pb.Type_initial_feed {
		return
	}

	for symbol, feed := range feedResponse.Feeds {
		var ltp float64
		var ltq *int32

		switch feedUnion := feed.FeedUnion.(type) {
		case *pb.Feed_Ltpc:
			ltp = float64(feedUnion.Ltpc.Ltp)
			if feedUnion.Ltpc.Ltq != 0 {
				ltqVal := int32(feedUnion.Ltpc.Ltq)
				ltq = &ltqVal
			}

		case *pb.Feed_FullFeed:
			fullFeed := feedUnion.FullFeed
			switch fullFeedUnion := fullFeed.FullFeedUnion.(type) {
			case *pb.FullFeed_MarketFF:
				if fullFeedUnion.MarketFF.Ltpc != nil {
					ltp = float64(fullFeedUnion.MarketFF.Ltpc.Ltp)
					if fullFeedUnion.MarketFF.Ltpc.Ltq != 0 {
						ltqVal := int32(fullFeedUnion.MarketFF.Ltpc.Ltq)
						ltq = &ltqVal
					}
				}
			case *pb.FullFeed_IndexFF:
				if fullFeedUnion.IndexFF.Ltpc != nil {
					ltp = float64(fullFeedUnion.IndexFF.Ltpc.Ltp)
					if fullFeedUnion.IndexFF.Ltpc.Ltq != 0 {
						ltqVal := int32(fullFeedUnion.IndexFF.Ltpc.Ltq)
						ltq = &ltqVal
					}
				}
			}

		case *pb.Feed_FirstLevelWithGreeks:
			if feedUnion.FirstLevelWithGreeks.Ltpc != nil {
				ltp = float64(feedUnion.FirstLevelWithGreeks.Ltpc.Ltp)
				if feedUnion.FirstLevelWithGreeks.Ltpc.Ltq != 0 {
					ltqVal := int32(feedUnion.FirstLevelWithGreeks.Ltpc.Ltq)
					ltq = &ltqVal
				}
			}
		}

		if ltp > 0 && wsm.onPriceUpdate != nil {
			wsm.onPriceUpdate(symbol, ltp, ltq)
		}
	}
}

func (wsm *WebSocketManager) handleDisconnect() {
	if !wsm.shouldReconnect {
		return
	}

	if wsm.reconnectAttempts < wsm.maxReconnectAttempts {
		wsm.reconnectAttempts++
		wsm.reconnectDelay *= 2

		log.Printf("Reconnecting attempt %d in %v", wsm.reconnectAttempts, wsm.reconnectDelay)

		time.AfterFunc(wsm.reconnectDelay, func() {
			if err := wsm.connect(); err != nil {
				log.Printf("Reconnection failed: %v", err)
			}
		})
	} else {
		log.Printf("Max reconnection attempts reached")
		wsm.Stop()
	}
}

func (wsm *WebSocketManager) Start() error {
	wsm.shouldReconnect = true
	return wsm.connect()
}

func (wsm *WebSocketManager) Stop() {
	wsm.shouldReconnect = false
	wsm.cancel()

	wsm.mu.Lock()
	defer wsm.mu.Unlock()

	if wsm.ws != nil {
		wsm.ws.Close()
		wsm.ws = nil
	}
}

func generateGUID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16]), nil
}

func (wsm *WebSocketManager) UpdateInstruments(instrumentKeys []string) error {
	wsm.mu.Lock()
	wsm.config.InstrumentKeys = instrumentKeys
	wsm.mu.Unlock()

	if wsm.ws != nil {
		return wsm.subscribe()
	}
	return nil
}
