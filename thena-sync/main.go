package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, x-api-key, x-thena-signature")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// Webhook payload structures

// WebhookPayload represents the top-level webhook structure
type WebhookPayload struct {
	Event   string                 `json:"event"`
	EventID string                 `json:"eventId"`
	Data    map[string]interface{} `json:"data"`
}

// ThenaWebhookRequest represents the detailed request object from webhook
type ThenaWebhookRequest struct {
	RequestID   interface{} `json:"requestId"` // can be int or string
	ThenaID     string      `json:"thena_id"`
	EventID     string      `json:"eventId"`
	Status      string      `json:"status"`
	SubStatus   string      `json:"subStatus"`
	Description string      `json:"description"`

	SubStatusDetails struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"subStatusDetails"`

	CustomerName string `json:"customer_name"`
	CRMData      struct {
		CRMID string `json:"crm_id"`
		Name  string `json:"name"`
	} `json:"crm_data"`

	ChannelID   string `json:"channelId"`
	ChannelName string `json:"channelName"`
	Permalink   string `json:"permalink"`
	RequestLink string `json:"requestLink"`

	Requestor struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Email       string `json:"email"`
		EmailDomain string `json:"emailDomain"`
	} `json:"requestor"`

	AssignedTo struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Email       string `json:"email"`
		EmailDomain string `json:"emailDomain"`
	} `json:"assignedTo"`

	AssignedBy struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Email       string `json:"email"`
		EmailDomain string `json:"emailDomain"`
	} `json:"assignedBy"`

	Sender struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Email       string `json:"email"`
		EmailDomain string `json:"emailDomain"`
	} `json:"sender"`

	CreatedAt               string `json:"createdAt"`
	UpdatedAt               string `json:"updatedAt"`
	FirstResponseAt         string `json:"first_response_at"`
	LastReplyByCustomerTS   string `json:"last_reply_by_customer_ts"`
	LastReplyByVendorTS     string `json:"last_reply_by_vendor_ts"`
	ReplyCount              int    `json:"replyCount"`
}

// extractRequestObject tries to find the request object from various payload structures
// Returns: (requestObject, fullPayload, error)
func extractRequestObject(body []byte) (map[string]interface{}, map[string]interface{}, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, nil, fmt.Errorf("invalid JSON: %v", err)
	}

	// Try extraction in order of priority:

	// A) Check if this is the request object directly (has requestId or thena_id or status)
	if hasRequestFields(raw) {
		return raw, raw, nil
	}

	// B) Check if data field exists and contains request
	if data, ok := raw["data"].(map[string]interface{}); ok {
		if hasRequestFields(data) {
			return data, raw, nil
		}

		// C) Check if data.request exists
		if request, ok := data["request"].(map[string]interface{}); ok {
			if hasRequestFields(request) {
				return request, raw, nil
			}
		}
	}

	// D) Check if body.request exists and contains request fields
	if request, ok := raw["request"].(map[string]interface{}); ok {
		if hasRequestFields(request) {
			return request, raw, nil
		}
	}

	// Unknown structure, return raw with indication
	return raw, raw, fmt.Errorf("unknown payload shape")
}

// hasRequestFields checks if a map contains typical request fields
func hasRequestFields(m map[string]interface{}) bool {
	_, hasRequestID := m["requestId"]
	_, hasThenaID := m["thena_id"]
	_, hasStatus := m["status"]
	return hasRequestID || hasThenaID || hasStatus
}

// extractEventID tries to find eventId from various locations in the payload
func extractEventID(requestData map[string]interface{}, fullPayload map[string]interface{}) string {
	// Try these keys in order on the request object first
	requestKeys := []string{"eventId", "event_id", "id"}
	for _, key := range requestKeys {
		if val, ok := requestData[key]; ok && val != nil {
			if str, ok := val.(string); ok && str != "" {
				return str
			}
		}
	}

	// Try on the full payload (top-level)
	fullKeys := []string{"eventId", "event_id", "id"}
	for _, key := range fullKeys {
		if val, ok := fullPayload[key]; ok && val != nil {
			if str, ok := val.(string); ok && str != "" {
				return str
			}
		}
	}

	// Try data.eventId and data.id
	if data, ok := fullPayload["data"].(map[string]interface{}); ok {
		dataKeys := []string{"eventId", "event_id", "id"}
		for _, key := range dataKeys {
			if val, ok := data[key]; ok && val != nil {
				if str, ok := val.(string); ok && str != "" {
					return str
				}
			}
		}
	}

	return ""
}

// getTopLevelKeys returns the top-level keys of a map
func getTopLevelKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// getStringField safely extracts a string field from a map
func getStringField(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// truncateString truncates a string to maxLen characters and removes newlines
func truncateString(s string, maxLen int) string {
	// Remove newlines and extra whitespace
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.Join(strings.Fields(s), " ") // normalize whitespace

	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// computeThenaUrl derives the Thena UI link from available fields
func computeThenaUrl(requestData map[string]interface{}) string {
	// Helper to safely get string value
	getString := func(key string) string {
		if val, ok := requestData[key]; ok && val != nil {
			return fmt.Sprintf("%v", val)
		}
		return ""
	}

	requestID := getString("requestId")
	if requestID == "" {
		return "" // Can't construct URL without requestId
	}

	permalink := getString("permalink")
	requestLink := getString("requestLink")
	teamID := getString("team_id")

	var thenaUrl string

	// 1) Identify the Thena UI link
	if requestLink != "" && strings.Contains(requestLink, "app.thena.ai") {
		thenaUrl = requestLink
	} else if permalink != "" && strings.Contains(permalink, "app.thena.ai") {
		thenaUrl = permalink
	} else {
		// 3) Fallback construction
		thenaUrl = fmt.Sprintf("https://app.thena.ai/requests?requestId=%s", requestID)
		if teamID != "" {
			thenaUrl += "&teamId=" + teamID
		}
		return thenaUrl
	}

	// 4) Normalization: add teamId if missing but available
	if !strings.Contains(thenaUrl, "teamId=") && teamID != "" && strings.Contains(thenaUrl, "requestId") {
		// Add teamId parameter
		if strings.Contains(thenaUrl, "?") {
			thenaUrl += "&teamId=" + teamID
		} else {
			thenaUrl += "?teamId=" + teamID
		}
	}

	return thenaUrl
}

// logWebhookSummary logs a concise summary of the webhook request
func logWebhookSummary(requestData map[string]interface{}, fullPayload map[string]interface{}) {
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("[THENA WEBHOOK SUMMARY]")

	// Helper to safely get string value
	getString := func(key string) string {
		if val, ok := requestData[key]; ok && val != nil {
			return fmt.Sprintf("%v", val)
		}
		return ""
	}

	// Helper to safely get int value
	getInt := func(key string) string {
		if val, ok := requestData[key]; ok && val != nil {
			switch v := val.(type) {
			case float64:
				return fmt.Sprintf("%d", int(v))
			case int:
				return fmt.Sprintf("%d", v)
			default:
				return fmt.Sprintf("%v", v)
			}
		}
		return ""
	}

	// Helper to safely get nested object
	getObject := func(key string) map[string]interface{} {
		if val, ok := requestData[key].(map[string]interface{}); ok {
			return val
		}
		return nil
	}

	// Extract eventId from multiple possible locations
	eventId := extractEventID(requestData, fullPayload)

	// IDs
	log.Printf("  requestId=%s thena_id=%s eventId=%s",
		getInt("requestId"),
		getString("thena_id"),
		eventId)

	// Status
	subStatusDetails := getObject("subStatusDetails")
	subStatusName := ""
	subStatusDesc := ""
	if subStatusDetails != nil {
		if name, ok := subStatusDetails["name"]; ok && name != nil {
			subStatusName = fmt.Sprintf("%v", name)
		}
		if desc, ok := subStatusDetails["description"]; ok && desc != nil {
			subStatusDesc = fmt.Sprintf("%v", desc)
		}
	}

	log.Printf("  status=%s subStatus=%s", getString("status"), getString("subStatus"))
	if subStatusName != "" {
		log.Printf("  subStatusName=%s subStatusDesc=%s", subStatusName, subStatusDesc)
	}

	// Customer / Channel
	crmData := getObject("crm_data")
	crmID := ""
	crmName := ""
	if crmData != nil {
		if id, ok := crmData["crm_id"]; ok && id != nil {
			crmID = fmt.Sprintf("%v", id)
		}
		if name, ok := crmData["name"]; ok && name != nil {
			crmName = fmt.Sprintf("%v", name)
		}
	}

	log.Printf("  customer=%s crmID=%s crmName=%s",
		getString("customer_name"), crmID, crmName)
	log.Printf("  channel=%s channelName=%s",
		getString("channelId"), getString("channelName"))

	// Links
	permalink := getString("permalink")
	requestLink := getString("requestLink")
	if permalink != "" || requestLink != "" {
		log.Printf("  permalink=%s requestLink=%s", permalink, requestLink)
	}

	// People
	requestor := getObject("requestor")
	if requestor != nil {
		log.Printf("  requestor: id=%v name=%v email=%v domain=%v",
			requestor["id"], requestor["name"], requestor["email"], requestor["emailDomain"])
	}

	assignedTo := getObject("assignedTo")
	if assignedTo != nil {
		log.Printf("  assignedTo: id=%v name=%v email=%v domain=%v",
			assignedTo["id"], assignedTo["name"], assignedTo["email"], assignedTo["emailDomain"])
	}

	assignedBy := getObject("assignedBy")
	if assignedBy != nil {
		log.Printf("  assignedBy: id=%v name=%v email=%v domain=%v",
			assignedBy["id"], assignedBy["name"], assignedBy["email"], assignedBy["emailDomain"])
	}

	sender := getObject("sender")
	if sender != nil {
		log.Printf("  sender: id=%v name=%v email=%v domain=%v",
			sender["id"], sender["name"], sender["email"], sender["emailDomain"])
	}

	// Times / Metrics
	log.Printf("  createdAt=%s updatedAt=%s replyCount=%s",
		getString("createdAt"), getString("updatedAt"), getInt("replyCount"))

	firstResponse := getString("first_response_at")
	lastCustomer := getString("last_reply_by_customer_ts")
	lastVendor := getString("last_reply_by_vendor_ts")
	if firstResponse != "" || lastCustomer != "" || lastVendor != "" {
		log.Printf("  firstResponseAt=%s lastCustomer=%s lastVendor=%s",
			firstResponse, lastCustomer, lastVendor)
	}

	// Description (truncated)
	description := getString("description")
	if description != "" {
		truncated := truncateString(description, 200)
		log.Printf("  description=\"%s\"", truncated)
	}

	// Thena URL (computed from available fields)
	thenaUrl := computeThenaUrl(requestData)
	if thenaUrl != "" {
		log.Printf("  thenaUrl=%s", thenaUrl)
	}

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// Webhook handlers

// verifyWebhookSignature validates the webhook request
func verifyWebhookSignature(body []byte, apiKey string, signature string, webhookSecret string) bool {
	if webhookSecret == "" {
		log.Println("⚠️  THENA_WEBHOOK_SECRET not set, skipping verification")
		return true
	}

	// Option A: Check x-api-key header
	if apiKey != "" && apiKey == webhookSecret {
		log.Println("✅ Verified via x-api-key header")
		return true
	}

	// Option B: Check x-thena-signature header (HMAC-SHA256)
	if signature != "" {
		mac := hmac.New(sha256.New, []byte(webhookSecret))
		mac.Write(body)
		expectedSignature := hex.EncodeToString(mac.Sum(nil))

		if hmac.Equal([]byte(signature), []byte(expectedSignature)) {
			log.Println("✅ Verified via x-thena-signature header")
			return true
		}
	}

	return false
}

// webhookHandler handles POST /webhooks/thena
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		return
	}

	// Read raw body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Error reading webhook body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error reading body"))
		return
	}
	defer r.Body.Close()

	// Verify webhook signature if secret is set
	webhookSecret := os.Getenv("THENA_WEBHOOK_SECRET")
	apiKey := r.Header.Get("X-Api-Key")
	signature := r.Header.Get("X-Thena-Signature")

	if !verifyWebhookSignature(body, apiKey, signature, webhookSecret) {
		log.Println("❌ Invalid webhook signature")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid signature"))
		return
	}

	// Extract request object from payload
	requestData, fullPayload, err := extractRequestObject(body)
	if err != nil {
		if err.Error() == "unknown payload shape" {
			// Log unknown payload structure with helpful context
			log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
			log.Printf("📥 Webhook received: %s %s", r.Method, r.URL.Path)
			log.Printf("⚠️  Unknown payload shape. Top-level keys: %v", getTopLevelKeys(requestData))

			// Log helpful string fields for categorization
			action := getStringField(requestData, "action")
			event := getStringField(requestData, "event")
			eventType := getStringField(requestData, "type")

			if action != "" || event != "" || eventType != "" {
				log.Printf("   Event metadata: action=%q event=%q type=%q", action, event, eventType)
			}

			log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		} else {
			log.Printf("❌ Failed to parse webhook payload: %v", err)
		}
	} else {
		// Log concise summary
		logWebhookSummary(requestData, fullPayload)
	}

	// Optionally log full payload if LOG_FULL_PAYLOAD is set
	if os.Getenv("LOG_FULL_PAYLOAD") == "true" {
		log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		log.Println("[FULL PAYLOAD DEBUG]")
		var jsonData interface{}
		if err := json.Unmarshal(body, &jsonData); err == nil {
			prettyJSON, _ := json.MarshalIndent(jsonData, "", "  ")
			log.Println(string(prettyJSON))
		} else {
			log.Printf("(Raw) %s", string(body))
		}
		log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	}

	// Respond 200 OK
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// healthzHandler handles GET /healthz
func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
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
	http.HandleFunc("/healthz", healthzHandler)
	http.HandleFunc("/webhooks/thena", webhookHandler)
	http.HandleFunc("/api/test-connection", enableCORS(testConnectionHandler))
	http.HandleFunc("/api/requests", enableCORS(getRequestsHandler))
	http.HandleFunc("/api/requests/", enableCORS(getRequestByIDHandler))
	http.HandleFunc("/api/cards-by-status", enableCORS(getCardsByStatusHandler))

	port := getEnvOrDefault("PORT", "9100")
	webhookSecret := os.Getenv("THENA_WEBHOOK_SECRET")

	log.Printf("Starting Thena Sync service on port %s", port)
	log.Printf("Base URL: %s", thenaClient.BaseURL)
	log.Printf("Webhook endpoint: /webhooks/thena")

	if webhookSecret == "" {
		log.Println("⚠️  THENA_WEBHOOK_SECRET not set - webhooks will accept all requests without verification")
	} else {
		log.Println("✅ THENA_WEBHOOK_SECRET configured - webhook signature verification enabled")
	}

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
