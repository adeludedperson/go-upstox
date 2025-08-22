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

	fmt.Println("=== Order Tracking Example ===")

	// Get order book (all orders for the day)
	fmt.Println("\n1. Fetching order book...")
	orders, err := manager.GetOrderBook()
	if err != nil {
		log.Fatalf("Failed to get order book: %v", err)
	}

	if len(orders) == 0 {
		fmt.Println("No orders found for today.")
	} else {
		fmt.Printf("Found %d orders:\n", len(orders))
		for i, order := range orders {
			fmt.Printf("  %d. Order ID: %s\n", i+1, order.OrderID)
			fmt.Printf("     Symbol: %s\n", order.TradingSymbol)
			fmt.Printf("     Type: %s %s\n", order.TransactionType, order.OrderType)
			fmt.Printf("     Quantity: %d (Filled: %d, Pending: %d)\n", 
				order.Quantity, order.FilledQuantity, order.PendingQuantity)
			fmt.Printf("     Status: %s\n", order.Status)
			fmt.Printf("     Price: ₹%.2f (Avg: ₹%.2f)\n", order.Price, order.AveragePrice)
			fmt.Printf("     Timestamp: %s\n", order.OrderTimestamp)
			if order.StatusMessage != "" {
				fmt.Printf("     Message: %s\n", order.StatusMessage)
			}
			fmt.Println()
		}

		// Get details of the first order
		if len(orders) > 0 {
			firstOrderID := orders[0].OrderID
			fmt.Printf("2. Getting detailed information for order %s...\n", firstOrderID)
			
			orderDetail, err := manager.GetOrderDetails(firstOrderID)
			if err != nil {
				log.Fatalf("Failed to get order details: %v", err)
			}

			fmt.Printf("✅ Order Details Retrieved:\n")
			fmt.Printf("   Order ID: %s\n", orderDetail.OrderID)
			fmt.Printf("   Exchange Order ID: %s\n", orderDetail.ExchangeOrderID)
			fmt.Printf("   Symbol: %s (%s)\n", orderDetail.TradingSymbol, orderDetail.InstrumentToken)
			fmt.Printf("   Type: %s %s\n", orderDetail.TransactionType, orderDetail.OrderType)
			fmt.Printf("   Product: %s | Validity: %s\n", orderDetail.Product, orderDetail.Validity)
			fmt.Printf("   Quantity: %d\n", orderDetail.Quantity)
			fmt.Printf("   Price: ₹%.2f\n", orderDetail.Price)
			fmt.Printf("   Status: %s\n", orderDetail.Status)
			fmt.Printf("   Filled Quantity: %d\n", orderDetail.FilledQuantity)
			fmt.Printf("   Average Price: ₹%.2f\n", orderDetail.AveragePrice)
			fmt.Printf("   Pending Quantity: %d\n", orderDetail.PendingQuantity)
			fmt.Printf("   Order Timestamp: %s\n", orderDetail.OrderTimestamp)
			fmt.Printf("   Exchange Timestamp: %s\n", orderDetail.ExchangeTimestamp)
			fmt.Printf("   Variety: %s\n", orderDetail.Variety)
			fmt.Printf("   Is AMO: %t\n", orderDetail.IsAMO)
			
			if orderDetail.Tag != "" {
				fmt.Printf("   Tag: %s\n", orderDetail.Tag)
			}
			
			if orderDetail.StatusMessage != "" {
				fmt.Printf("   Status Message: %s\n", orderDetail.StatusMessage)
			}
		}
	}

	// Filter orders by status
	fmt.Println("\n3. Filtering orders by status...")
	completedOrders := 0
	pendingOrders := 0
	rejectedOrders := 0

	for _, order := range orders {
		switch order.Status {
		case "complete":
			completedOrders++
		case "open", "pending":
			pendingOrders++
		case "rejected", "cancelled":
			rejectedOrders++
		}
	}

	fmt.Printf("Order Summary:\n")
	fmt.Printf("  ✅ Completed: %d\n", completedOrders)
	fmt.Printf("  ⏳ Pending: %d\n", pendingOrders)
	fmt.Printf("  ❌ Rejected/Cancelled: %d\n", rejectedOrders)
}