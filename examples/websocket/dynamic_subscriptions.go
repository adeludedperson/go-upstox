package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	ws.OnMarketInfo(func(info upstox.MarketInfoMessage) {
		fmt.Printf("📊 Market Status: %d instruments active\n", len(info.MarketInfo.SegmentStatus))
	})

	ws.OnLiveFeed(func(feed upstox.LiveFeedMessage) {
		fmt.Printf("📈 Feed Update [%d] - %d instruments\n", feed.CurrentTS, len(feed.Feeds))

		for symbol, data := range feed.Feeds {
			var price float64
			var mode string

			switch {
			case data.LTPC != nil:
				price = data.LTPC.LTP
				mode = "LTPC"
			case data.FullFeed != nil && data.FullFeed.MarketFF != nil && data.FullFeed.MarketFF.LTPC != nil:
				price = data.FullFeed.MarketFF.LTPC.LTP
				mode = "FULL"
			case data.FirstLevelWithGreeks != nil && data.FirstLevelWithGreeks.LTPC != nil:
				price = data.FirstLevelWithGreeks.LTPC.LTP
				mode = "GREEKS"
			}

			if price > 0 {
				fmt.Printf("   %s [%s]: ₹%.2f\n", symbol, mode, price)
			}
		}
		fmt.Println()
	})

	if err := ws.Connect(); err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer ws.Close()

	// Initial subscriptions
	initialInstruments := []string{
		"NSE_EQ|INE062A01020", // ACC Limited
		"NSE_EQ|INE467B01029", // Asian Paints
	}

	fmt.Printf("🔗 Initial LTPC subscription: %v\n", initialInstruments)
	if err := ws.Subscribe(initialInstruments...); err != nil {
		log.Fatal("Failed to subscribe:", err)
	}

	// Demonstrate dynamic subscription management
	go func() {
		time.Sleep(5 * time.Second)

		// Add more instruments
		fmt.Println("➕ Adding more instruments...")
		newInstruments := []string{
			"NSE_INDEX|Nifty 50",
			"NSE_INDEX|Nifty Bank",
		}
		if err := ws.Subscribe(newInstruments...); err != nil {
			log.Printf("Failed to add instruments: %v", err)
		}

		time.Sleep(5 * time.Second)

		// Change mode for existing instruments
		fmt.Println("🔄 Changing to FULL mode for equity instruments...")
		if err := ws.ChangeMode("full", initialInstruments...); err != nil {
			log.Printf("Failed to change mode: %v", err)
		}

		time.Sleep(5 * time.Second)

		// Unsubscribe from some instruments
		fmt.Println("➖ Unsubscribing from one instrument...")
		if err := ws.Unsubscribe("NSE_EQ|INE467B01029"); err != nil {
			log.Printf("Failed to unsubscribe: %v", err)
		}

		time.Sleep(5 * time.Second)

		// Subscribe to new instrument with specific mode
		fmt.Println("🎯 Adding new instrument with FULL mode...")
		if err := ws.SubscribeWithMode("full", "NSE_EQ|INE009A01021"); err != nil { // Infosys
			log.Printf("Failed to subscribe with mode: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("📡 Streaming data with dynamic subscription management...")
	fmt.Println("🔄 Watch for automatic subscription changes every 5 seconds")
	fmt.Println("⏹️  Press Ctrl+C to exit")

	<-quit
	fmt.Println("\n👋 Shutting down...")
}
