package upstox

import "time"

type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

type SubscriptionMode string

const (
	ModeLTPC         SubscriptionMode = "ltpc"
	ModeFull         SubscriptionMode = "full"
	ModeOptionGreeks SubscriptionMode = "option_greeks"
	ModeFullD30      SubscriptionMode = "full_d30"
)

type MarketStatus string

const (
	MarketStatusPreOpenStart MarketStatus = "PRE_OPEN_START"
	MarketStatusPreOpenEnd   MarketStatus = "PRE_OPEN_END"
	MarketStatusNormalOpen   MarketStatus = "NORMAL_OPEN"
	MarketStatusNormalClose  MarketStatus = "NORMAL_CLOSE"
	MarketStatusClosingStart MarketStatus = "CLOSING_START"
	MarketStatusClosingEnd   MarketStatus = "CLOSING_END"
)

type LTPCData struct {
	LTP float64 `json:"ltp"`
	LTT int64   `json:"ltt"`
	LTQ int64   `json:"ltq"`
	CP  float64 `json:"cp"`
}

type Quote struct {
	BidQ int64   `json:"bidQ"`
	BidP float64 `json:"bidP"`
	AskQ int64   `json:"askQ"`
	AskP float64 `json:"askP"`
}

type OptionGreeks struct {
	Delta float64 `json:"delta"`
	Theta float64 `json:"theta"`
	Gamma float64 `json:"gamma"`
	Vega  float64 `json:"vega"`
	Rho   float64 `json:"rho"`
}

type OHLC struct {
	Interval string  `json:"interval"`
	Open     float64 `json:"open"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"`
	Close    float64 `json:"close"`
	Volume   int64   `json:"vol"`
	TS       int64   `json:"ts"`
}

type MarketFullFeed struct {
	LTPC         *LTPCData     `json:"ltpc,omitempty"`
	MarketLevel  []Quote       `json:"marketLevel,omitempty"`
	OptionGreeks *OptionGreeks `json:"optionGreeks,omitempty"`
	MarketOHLC   []OHLC        `json:"marketOHLC,omitempty"`
	ATP          float64       `json:"atp,omitempty"`
	VTT          int64         `json:"vtt,omitempty"`
	OI           float64       `json:"oi,omitempty"`
	IV           float64       `json:"iv,omitempty"`
	TBQ          float64       `json:"tbq,omitempty"`
	TSQ          float64       `json:"tsq,omitempty"`
}

type IndexFullFeed struct {
	LTPC       *LTPCData `json:"ltpc,omitempty"`
	MarketOHLC []OHLC    `json:"marketOHLC,omitempty"`
}

type FullFeedData struct {
	MarketFF *MarketFullFeed `json:"marketFF,omitempty"`
	IndexFF  *IndexFullFeed  `json:"indexFF,omitempty"`
}

type FirstLevelWithGreeks struct {
	LTPC         *LTPCData     `json:"ltpc,omitempty"`
	FirstDepth   *Quote        `json:"firstDepth,omitempty"`
	OptionGreeks *OptionGreeks `json:"optionGreeks,omitempty"`
	VTT          int64         `json:"vtt,omitempty"`
	OI           float64       `json:"oi,omitempty"`
	IV           float64       `json:"iv,omitempty"`
}

type FeedData struct {
	LTPC                 *LTPCData             `json:"ltpc,omitempty"`
	FullFeed             *FullFeedData         `json:"fullFeed,omitempty"`
	FirstLevelWithGreeks *FirstLevelWithGreeks `json:"firstLevelWithGreeks,omitempty"`
	RequestMode          SubscriptionMode      `json:"requestMode"`
}

type MarketInfo struct {
	SegmentStatus map[string]MarketStatus `json:"segmentStatus"`
}

type MarketInfoMessage struct {
	Type       string      `json:"type"`
	CurrentTS  int64       `json:"currentTs"`
	MarketInfo *MarketInfo `json:"marketInfo"`
}

type LiveFeedMessage struct {
	Type      string               `json:"type"`
	Feeds     map[string]*FeedData `json:"feeds"`
	CurrentTS int64                `json:"currentTs"`
}

type MarketInfoCallback func(MarketInfoMessage)
type LiveFeedCallback func(LiveFeedMessage)

type SubscriptionRequest struct {
	GUID   string `json:"guid"`
	Method string `json:"method"`
	Data   struct {
		Mode           string   `json:"mode"`
		InstrumentKeys []string `json:"instrumentKeys"`
	} `json:"data"`
}

type AuthorizeResponse struct {
	Status string `json:"status"`
	Data   struct {
		AuthorizedRedirectURI string `json:"authorized_redirect_uri"`
	} `json:"data"`
}

type InstrumentSubscription struct {
	Mode SubscriptionMode
	Time time.Time
}

type ProductType string

const (
	ProductIntraday ProductType = "I"
	ProductDelivery ProductType = "D"
	ProductMTF      ProductType = "MTF"
)

type OrderType string

const (
	OrderTypeMarket OrderType = "MARKET"
	OrderTypeLimit  OrderType = "LIMIT"
	OrderTypeSL     OrderType = "SL"
	OrderTypeSLM    OrderType = "SL-M"
)

type ValidityType string

const (
	ValidityDay ValidityType = "DAY"
	ValidityIOC ValidityType = "IOC"
)

type OrderRequest struct {
	Quantity          int     `json:"quantity"`
	Product           string  `json:"product"`
	Validity          string  `json:"validity"`
	Price             float64 `json:"price"`
	Tag               string  `json:"tag,omitempty"`
	InstrumentToken   string  `json:"instrument_token"`
	OrderType         string  `json:"order_type"`
	TransactionType   string  `json:"transaction_type"`
	DisclosedQuantity int     `json:"disclosed_quantity"`
	TriggerPrice      float64 `json:"trigger_price"`
	IsAMO             bool    `json:"is_amo"`
	Slice             bool    `json:"slice"`
}

type OrderResponseData struct {
	OrderIDs []string `json:"order_ids"`
}

type OrderMetadata struct {
	Latency int `json:"latency"`
}

type OrderError struct {
	ErrorCode     string `json:"error_code"`
	Message       string `json:"message"`
	PropertyPath  string `json:"property_path"`
	InvalidValue  string `json:"invalid_value"`
	InstrumentKey string `json:"instrument_key"`
	OrderID       string `json:"order_id"`
}

type OrderSummary struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Error   int `json:"error"`
}

type OrderResponse struct {
	Status   string             `json:"status"`
	Data     *OrderResponseData `json:"data,omitempty"`
	Metadata *OrderMetadata     `json:"metadata,omitempty"`
	Errors   []OrderError       `json:"errors,omitempty"`
	Summary  *OrderSummary      `json:"summary,omitempty"`
}

type Position struct {
	Exchange              string  `json:"exchange"`
	Multiplier            float64 `json:"multiplier"`
	Value                 float64 `json:"value"`
	PNL                   float64 `json:"pnl"`
	Product               string  `json:"product"`
	InstrumentToken       string  `json:"instrument_token"`
	AveragePrice          float64 `json:"average_price"`
	BuyValue              float64 `json:"buy_value"`
	OvernightQuantity     int     `json:"overnight_quantity"`
	DayBuyValue           float64 `json:"day_buy_value"`
	DayBuyPrice           float64 `json:"day_buy_price"`
	OvernightBuyAmount    float64 `json:"overnight_buy_amount"`
	OvernightBuyQuantity  int     `json:"overnight_buy_quantity"`
	DayBuyQuantity        int     `json:"day_buy_quantity"`
	DaySellValue          float64 `json:"day_sell_value"`
	DaySellPrice          float64 `json:"day_sell_price"`
	OvernightSellAmount   float64 `json:"overnight_sell_amount"`
	OvernightSellQuantity int     `json:"overnight_sell_quantity"`
	DaySellQuantity       int     `json:"day_sell_quantity"`
	Quantity              int     `json:"quantity"`
	LastPrice             float64 `json:"last_price"`
	Unrealised            float64 `json:"unrealised"`
	Realised              float64 `json:"realised"`
	SellValue             float64 `json:"sell_value"`
	TradingSymbol         string  `json:"trading_symbol"`
	ClosePrice            float64 `json:"close_price"`
	BuyPrice              float64 `json:"buy_price"`
	SellPrice             float64 `json:"sell_price"`
}

type Order struct {
	Exchange          string  `json:"exchange"`
	Product           string  `json:"product"`
	Price             float64 `json:"price"`
	Quantity          int     `json:"quantity"`
	Status            string  `json:"status"`
	GUID              string  `json:"guid"`
	Tag               string  `json:"tag"`
	InstrumentToken   string  `json:"instrument_token"`
	PlacedBy          string  `json:"placed_by"`
	TradingSymbol     string  `json:"trading_symbol"`
	OrderType         string  `json:"order_type"`
	Validity          string  `json:"validity"`
	TriggerPrice      float64 `json:"trigger_price"`
	DisclosedQuantity int     `json:"disclosed_quantity"`
	TransactionType   string  `json:"transaction_type"`
	AveragePrice      float64 `json:"average_price"`
	FilledQuantity    int     `json:"filled_quantity"`
	PendingQuantity   int     `json:"pending_quantity"`
	StatusMessage     string  `json:"status_message"`
	StatusMessageRaw  string  `json:"status_message_raw"`
	ExchangeOrderID   string  `json:"exchange_order_id"`
	ParentOrderID     string  `json:"parent_order_id"`
	OrderID           string  `json:"order_id"`
	Variety           string  `json:"variety"`
	OrderTimestamp    string  `json:"order_timestamp"`
	ExchangeTimestamp string  `json:"exchange_timestamp"`
	IsAMO             bool    `json:"is_amo"`
	OrderRequestID    string  `json:"order_request_id"`
	OrderRefID        string  `json:"order_ref_id"`
}

type PositionResponse struct {
	Status string     `json:"status"`
	Data   []Position `json:"data"`
}

type OrderBookResponse struct {
	Status string  `json:"status"`
	Data   []Order `json:"data"`
}

type OrderDetailResponse struct {
	Status string `json:"status"`
	Data   Order  `json:"data"`
}

type MarginData struct {
	UsedMargin      float64 `json:"used_margin"`
	PayinAmount     float64 `json:"payin_amount"`
	SpanMargin      float64 `json:"span_margin"`
	AdhocMargin     float64 `json:"adhoc_margin"`
	NotionalCash    float64 `json:"notional_cash"`
	AvailableMargin float64 `json:"available_margin"`
	ExposureMargin  float64 `json:"exposure_margin"`
}

type FundsData struct {
	Equity    MarginData `json:"equity"`
	Commodity MarginData `json:"commodity"`
}

type FundsResponse struct {
	Status string    `json:"status"`
	Data   FundsData `json:"data"`
}
