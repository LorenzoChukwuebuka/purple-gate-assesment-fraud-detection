# TransactionGuard: Real-time Fraud Detection System

## Overview

TransactionGuard is a high-performance system designed to identify suspicious patterns in financial transaction streams. It processes large volumes of transaction data in real-time, flags potential fraudulent activities, and provides APIs for querying and visualization.

## Features

- **Real-time Stream Processing**: Efficiently processes millions of transactions with minimal latency
- **Multiple Fraud Detection Patterns**: 
  - More than 5 transactions by the same user in under 1 minute
  - Transactions exceeding $10,000 in a single day by the same user
  - Transactions from different locations within 2 minutes
- **RESTful API Interface**: JSON API for querying flagged transactions
- **Geographic Visualization**: Interactive map showing fraud hotspots and suspicious activities
- **High Performance**: Optimized for large datasets using efficient algorithms and data structures
- **Scalable Architecture**: Designed to scale horizontally for increased load

## System Architecture

### Architecture Diagram

```
┌─────────────────┐     ┌───────────────────┐     ┌─────────────────┐
│  Transaction    │     │                   │     │    SQLite       │
│  Data Source    │────▶│  Go Backend       │────▶│    Database     │
│  (JSON Stream)  │     │  Processing Engine│     │                 │
└─────────────────┘     └─────────┬─────────┘     └─────────────────┘
                                  │
                                  │
                                  ▼
                        ┌─────────────────────┐
                        │  RESTful API        │
                        │  (Go/mux)           │
                        └──────────┬──────────┘
                                   │
                                   │
                                   ▼
                        ┌─────────────────────┐
                        │  Vue.js Frontend    │
                        │  with MapVisualize  │
                        └─────────────────────┘
```

### Components

1. **Transaction Processor**: Golang service that ingests transaction data and performs real-time fraud detection using efficient algorithms.
   - Uses Go's concurrency features (goroutines and channels) for parallel processing
   - Implements sliding window algorithms for time-based pattern detection
   - Processes large JSON files efficiently with streaming parsers

2. **SQLite Database**: Lightweight but powerful database for storing flagged transactions and user information.
   - Optimized schema design with appropriate indexes for fast lookups
   - Transaction support for data integrity
   - Efficient storage of temporal and spatial data

3. **RESTful API (Go/mux)**: API server built using the Gorilla Mux router.
   - Endpoints for querying fraud data by userId and other parameters
   - Authentication and rate limiting
   - Comprehensive API documentation

4. **Vue.js Frontend**: Interactive dashboard for visualizing and analyzing fraud patterns.
   - Map visualization showing geographic distribution of flagged transactions
   - Time-series charts displaying fraud patterns over time
   - User activity inspection tools

## Technical Implementation

### Backend (Go)

#### Data Models

```go
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
```

 

#### Fraud Detection Algorithms

1. **High Frequency Detection**:
   - Uses a sliding window algorithm to track transactions within a 1-minute window
   - Maintains an efficient map of user IDs to transaction counts
   - Time complexity: O(1) per transaction

2. **High Volume Detection**:
   - Aggregates transaction amounts by user ID and date
   - Flags when daily amount exceeds threshold
   - Uses efficient memory management for large datasets

3. **Location Hopping Detection**:
   - Employs spatial-temporal tracking for each user
   - Uses Haversine formula to calculate distances between consecutive transaction locations
   - Flags transactions with physically impossible travel times

 

### Frontend (Vue.js)

- Vue 3 with Composition API
- Vue Router for navigation
- Leaflet.js for map visualization
- Chart.js for time-series data visualization
- Axios for API communication

## Setup and Installation

### Prerequisites

- Go 1.18 or higher
- SQLite 3
- Node.js 16+ and npm
- Git

### Backend Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/transactionguard.git
   cd backend
   ```

2. Install Go dependencies:
   ```bash
   go mod download
   ```


4. Run the backend server:
   ```bash
   go run main.go
   ```

   The server will start on `http://localhost:6080`

### Frontend Setup

1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Start the development server:
   ```bash
   npm run serve
   ```

   The frontend will be available at `http://localhost:5173`

 
```

## Usage Examples

### Processing a Transaction File

```bash
# Process a JSON file and detect fraud patterns
go run seeder/seed_data.go
```

### Querying the API

```bash
# Get all flagged transactions for a specific user
curl http://localhost:6080/api/fraud-check?userId=user123

```

## Performance Optimization

The system is optimized for performance in several ways:

1. **Efficient Data Structures**: Using specialized data structures for temporal pattern matching
2. **Concurrent Processing**: Leveraging Go's goroutines for parallel processing
3. **Memory Management**: Implementing buffer pools and object recycling to minimize GC pressure
4. **Database Optimization**: Proper indexing and query optimization
5. **Streaming Processing**: Processing transactions as they arrive without loading entire dataset into memory

 

## Security Considerations

- API endpoints are protected with rate limiting to prevent abuse
- All user inputs are validated and sanitized
- Database queries use prepared statements to prevent SQL injection
- Frontend implements CSRF protection for form submissions

 
## Acknowledgments

- [Gorilla Mux](https://github.com/gorilla/mux) for HTTP routing
- [Vue.js](https://vuejs.org/) for the frontend framework
- [Leaflet](https://leafletjs.com/) for map visualization