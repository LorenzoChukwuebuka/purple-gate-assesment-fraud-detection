package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"github.com/gorilla/handlers"
)

// Transaction represents a financial transaction from the input data
type Transaction struct {
	TransactionID string    `json:"transactionId"`
	UserID        string    `json:"userId"`
	Amount        float64   `json:"amount"`
	Timestamp     time.Time `json:"timestamp"`
	Merchant      string    `json:"merchant"`
	Location      string    `json:"location"`
}

// FlaggedTransaction represents a transaction that has been flagged as suspicious
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

// UserTransactionCache maintains a cache of recent transactions for each user
type UserTransactionCache struct {
	mu           sync.RWMutex
	transactions map[string][]Transaction
}

// NewUserTransactionCache creates a new transaction cache
func NewUserTransactionCache() *UserTransactionCache {
	return &UserTransactionCache{
		transactions: make(map[string][]Transaction),
	}
}

// Add adds a transaction to the cache
func (c *UserTransactionCache) Add(t Transaction) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.transactions[t.UserID] = append(c.transactions[t.UserID], t)
}

// Get retrieves transactions for a user
func (c *UserTransactionCache) Get(userID string) []Transaction {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.transactions[userID]
}

// CleanUp removes old transactions from the cache
func (c *UserTransactionCache) CleanUp(cutoff time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for userID, transactions := range c.transactions {
		var recent []Transaction
		for _, t := range transactions {
			if t.Timestamp.After(cutoff) {
				recent = append(recent, t)
			}
		}
		if len(recent) > 0 {
			c.transactions[userID] = recent
		} else {
			delete(c.transactions, userID)
		}
	}
}

// FraudDetector detects suspicious transactions
type FraudDetector struct {
	db           *gorm.DB
	cache        *UserTransactionCache
	dailyAmounts map[string]map[string]float64 // userID -> date -> amount
	mu           sync.Mutex
}

// NewFraudDetector creates a new fraud detector
func NewFraudDetector(db *gorm.DB) *FraudDetector {
	return &FraudDetector{
		db:           db,
		cache:        NewUserTransactionCache(),
		dailyAmounts: make(map[string]map[string]float64),
	}
}

// ProcessTransaction processes a transaction and flags it if it's suspicious
func (fd *FraudDetector) ProcessTransaction(t Transaction) {
	// Add transaction to cache
	fd.cache.Add(t)

	// Check for suspicious patterns
	fd.checkHighFrequency(t)
	fd.checkLargeAmount(t)
	fd.checkDifferentLocations(t)

	// Periodically clean up cache
	if time.Now().Second()%30 == 0 {
		fd.cache.CleanUp(time.Now().Add(-24 * time.Hour))
	}
}

// checkHighFrequency checks if a user has made more than 5 transactions in under 1 minute
func (fd *FraudDetector) checkHighFrequency(t Transaction) {
	transactions := fd.cache.Get(t.UserID)
	if len(transactions) < 5 {
		return
	}

	// Count transactions in the last minute
	count := 0
	for _, tx := range transactions {
		if tx.Timestamp.After(t.Timestamp.Add(-1 * time.Minute)) {
			count++
		}
	}

	if count > 5 {
		fd.flagTransaction(t, "High frequency: More than 5 transactions in 1 minute")
	}
}

// checkLargeAmount checks if a user has made transactions exceeding $10,000 in a single day
func (fd *FraudDetector) checkLargeAmount(t Transaction) {
	fd.mu.Lock()
	defer fd.mu.Unlock()

	dateStr := t.Timestamp.Format("2006-01-02")

	// Initialize maps if needed
	if _, exists := fd.dailyAmounts[t.UserID]; !exists {
		fd.dailyAmounts[t.UserID] = make(map[string]float64)
	}

	// Add amount to daily total
	fd.dailyAmounts[t.UserID][dateStr] += t.Amount

	// Check if daily total exceeds threshold
	if fd.dailyAmounts[t.UserID][dateStr] > 10000 {
		fd.flagTransaction(t, "Large amount: Transactions exceeding $10,000 in a single day")
	}
}

// checkDifferentLocations checks if a user has made transactions from different locations within 2 minutes
func (fd *FraudDetector) checkDifferentLocations(t Transaction) {
	transactions := fd.cache.Get(t.UserID)

	for _, tx := range transactions {
		// Skip the current transaction
		if tx.TransactionID == t.TransactionID {
			continue
		}

		// Check if the transaction is within 2 minutes
		if tx.Timestamp.After(t.Timestamp.Add(-2*time.Minute)) &&
			tx.Timestamp.Before(t.Timestamp.Add(2*time.Minute)) {
			// Check if the location is different
			if tx.Location != t.Location {
				fd.flagTransaction(t, "Location jump: Transactions from different locations within 2 minutes")
				break
			}
		}
	}
}

// flagTransaction flags a transaction as suspicious
func (fd *FraudDetector) flagTransaction(t Transaction, reason string) {
	flagged := FlaggedTransaction{
		TransactionID: t.TransactionID,
		UserID:        t.UserID,
		Amount:        t.Amount,
		Timestamp:     t.Timestamp,
		Merchant:      t.Merchant,
		Location:      t.Location,
		Reason:        reason,
	}

	result := fd.db.Create(&flagged)
	if result.Error != nil {
		log.Printf("Error flagging transaction: %v", result.Error)
	}
}

// API handlers

// GetFlaggedTransactions returns flagged transactions for a user
func (fd *FraudDetector) GetFlaggedTransactions(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	log.Printf("Fetching flagged transactions for user: %s", userID)

	var flagged []FlaggedTransaction
	result := fd.db.Where("user_id = ?", userID).Find(&flagged)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d flagged transactions", len(flagged))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flagged)
}

// ProcessExistingTransactions processes all transactions from the database
func (fd *FraudDetector) ProcessExistingTransactions() error {
	var transactions []Transaction

	// Query all transactions - for very large databases, use pagination
	result := fd.db.Find(&transactions)
	if result.Error != nil {
		return result.Error
	}

	log.Printf("Processing %d existing transactions from database", len(transactions))

	// Process in batches
	batchSize := 100
	for i := 0; i < len(transactions); i += batchSize {
		end := i + batchSize
		if end > len(transactions) {
			end = len(transactions)
		}

		for j := i; j < end; j++ {
			fd.ProcessTransaction(transactions[j])
		}
	}

	return nil
}

func main() {
	// Connect to database
	db, err := gorm.Open(sqlite.Open("fraud.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Migrate schema
	db.AutoMigrate(&Transaction{}, &FlaggedTransaction{})

	// Create fraud detector
	fd := NewFraudDetector(db)

	// Process existing transactions
	if err := fd.ProcessExistingTransactions(); err != nil {
		log.Printf("Error processing existing transactions: %v", err)
	}

	// Set up HTTP router
	r := mux.NewRouter()
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/fraud-check", fd.GetFlaggedTransactions).Methods("GET")

	// Enable CORS
	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)

	// Apply CORS middleware and start server
	handler := corsMiddleware(r)

	// Start server
	port := 6080
	log.Printf("Starting server on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), handler))
}
