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
	LTPC          *LTPCData      `json:"ltpc,omitempty"`
	MarketLevel   []Quote        `json:"marketLevel,omitempty"`
	OptionGreeks  *OptionGreeks  `json:"optionGreeks,omitempty"`
	MarketOHLC    []OHLC         `json:"marketOHLC,omitempty"`
	ATP           float64        `json:"atp,omitempty"`
	VTT           int64          `json:"vtt,omitempty"`
	OI            float64        `json:"oi,omitempty"`
	IV            float64        `json:"iv,omitempty"`
	TBQ           float64        `json:"tbq,omitempty"`
	TSQ           float64        `json:"tsq,omitempty"`
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
	Type        string      `json:"type"`
	CurrentTS   int64       `json:"currentTs"`
	MarketInfo  *MarketInfo `json:"marketInfo"`
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