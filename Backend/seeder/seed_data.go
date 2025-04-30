// seeder.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Transaction model - same as in main.go
type Transaction struct {
	TransactionID string    `json:"transactionId" gorm:"primaryKey"`
	UserID        string    `json:"userId" gorm:"index"`
	Amount        float64   `json:"amount"`
	Timestamp     time.Time `json:"timestamp"`
	Merchant      string    `json:"merchant"`
	Location      string    `json:"location"`
}

// FlaggedTransaction model - same as in main.go
type FlaggedTransaction struct {
	gorm.Model
	TransactionID string    `json:"transactionId" gorm:"index"`
	UserID        string    `json:"userId" gorm:"index"`
	Amount        float64   `json:"amount"`
	Timestamp     time.Time `json:"timestamp"`
	Merchant      string    `json:"merchant"`
	Location      string    `json:"location"`
	Reason        string    `json:"reason"`
}

func main() {
	// Set random seed
	rand.Seed(time.Now().UnixNano())

	// Connect to database
	db, err := gorm.Open(sqlite.Open("fraud.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create tables if they don't exist
	db.AutoMigrate(&Transaction{}, &FlaggedTransaction{})

	// Clear existing data (optional)
	db.Exec("DELETE FROM transactions")
	db.Exec("DELETE FROM flagged_transactions")

	// Generate and seed transaction data
	generateAndSeedTransactions(db, 10000) // Seed 10,000 transactions

	// Generate some suspicious patterns for testing
	generateSuspiciousPatterns(db)

	fmt.Println("Database seeded successfully!")

	// Optionally create a JSON export of the data
	exportTransactionsToJSON(db, "exported_transactions.json")
}

func generateAndSeedTransactions(db *gorm.DB, count int) {
	// Users and merchants for more realistic data
	users := []string{"user1", "user2", "user3", "user4", "user5", "user6", "user7", "user8", "user9", "user10"}
	merchants := []string{"Amazon", "Walmart", "Target", "Costco", "Starbucks", "McDonald's", "Apple", "Microsoft", "Netflix", "Spotify"}
	
	// Locations (roughly US cities)
	locations := []string{
		"40.7128,-74.0060", // New York
		"34.0522,-118.2437", // Los Angeles
		"41.8781,-87.6298", // Chicago
		"29.7604,-95.3698", // Houston
		"33.4484,-112.0740", // Phoenix
		"39.9526,-75.1652", // Philadelphia
		"29.4241,-98.4936", // San Antonio
		"32.7157,-117.1611", // San Diego
		"30.2672,-97.7431", // Austin
		"37.7749,-122.4194", // San Francisco
	}

	// Time range for transactions (last 30 days)
	endTime := time.Now()
	startTime := endTime.AddDate(0, -1, 0)
	timeRange := endTime.Sub(startTime).Seconds()

	// Create transactions in batches
	batchSize := 100
	transactions := make([]Transaction, 0, batchSize)

	for i := 0; i < count; i++ {
		// Create random transaction
		txTime := startTime.Add(time.Duration(rand.Float64()*timeRange) * time.Second)
		userID := users[rand.Intn(len(users))]
		
		transaction := Transaction{
			TransactionID: fmt.Sprintf("tx-%d", i+1),
			UserID:        userID,
			Amount:        float64(rand.Intn(100000)) / 100.0, // Random amount up to $1000
			Timestamp:     txTime,
			Merchant:      merchants[rand.Intn(len(merchants))],
			Location:      locations[rand.Intn(len(locations))],
		}
		
		transactions = append(transactions, transaction)
		
		// Insert batch when full
		if len(transactions) >= batchSize || i == count-1 {
			db.CreateInBatches(transactions, batchSize)
			transactions = transactions[:0] // Clear slice but keep capacity
		}
	}
}

func generateSuspiciousPatterns(db *gorm.DB) {
	// 1. High frequency pattern (more than 5 transactions in under 1 minute)
	baseTime := time.Now().Add(-1 * time.Hour)
	for i := 0; i < 7; i++ {
		txTime := baseTime.Add(time.Duration(i*8) * time.Second)
		db.Create(&Transaction{
			TransactionID: fmt.Sprintf("high-freq-%d", i),
			UserID:        "suspicious1",
			Amount:        99.99,
			Timestamp:     txTime,
			Merchant:      "FastFood",
			Location:      "40.7128,-74.0060",
		})
	}
	
	// 2. Large amount pattern (exceeding $10,000 in a single day)
	baseTime = time.Now().Add(-2 * time.Hour)
	totalAmount := 0.0
	for i := 0; i < 3; i++ {
		amount := 3500.0 + float64(i*100)
		totalAmount += amount
		db.Create(&Transaction{
			TransactionID: fmt.Sprintf("large-amount-%d", i),
			UserID:        "suspicious2",
			Amount:        amount,
			Timestamp:     baseTime.Add(time.Duration(i) * time.Hour),
			Merchant:      "LuxuryStore",
			Location:      "34.0522,-118.2437",
		})
	}
	
	// 3. Different locations pattern (transactions from different locations within 2 minutes)
	baseTime = time.Now().Add(-3 * time.Hour)
	locations := []string{"40.7128,-74.0060", "29.7604,-95.3698"}
	for i := 0; i < 2; i++ {
		db.Create(&Transaction{
			TransactionID: fmt.Sprintf("location-jump-%d", i),
			UserID:        "suspicious3",
			Amount:        250.00,
			Timestamp:     baseTime.Add(time.Duration(i) * time.Minute),
			Merchant:      "FastTravel",
			Location:      locations[i],
		})
	}
}

func exportTransactionsToJSON(db *gorm.DB, filename string) {
	var transactions []Transaction
	db.Find(&transactions)
	
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Error creating export file: %v", err)
		return
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(transactions); err != nil {
		log.Printf("Error encoding transactions: %v", err)
	}
}