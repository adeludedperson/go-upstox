package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/adeludedperson/go-upstox"
)

func main() {
	clientID := os.Getenv("UPSTOX_CLIENT_ID")
	clientSecret := os.Getenv("UPSTOX_CLIENT_SECRET")
	accessToken := os.Getenv("UPSTOX_ACCESS_TOKEN")

	if clientID == "" || clientSecret == "" || accessToken == "" {
		log.Fatal("Please set UPSTOX_CLIENT_ID, UPSTOX_CLIENT_SECRET, and UPSTOX_ACCESS_TOKEN environment variables")
	}

	manager := upstox.NewManager(clientID, clientSecret, accessToken)
	ws := manager.NewWebSocket()

	ws.OnLiveFeed(func(feed upstox.LiveFeedMessage) {
		fmt.Printf("📈 Full Market Data Update - %d\n", feed.CurrentTS)
		
		for symbol, data := range feed.Feeds {
			fmt.Printf("🏢 Symbol: %s\n", symbol)
			
			if data.FullFeed != nil && data.FullFeed.MarketFF != nil {
				market := data.FullFeed.MarketFF
				
				// LTPC Data
				if market.LTPC != nil {
					fmt.Printf("   💰 LTPC: LTP=₹%.2f, LTQ=%d, CP=₹%.2f\n",
						market.LTPC.LTP, market.LTPC.LTQ, market.LTPC.CP)
				}
				
				// Market Statistics
				fmt.Printf("   📊 Stats: ATP=₹%.2f, VTT=%d, OI=%.0f\n",
					market.ATP, market.VTT, market.OI)
				
				if market.IV > 0 {
					fmt.Printf("   📉 Volatility: IV=%.2f%%\n", market.IV*100)
				}
				
				fmt.Printf("   🟢 Buy Qty: %.0f, 🔴 Sell Qty: %.0f\n",
					market.TBQ, market.TSQ)
				
				// Market Depth (first 5 levels)
				if len(market.MarketLevel) > 0 {
					fmt.Println("   📊 Market Depth:")
					for i, quote := range market.MarketLevel {
						if i >= 5 { // Show only first 5 levels
							break
						}
						fmt.Printf("      L%d: Bid ₹%.2f(%d) | Ask ₹%.2f(%d)\n",
							i+1, quote.BidP, quote.BidQ, quote.AskP, quote.AskQ)
					}
				}
				
				// Option Greeks (if available)
				if market.OptionGreeks != nil {
					fmt.Printf("   🔢 Greeks: Δ=%.4f, Γ=%.4f, Θ=%.4f, ν=%.4f, ρ=%.4f\n",
						market.OptionGreeks.Delta, market.OptionGreeks.Gamma,
						market.OptionGreeks.Theta, market.OptionGreeks.Vega,
						market.OptionGreeks.Rho)
				}
				
				// OHLC Data (if available)
				if len(market.MarketOHLC) > 0 {
					for _, ohlc := range market.MarketOHLC {
						fmt.Printf("   📊 OHLC [%s]: O=₹%.2f, H=₹%.2f, L=₹%.2f, C=₹%.2f, Vol=%d\n",
							ohlc.Interval, ohlc.Open, ohlc.High, ohlc.Low, ohlc.Close, ohlc.Volume)
					}
				}
			}
			
			fmt.Println()
		}
	})

	if err := ws.Connect(); err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer ws.Close()

	instruments := []string{
		"NSE_EQ|INE062A01020", // ACC Limited - Full market data
		"NSE_EQ|INE467B01029", // Asian Paints - Full market data
	}

	fmt.Printf("🔗 Subscribing to FULL market data for: %v\n", instruments)
	if err := ws.SubscribeWithMode("full", instruments...); err != nil {
		log.Fatal("Failed to subscribe:", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("📡 Streaming full market data... Press Ctrl+C to exit")
	<-quit
	fmt.Println("\n👋 Shutting down...")
}