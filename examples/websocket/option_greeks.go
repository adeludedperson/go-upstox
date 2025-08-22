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
		fmt.Printf("📊 Option Greeks Update - %d\n", feed.CurrentTS)

		for symbol, data := range feed.Feeds {
			fmt.Printf("🎯 Option: %s\n", symbol)

			if data.FirstLevelWithGreeks != nil {
				greeks := data.FirstLevelWithGreeks

				// LTPC Data
				if greeks.LTPC != nil {
					fmt.Printf("   💰 Price: LTP=₹%.2f, LTQ=%d, CP=₹%.2f\n",
						greeks.LTPC.LTP, greeks.LTPC.LTQ, greeks.LTPC.CP)
				}

				// First Level Market Depth
				if greeks.FirstDepth != nil {
					fmt.Printf("   📊 Depth: Bid ₹%.2f(%d) | Ask ₹%.2f(%d)\n",
						greeks.FirstDepth.BidP, greeks.FirstDepth.BidQ,
						greeks.FirstDepth.AskP, greeks.FirstDepth.AskQ)
				}

				// Option Statistics
				fmt.Printf("   📈 Stats: VTT=%d, OI=%.0f\n", greeks.VTT, greeks.OI)

				if greeks.IV > 0 {
					fmt.Printf("   📉 Implied Volatility: %.2f%%\n", greeks.IV*100)
				}

				// Option Greeks
				if greeks.OptionGreeks != nil {
					fmt.Printf("   🔢 Greeks:\n")
					fmt.Printf("      Delta (Δ): %+.4f (Price sensitivity)\n", greeks.OptionGreeks.Delta)
					fmt.Printf("      Gamma (Γ): %+.4f (Delta sensitivity)\n", greeks.OptionGreeks.Gamma)
					fmt.Printf("      Theta (Θ): %+.4f (Time decay)\n", greeks.OptionGreeks.Theta)
					fmt.Printf("      Vega  (ν): %+.4f (Volatility sensitivity)\n", greeks.OptionGreeks.Vega)
					fmt.Printf("      Rho   (ρ): %+.4f (Interest rate sensitivity)\n", greeks.OptionGreeks.Rho)

					// Risk interpretation
					if greeks.OptionGreeks.Delta > 0.5 {
						fmt.Printf("   ✅ Deep ITM Call / OTM Put\n")
					} else if greeks.OptionGreeks.Delta > 0 {
						fmt.Printf("   🟡 OTM Call / ITM Put\n")
					} else if greeks.OptionGreeks.Delta > -0.5 {
						fmt.Printf("   🟡 ITM Call / OTM Put\n")
					} else {
						fmt.Printf("   ✅ Deep ITM Put / OTM Call\n")
					}

					if greeks.OptionGreeks.Theta < -0.1 {
						fmt.Printf("   ⚠️  High time decay\n")
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

	// Example option instruments (replace with actual option instrument keys)
	instruments := []string{
		"NSE_FO|45450", // Example option contract
		"NSE_FO|45451", // Example option contract
	}

	fmt.Printf("🔗 Subscribing to Option Greeks for: %v\n", instruments)
	if err := ws.SubscribeWithMode("option_greeks", instruments...); err != nil {
		log.Fatal("Failed to subscribe:", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("📡 Streaming option greeks... Press Ctrl+C to exit")
	fmt.Println("💡 Note: Replace instrument keys with actual option contracts")
	<-quit
	fmt.Println("\n👋 Shutting down...")
}
