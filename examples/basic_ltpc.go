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

	ws.OnMarketInfo(func(info upstox.MarketInfoMessage) {
		fmt.Printf("📊 Market Status Update - %d\n", info.CurrentTS)
		for segment, status := range info.MarketInfo.SegmentStatus {
			fmt.Printf("   %s: %s\n", segment, status)
		}
		fmt.Println()
	})

	ws.OnLiveFeed(func(feed upstox.LiveFeedMessage) {
		fmt.Printf("💹 Live Feed Update - %d\n", feed.CurrentTS)
		for symbol, data := range feed.Feeds {
			if data.LTPC != nil {
				fmt.Printf("   %s: LTP=₹%.2f, LTQ=%d, CP=₹%.2f\n",
					symbol, data.LTPC.LTP, data.LTPC.LTQ, data.LTPC.CP)
			}
		}
		fmt.Println()
	})

	if err := ws.Connect(); err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer ws.Close()

	instruments := []string{
		"NSE_EQ|INE062A01020", // ACC Limited
		"NSE_EQ|INE467B01029", // Asian Paints
		"NSE_INDEX|Nifty 50",  // Nifty 50 Index
	}

	fmt.Printf("🔗 Subscribing to LTPC data for: %v\n", instruments)
	if err := ws.Subscribe(instruments...); err != nil {
		log.Fatal("Failed to subscribe:", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("📡 Streaming LTPC data... Press Ctrl+C to exit")
	<-quit
	fmt.Println("\n👋 Shutting down...")
}