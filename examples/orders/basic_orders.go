package main

import (
	"fmt"
	"log"

	"github.com/adeludedperson/go-upstox"
)

func main() {
	// Initialize the manager with your credentials
	clientID := "your_client_id"
	clientSecret := "your_client_secret"
	accessToken := "your_access_token"

	manager := upstox.NewManager(clientID, clientSecret, accessToken)

	// Example instrument token for SBIN (State Bank of India)
	instrumentToken := "NSE_EQ|INE062A01020"

	fmt.Println("=== Basic Order Placement Example ===")

	// Place a buy order for 1 share
	fmt.Printf("Placing buy order for 1 share of %s...\n", instrumentToken)
	buyResp, err := manager.PlaceBuyOrder(instrumentToken, 1)
	if err != nil {
		log.Fatalf("Failed to place buy order: %v", err)
	}

	if len(buyResp.Data.OrderIDs) > 0 {
		fmt.Printf("✅ Buy order placed successfully! Order ID: %s\n", buyResp.Data.OrderIDs[0])
	} else {
		fmt.Println("❌ Buy order failed")
		if len(buyResp.Errors) > 0 {
			fmt.Printf("Error: %s\n", buyResp.Errors[0].Message)
		}
	}

	// Place a sell order for 1 share
	fmt.Printf("\nPlacing sell order for 1 share of %s...\n", instrumentToken)
	sellResp, err := manager.PlaceSellOrder(instrumentToken, 1)
	if err != nil {
		log.Fatalf("Failed to place sell order: %v", err)
	}

	if len(sellResp.Data.OrderIDs) > 0 {
		fmt.Printf("✅ Sell order placed successfully! Order ID: %s\n", sellResp.Data.OrderIDs[0])
	} else {
		fmt.Println("❌ Sell order failed")
		if len(sellResp.Errors) > 0 {
			fmt.Printf("Error: %s\n", sellResp.Errors[0].Message)
		}
	}

	// Alternative: Use PlaceMarketOrder directly
	fmt.Printf("\nPlacing market order using PlaceMarketOrder method...\n")
	marketResp, err := manager.PlaceMarketOrder(instrumentToken, 2, "BUY")
	if err != nil {
		log.Fatalf("Failed to place market order: %v", err)
	}

	if len(marketResp.Data.OrderIDs) > 0 {
		fmt.Printf("✅ Market order placed successfully! Order ID: %s\n", marketResp.Data.OrderIDs[0])
		fmt.Printf("API Latency: %d ms\n", marketResp.Metadata.Latency)
	}
}