package upstox

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	pb "github.com/adeludedperson/go-upstox/pb"
)

const (
	authorizeURL = "https://api.upstox.com/v3/feed/market-data-feed/authorize"
	baseWsURL    = "wss://api.upstox.com/v3/feed/market-data-feed"
)

type WebSocket struct {
	manager           *Manager
	conn              *websocket.Conn
	subscriptions     map[string]InstrumentSubscription
	marketInfoHandler MarketInfoCallback
	liveFeedHandler   LiveFeedCallback
	done              chan struct{}
	mu                sync.RWMutex
	connected         bool
}

func (ws *WebSocket) Subscribe(instrumentKeys ...string) error {
	return ws.SubscribeWithMode(string(ModeLTPC), instrumentKeys...)
}

func (ws *WebSocket) SubscribeWithMode(mode string, instrumentKeys ...string) error {
	if len(instrumentKeys) == 0 {
		return fmt.Errorf("no instrument keys provided")
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	subscriptionMode := SubscriptionMode(mode)
	for _, key := range instrumentKeys {
		ws.subscriptions[key] = InstrumentSubscription{
			Mode: subscriptionMode,
			Time: time.Now(),
		}
	}

	if ws.connected && ws.conn != nil {
		return ws.sendSubscriptionRequest("sub", mode, instrumentKeys)
	}

	return nil
}

func (ws *WebSocket) Unsubscribe(instrumentKeys ...string) error {
	if len(instrumentKeys) == 0 {
		return fmt.Errorf("no instrument keys provided")
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	var mode string
	for _, key := range instrumentKeys {
		if sub, exists := ws.subscriptions[key]; exists {
			mode = string(sub.Mode)
			delete(ws.subscriptions, key)
		}
	}

	if ws.connected && ws.conn != nil && mode != "" {
		return ws.sendSubscriptionRequest("unsub", mode, instrumentKeys)
	}

	return nil
}

func (ws *WebSocket) ChangeMode(mode string, instrumentKeys ...string) error {
	if len(instrumentKeys) == 0 {
		return fmt.Errorf("no instrument keys provided")
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	subscriptionMode := SubscriptionMode(mode)
	for _, key := range instrumentKeys {
		if _, exists := ws.subscriptions[key]; exists {
			ws.subscriptions[key] = InstrumentSubscription{
				Mode: subscriptionMode,
				Time: time.Now(),
			}
		}
	}

	if ws.connected && ws.conn != nil {
		return ws.sendSubscriptionRequest("change_mode", mode, instrumentKeys)
	}

	return nil
}

func (ws *WebSocket) OnMarketInfo(callback MarketInfoCallback) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.marketInfoHandler = callback
	return nil
}

func (ws *WebSocket) OnLiveFeed(callback LiveFeedCallback) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.liveFeedHandler = callback
	return nil
}

func (ws *WebSocket) Connect() error {
	wsURL, err := ws.getAuthorizedWebSocketURL()
	if err != nil {
		return fmt.Errorf("failed to get authorized WebSocket URL: %w", err)
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+ws.manager.GetAccessToken())

	conn, _, err := dialer.Dial(wsURL, headers)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	ws.mu.Lock()
	ws.conn = conn
	ws.connected = true
	ws.mu.Unlock()

	go ws.readMessages()

	return ws.subscribeToExistingInstruments()
}

func (ws *WebSocket) Close() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.connected = false

	close(ws.done)

	if ws.conn != nil {
		return ws.conn.Close()
	}

	return nil
}

func (ws *WebSocket) getAuthorizedWebSocketURL() (string, error) {
	req, err := http.NewRequest("GET", authorizeURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+ws.manager.GetAccessToken())
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

func (ws *WebSocket) sendSubscriptionRequest(method, mode string, instrumentKeys []string) error {
	request := SubscriptionRequest{
		GUID:   uuid.New().String(),
		Method: method,
	}
	request.Data.Mode = mode
	request.Data.InstrumentKeys = instrumentKeys

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription request: %w", err)
	}

	return ws.conn.WriteMessage(websocket.BinaryMessage, requestBytes)
}

func (ws *WebSocket) subscribeToExistingInstruments() error {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	modeGroups := make(map[SubscriptionMode][]string)
	for instrument, subscription := range ws.subscriptions {
		modeGroups[subscription.Mode] = append(modeGroups[subscription.Mode], instrument)
	}

	for mode, instruments := range modeGroups {
		if err := ws.sendSubscriptionRequest("sub", string(mode), instruments); err != nil {
			return fmt.Errorf("failed to subscribe to existing instruments: %w", err)
		}
	}

	return nil
}

func (ws *WebSocket) readMessages() {
	defer func() {
		ws.mu.Lock()
		ws.connected = false
		ws.mu.Unlock()
	}()

	for {
		select {
		case <-ws.done:
			return
		default:
			_, messageData, err := ws.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Printf("WebSocket error: %v\n", err)
				}
				return
			}

			ws.processMessage(messageData)
		}
	}
}

func (ws *WebSocket) processMessage(data []byte) {
	var feedResponse pb.FeedResponse
	if err := proto.Unmarshal(data, &feedResponse); err != nil {
		fmt.Printf("Failed to unmarshal protobuf message: %v\n", err)
		return
	}

	switch feedResponse.Type {
	case pb.Type_market_info:
		if ws.marketInfoHandler != nil {
			marketInfo := ws.convertMarketInfo(&feedResponse)
			ws.marketInfoHandler(marketInfo)
		}

	case pb.Type_live_feed:
		if ws.liveFeedHandler != nil {
			liveFeed := ws.convertLiveFeed(&feedResponse)
			ws.liveFeedHandler(liveFeed)
		}
	}
}

func (ws *WebSocket) convertMarketInfo(feedResponse *pb.FeedResponse) MarketInfoMessage {
	message := MarketInfoMessage{
		Type:      "market_info",
		CurrentTS: feedResponse.CurrentTs,
	}

	if feedResponse.MarketInfo != nil {
		message.MarketInfo = &MarketInfo{
			SegmentStatus: make(map[string]MarketStatus),
		}

		for segment, status := range feedResponse.MarketInfo.SegmentStatus {
			message.MarketInfo.SegmentStatus[segment] = ws.convertMarketStatus(status)
		}
	}

	return message
}

func (ws *WebSocket) convertMarketStatus(status pb.MarketStatus) MarketStatus {
	switch status {
	case pb.MarketStatus_PRE_OPEN_START:
		return MarketStatusPreOpenStart
	case pb.MarketStatus_PRE_OPEN_END:
		return MarketStatusPreOpenEnd
	case pb.MarketStatus_NORMAL_OPEN:
		return MarketStatusNormalOpen
	case pb.MarketStatus_NORMAL_CLOSE:
		return MarketStatusNormalClose
	case pb.MarketStatus_CLOSING_START:
		return MarketStatusClosingStart
	case pb.MarketStatus_CLOSING_END:
		return MarketStatusClosingEnd
	default:
		return MarketStatusNormalOpen
	}
}

func (ws *WebSocket) convertLiveFeed(feedResponse *pb.FeedResponse) LiveFeedMessage {
	message := LiveFeedMessage{
		Type:      "live_feed",
		CurrentTS: feedResponse.CurrentTs,
		Feeds:     make(map[string]*FeedData),
	}

	for instrument, pbFeed := range feedResponse.Feeds {
		feedData := &FeedData{}

		switch pbFeed.FeedUnion.(type) {
		case *pb.Feed_Ltpc:
			ltpc := pbFeed.GetLtpc()
			feedData.LTPC = &LTPCData{
				LTP: ltpc.Ltp,
				LTT: ltpc.Ltt,
				LTQ: ltpc.Ltq,
				CP:  ltpc.Cp,
			}

		case *pb.Feed_FullFeed:
			fullFeed := pbFeed.GetFullFeed()
			feedData.FullFeed = ws.convertFullFeed(fullFeed)

		case *pb.Feed_FirstLevelWithGreeks:
			greeks := pbFeed.GetFirstLevelWithGreeks()
			feedData.FirstLevelWithGreeks = ws.convertFirstLevelWithGreeks(greeks)
		}

		feedData.RequestMode = ws.convertRequestMode(pbFeed.RequestMode)
		message.Feeds[instrument] = feedData
	}

	return message
}

func (ws *WebSocket) convertFullFeed(pbFullFeed *pb.FullFeed) *FullFeedData {
	fullFeed := &FullFeedData{}

	switch pbFullFeed.FullFeedUnion.(type) {
	case *pb.FullFeed_MarketFF:
		marketFF := pbFullFeed.GetMarketFF()
		fullFeed.MarketFF = &MarketFullFeed{}

		if marketFF.Ltpc != nil {
			fullFeed.MarketFF.LTPC = &LTPCData{
				LTP: marketFF.Ltpc.Ltp,
				LTT: marketFF.Ltpc.Ltt,
				LTQ: marketFF.Ltpc.Ltq,
				CP:  marketFF.Ltpc.Cp,
			}
		}

		if marketFF.MarketLevel != nil {
			for _, quote := range marketFF.MarketLevel.BidAskQuote {
				fullFeed.MarketFF.MarketLevel = append(fullFeed.MarketFF.MarketLevel, Quote{
					BidQ: quote.BidQ,
					BidP: quote.BidP,
					AskQ: quote.AskQ,
					AskP: quote.AskP,
				})
			}
		}

		if marketFF.OptionGreeks != nil {
			fullFeed.MarketFF.OptionGreeks = &OptionGreeks{
				Delta: marketFF.OptionGreeks.Delta,
				Theta: marketFF.OptionGreeks.Theta,
				Gamma: marketFF.OptionGreeks.Gamma,
				Vega:  marketFF.OptionGreeks.Vega,
				Rho:   marketFF.OptionGreeks.Rho,
			}
		}

		if marketFF.MarketOHLC != nil {
			for _, ohlc := range marketFF.MarketOHLC.Ohlc {
				fullFeed.MarketFF.MarketOHLC = append(fullFeed.MarketFF.MarketOHLC, OHLC{
					Interval: ohlc.Interval,
					Open:     ohlc.Open,
					High:     ohlc.High,
					Low:      ohlc.Low,
					Close:    ohlc.Close,
					Volume:   ohlc.Vol,
					TS:       ohlc.Ts,
				})
			}
		}

		fullFeed.MarketFF.ATP = marketFF.Atp
		fullFeed.MarketFF.VTT = marketFF.Vtt
		fullFeed.MarketFF.OI = marketFF.Oi
		fullFeed.MarketFF.IV = marketFF.Iv
		fullFeed.MarketFF.TBQ = marketFF.Tbq
		fullFeed.MarketFF.TSQ = marketFF.Tsq

	case *pb.FullFeed_IndexFF:
		indexFF := pbFullFeed.GetIndexFF()
		fullFeed.IndexFF = &IndexFullFeed{}

		if indexFF.Ltpc != nil {
			fullFeed.IndexFF.LTPC = &LTPCData{
				LTP: indexFF.Ltpc.Ltp,
				LTT: indexFF.Ltpc.Ltt,
				LTQ: indexFF.Ltpc.Ltq,
				CP:  indexFF.Ltpc.Cp,
			}
		}

		if indexFF.MarketOHLC != nil {
			for _, ohlc := range indexFF.MarketOHLC.Ohlc {
				fullFeed.IndexFF.MarketOHLC = append(fullFeed.IndexFF.MarketOHLC, OHLC{
					Interval: ohlc.Interval,
					Open:     ohlc.Open,
					High:     ohlc.High,
					Low:      ohlc.Low,
					Close:    ohlc.Close,
					Volume:   ohlc.Vol,
					TS:       ohlc.Ts,
				})
			}
		}
	}

	return fullFeed
}

func (ws *WebSocket) convertFirstLevelWithGreeks(pbGreeks *pb.FirstLevelWithGreeks) *FirstLevelWithGreeks {
	greeks := &FirstLevelWithGreeks{
		VTT: pbGreeks.Vtt,
		OI:  pbGreeks.Oi,
		IV:  pbGreeks.Iv,
	}

	if pbGreeks.Ltpc != nil {
		greeks.LTPC = &LTPCData{
			LTP: pbGreeks.Ltpc.Ltp,
			LTT: pbGreeks.Ltpc.Ltt,
			LTQ: pbGreeks.Ltpc.Ltq,
			CP:  pbGreeks.Ltpc.Cp,
		}
	}

	if pbGreeks.FirstDepth != nil {
		greeks.FirstDepth = &Quote{
			BidQ: pbGreeks.FirstDepth.BidQ,
			BidP: pbGreeks.FirstDepth.BidP,
			AskQ: pbGreeks.FirstDepth.AskQ,
			AskP: pbGreeks.FirstDepth.AskP,
		}
	}

	if pbGreeks.OptionGreeks != nil {
		greeks.OptionGreeks = &OptionGreeks{
			Delta: pbGreeks.OptionGreeks.Delta,
			Theta: pbGreeks.OptionGreeks.Theta,
			Gamma: pbGreeks.OptionGreeks.Gamma,
			Vega:  pbGreeks.OptionGreeks.Vega,
			Rho:   pbGreeks.OptionGreeks.Rho,
		}
	}

	return greeks
}

func (ws *WebSocket) convertRequestMode(mode pb.RequestMode) SubscriptionMode {
	switch mode {
	case pb.RequestMode_ltpc:
		return ModeLTPC
	case pb.RequestMode_full_d5:
		return ModeFull
	case pb.RequestMode_option_greeks:
		return ModeOptionGreeks
	case pb.RequestMode_full_d30:
		return ModeFullD30
	default:
		return ModeLTPC
	}
}