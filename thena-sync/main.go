package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// ThenaClient handles authentication and API requests
type ThenaClient struct {
	APIKey  string
	BaseURL string
}

// ThenaRequest represents a request from Thena
type ThenaRequest struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   int64     `json:"created_at"`
	UpdatedAt   int64     `json:"updated_at"`
	Submitter   Submitter `json:"submitter"`
	Priority    string    `json:"priority"`
	Category    string    `json:"category"`
	Assignee    Assignee  `json:"assignee"`
	Customer    Customer  `json:"customer"`
}

type Submitter struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Assignee struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Customer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ThenaResponse represents the API response
type ThenaResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Requests []ThenaRequest `json:"requests"`
		PageInfo PageInfo       `json:"page_info"`
	} `json:"data"`
}

type PageInfo struct {
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
	TotalItems  int `json:"total_items"`
	PerPage     int `json:"per_page"`
}

// Card represents a simplified card for the dashboard
type Card struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	CustomerName string `json:"customerName"`
	AssignedTo   string `json:"assignedTo"`
	CreatedAt    string `json:"createdAt"`
	Priority     string `json:"priority"`
}

var thenaClient *ThenaClient

// Initialize Thena client
func NewThenaClient() *ThenaClient {
	return &ThenaClient{
		APIKey:  os.Getenv("THENA_API_KEY"),
		BaseURL: getEnvOrDefault("THENA_BASE_URL", "https://bolt.thena.ai"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetAllRequests fetches all requests from Thena
func (tc *ThenaClient) GetAllRequests(limit, page int, sortOrder string) (*ThenaResponse, error) {
	if tc.APIKey == "" {
		return nil, fmt.Errorf("Thena API key not configured")
	}

	url := fmt.Sprintf("%s/rest/v2/requests", tc.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Add query parameters
	q := req.URL.Query()
	if limit > 0 {
		q.Add("limit", fmt.Sprintf("%d", limit))
	}
	if page > 0 {
		q.Add("page", fmt.Sprintf("%d", page))
	}
	if sortOrder != "" {
		q.Add("sort_order", sortOrder)
	}
	req.URL.RawQuery = q.Encode()

	// Add authentication header
	req.Header.Set("Authorization", "Bearer "+tc.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var thenaResp ThenaResponse
	if err := json.Unmarshal(body, &thenaResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &thenaResp, nil
}

// GetRequestByID fetches a single request by ID
func (tc *ThenaClient) GetRequestByID(requestID string) (*ThenaRequest, error) {
	if tc.APIKey == "" {
		return nil, fmt.Errorf("Thena API key not configured")
	}

	url := fmt.Sprintf("%s/rest/v2/requests/%s", tc.BaseURL, requestID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+tc.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Status  string        `json:"status"`
		Data    ThenaRequest  `json:"data"`
		Message string        `json:"message"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &apiResp.Data, nil
}

// MapRequestToCard converts a Thena request to a Card
func MapRequestToCard(req ThenaRequest) Card {
	createdAt := time.Unix(req.CreatedAt, 0).Format(time.RFC3339)

	return Card{
		ID:           req.ID,
		Title:        req.Title,
		Description:  req.Description,
		Status:       req.Status,
		CustomerName: req.Customer.Name,
		AssignedTo:   req.Assignee.Name,
		CreatedAt:    createdAt,
		Priority:     req.Priority,
	}
}

// HTTP Handlers

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"status":         "healthy",
		"service":        "thena-sync",
		"time":           time.Now().Format(time.RFC3339),
		"thenaConnected": thenaClient.APIKey != "",
	}

	json.NewEncoder(w).Encode(response)
}

func testConnectionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if thenaClient.APIKey == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Thena API key not configured",
			"mode":    "mock",
		})
		return
	}

	// Try to fetch requests with limit 1 to test connection
	_, err := thenaClient.GetAllRequests(1, 1, "desc")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to connect: %v", err),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Successfully connected to Thena API",
		"baseURL": thenaClient.BaseURL,
	})
}

func getRequestsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if thenaClient.APIKey == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  false,
			"message":  "Thena not connected. Please configure THENA_API_KEY.",
			"mode":     "mock",
			"requests": []Card{},
		})
		return
	}

	// Get query parameters
	limit := 100 // default limit
	page := 1    // default page
	sortOrder := "desc"

	thenaResp, err := thenaClient.GetAllRequests(limit, page, sortOrder)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  false,
			"message":  fmt.Sprintf("Failed to fetch requests: %v", err),
			"requests": []Card{},
		})
		return
	}

	// Map to cards
	cards := make([]Card, 0, len(thenaResp.Data.Requests))
	for _, req := range thenaResp.Data.Requests {
		cards = append(cards, MapRequestToCard(req))
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"totalCount": thenaResp.Data.PageInfo.TotalItems,
		"page":      thenaResp.Data.PageInfo.CurrentPage,
		"totalPages": thenaResp.Data.PageInfo.TotalPages,
		"requests":  cards,
	})
}

func getRequestByIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract ID from path
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Request ID required"})
		return
	}
	requestID := pathParts[3]

	if thenaClient.APIKey == "" {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Thena not connected",
		})
		return
	}

	request, err := thenaClient.GetRequestByID(requestID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to get request: %v", err),
		})
		return
	}

	card := MapRequestToCard(*request)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"request": card,
	})
}

func getCardsByStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if thenaClient.APIKey == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Thena not connected. Please configure THENA_API_KEY.",
			"mode":    "mock",
			"cards":   map[string][]Card{},
		})
		return
	}

	thenaResp, err := thenaClient.GetAllRequests(100, 1, "desc")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to fetch requests: %v", err),
			"cards":   map[string][]Card{},
		})
		return
	}

	// Group by status
	cardsByStatus := make(map[string][]Card)
	for _, req := range thenaResp.Data.Requests {
		card := MapRequestToCard(req)
		status := card.Status
		if status == "" {
			status = "unknown"
		}
		cardsByStatus[status] = append(cardsByStatus[status], card)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"cards":   cardsByStatus,
	})
}

func splitPath(path string) []string {
	parts := []string{}
	for _, part := range splitString(path, '/') {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

func splitString(s string, sep rune) []string {
	result := []string{}
	current := ""
	for _, char := range s {
		if char == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	result = append(result, current)
	return result
}

// CORS middleware
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	// Initialize Thena client
	thenaClient = NewThenaClient()

	if thenaClient.APIKey == "" {
		log.Println("THENA_API_KEY not configured. Running in mock mode.")
		log.Println("To connect to Thena, set the THENA_API_KEY environment variable.")
	} else {
		log.Println("Thena API key configured. Service ready.")
	}

	// Setup routes
	http.HandleFunc("/health", enableCORS(healthHandler))
	http.HandleFunc("/api/test-connection", enableCORS(testConnectionHandler))
	http.HandleFunc("/api/requests", enableCORS(getRequestsHandler))
	http.HandleFunc("/api/requests/", enableCORS(getRequestByIDHandler))
	http.HandleFunc("/api/cards-by-status", enableCORS(getCardsByStatusHandler))

	port := getEnvOrDefault("PORT", "9100")
	log.Printf("Starting Thena Sync service on port %s", port)
	log.Printf("Base URL: %s", thenaClient.BaseURL)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
