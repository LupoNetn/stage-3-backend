package main

import (
	"fmt"
	"time"
	"github.com/google/uuid"
	"github.com/luponetn/hng-stage-1/utils"
	"os"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env to get JWT_SECRET
	godotenv.Load()
	
	userID, _ := uuid.NewV7()
	
	// Generate token for an analyst
	token, err := utils.GenerateToken(
		userID.String(),
		"analyst_tester",
		"analyst_github_123",
		"analyst@example.com",
		"analyst",
		24*time.Hour,
	)
	
	if err != nil {
		fmt.Printf("Error generating token: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("\n--- ANALYST TEST TOKEN ---")
	fmt.Println(token)
	fmt.Println("--------------------------")
}
