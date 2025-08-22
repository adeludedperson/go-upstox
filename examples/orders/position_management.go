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

	fmt.Println("=== Position Management Example ===")

	// Get current positions
	fmt.Println("\n1. Fetching current positions...")
	positions, err := manager.GetPositions()
	if err != nil {
		log.Fatalf("Failed to get positions: %v", err)
	}

	if len(positions) == 0 {
		fmt.Println("No positions found.")
	} else {
		fmt.Printf("Found %d positions:\n", len(positions))
		for i, pos := range positions {
			fmt.Printf("  %d. %s (%s)\n", i+1, pos.TradingSymbol, pos.InstrumentToken)
			fmt.Printf("     Quantity: %d\n", pos.Quantity)
			fmt.Printf("     P&L: ₹%.2f\n", pos.PNL)
			fmt.Printf("     Last Price: ₹%.2f\n", pos.LastPrice)
			fmt.Printf("     Unrealized P&L: ₹%.2f\n", pos.Unrealised)
			fmt.Printf("     Realized P&L: ₹%.2f\n", pos.Realised)
			fmt.Println()
		}
	}

	// Example: Close a specific position
	instrumentToken := "NSE_EQ|INE062A01020" // SBIN
	fmt.Printf("2. Attempting to close position for %s...\n", instrumentToken)
	
	closeResp, err := manager.ClosePosition(instrumentToken)
	if err != nil {
		fmt.Printf("❌ Failed to close position: %v\n", err)
	} else {
		if len(closeResp.Data.OrderIDs) > 0 {
			fmt.Printf("✅ Position closed successfully! Order ID: %s\n", closeResp.Data.OrderIDs[0])
		}
	}

	// Emergency: Close all positions
	fmt.Println("\n3. Emergency exit - closing all positions...")
	fmt.Println("⚠️  This will close ALL open positions!")
	
	// Uncomment the following lines to actually close all positions
	/*
	responses, err := manager.CloseAllPositions()
	if err != nil {
		log.Fatalf("Failed to close all positions: %v", err)
	}

	fmt.Printf("✅ All positions closed! %d exit orders placed.\n", len(responses))
	for i, resp := range responses {
		if len(resp.Data.OrderIDs) > 0 {
			fmt.Printf("  Exit order %d: %s\n", i+1, resp.Data.OrderIDs[0])
		}
	}
	*/
	fmt.Println("(Commented out for safety - uncomment to execute)")
}