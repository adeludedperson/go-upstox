package main

import (
	"fmt"
	"log"

	"github.com/rehanshaik/go-upstox"
)

func main() {
	// Replace with your actual credentials and access token
	manager := upstox.NewManager("your_client_id", "your_client_secret", "your_access_token")

	// Get all funds (both equity and commodity)
	fmt.Println("Getting all funds...")
	funds, err := manager.GetFundsAndMargin()
	if err != nil {
		log.Fatalf("Failed to get funds: %v", err)
	}

	fmt.Printf("Status: %s\n", funds.Status)
	fmt.Println("\nEquity Funds:")
	fmt.Printf("  Used Margin: %.2f\n", funds.Data.Equity.UsedMargin)
	fmt.Printf("  Available Margin: %.2f\n", funds.Data.Equity.AvailableMargin)
	fmt.Printf("  Payin Amount: %.2f\n", funds.Data.Equity.PayinAmount)
	fmt.Printf("  Span Margin: %.2f\n", funds.Data.Equity.SpanMargin)
	fmt.Printf("  Exposure Margin: %.2f\n", funds.Data.Equity.ExposureMargin)
	fmt.Printf("  Adhoc Margin: %.2f\n", funds.Data.Equity.AdhocMargin)
	fmt.Printf("  Notional Cash: %.2f\n", funds.Data.Equity.NotionalCash)

	fmt.Println("\nCommodity Funds:")
	fmt.Printf("  Used Margin: %.2f\n", funds.Data.Commodity.UsedMargin)
	fmt.Printf("  Available Margin: %.2f\n", funds.Data.Commodity.AvailableMargin)
	fmt.Printf("  Payin Amount: %.2f\n", funds.Data.Commodity.PayinAmount)
	fmt.Printf("  Span Margin: %.2f\n", funds.Data.Commodity.SpanMargin)
	fmt.Printf("  Exposure Margin: %.2f\n", funds.Data.Commodity.ExposureMargin)
	fmt.Printf("  Adhoc Margin: %.2f\n", funds.Data.Commodity.AdhocMargin)
	fmt.Printf("  Notional Cash: %.2f\n", funds.Data.Commodity.NotionalCash)

	// Get only equity funds
	fmt.Println("\n\nGetting equity funds only...")
	equityFunds, err := manager.GetFundsAndMargin("SEC")
	if err != nil {
		log.Fatalf("Failed to get equity funds: %v", err)
	}

	fmt.Printf("Equity Available Margin: %.2f\n", equityFunds.Data.Equity.AvailableMargin)

	// Get only commodity funds
	fmt.Println("\nGetting commodity funds only...")
	commodityFunds, err := manager.GetFundsAndMargin("COM")
	if err != nil {
		log.Fatalf("Failed to get commodity funds: %v", err)
	}

	fmt.Printf("Commodity Available Margin: %.2f\n", commodityFunds.Data.Commodity.AvailableMargin)
}