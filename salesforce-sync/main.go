package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Config holds all configuration for the Salesforce integration
type Config struct {
	ClientID           string
	Username           string
	Aud                string // production or sandbox
	PrivateKey         *rsa.PrivateKey
	APIVersion         string
	HTTPTimeout        time.Duration
	MaxRetries         int
	BackoffBase        time.Duration
	EnableCache        bool
	CacheTTL           time.Duration
	LoginURL           string
}

// SalesforceClient handles JWT-based authentication and API requests
type SalesforceClient struct {
	config        *Config
	instanceURL   string
	accessToken   string
	tokenExpiry   time.Time
	tokenMutex    sync.RWMutex
	httpClient    *http.Client
	cache         map[string]*cacheEntry
	cacheMutex    sync.RWMutex
}

type cacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

// JWTClaims represents the JWT claims for Salesforce
type JWTClaims struct {
	Iss string `json:"iss"` // Client ID
	Sub string `json:"sub"` // Username
	Aud string `json:"aud"` // Login URL
	Exp int64  `json:"exp"` // Expiration time
}

// TokenResponse represents Salesforce OAuth token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	InstanceURL string `json:"instance_url"`
	ID          string `json:"id"`
	TokenType   string `json:"token_type"`
}

// SalesforceAccount represents a Salesforce Account object
type SalesforceAccount struct {
	ID                     string  `json:"Id"`
	Name                   string  `json:"Name"`
	Description            string  `json:"Description"`
	CreatedDate            string  `json:"CreatedDate"`
	LastModifiedDate       string  `json:"LastModifiedDate"`
	LastActivityDate       string  `json:"LastActivityDate"`
	Owner                  Owner   `json:"Owner"`
	CustomerStatus         string  `json:"Customer_Status__c"`
	Tags                   string  `json:"Tags__c"`
	AnnualRevenue          float64 `json:"Annual_Revenue__c"`
	RenewalDate            string  `json:"Renewal_Date__c"`
	DaysUntilRenewal       int     `json:"Days_Until_Renewal__c"`
	NumberOfUnits          int     `json:"Number_of_Units__c"`
	AccountsCreated        int     `json:"Accounts_Created__c"`
	MonthlyActiveUsers     int     `json:"Monthly_Active_Users__c"`
	InstallType            string  `json:"Install_Type__c"`
	Region                 string  `json:"Region__c"`
	SAOwner                string  `json:"SA_Owner__c"`
}

type Owner struct {
	Name string `json:"Name"`
}

// QueryResponse represents a Salesforce SOQL query response
type QueryResponse struct {
	TotalSize int                 `json:"totalSize"`
	Done      bool                `json:"done"`
	Records   []SalesforceAccount `json:"records"`
}

// Client represents the mapped client data structure
type Client struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	Status                 string   `json:"status"`
	Owner                  string   `json:"owner"`
	CreatedAt              string   `json:"createdAt"`
	LastActivity           string   `json:"lastActivity"`
	Tags                   []string `json:"tags"`
	Summary                string   `json:"summary"`
	TotalARR               string   `json:"totalARR"`
	NextRenewalDate        string   `json:"nextRenewalDate"`
	DaysUntilRenewal       int      `json:"daysUntilRenewal"`
	NumberOfUnits          int      `json:"numberOfUnits"`
	CurrentAccountsCreated int      `json:"currentAccountsCreated"`
	CurrentMAU             int      `json:"currentMAU"`
	InstallType            string   `json:"installType"`
	Region                 string   `json:"region"`
	SAOwner                string   `json:"saOwner"`
}

var sfClient *SalesforceClient

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Required fields
	clientID := os.Getenv("SALESFORCE_CLIENT_ID")
	username := os.Getenv("SALESFORCE_USERNAME")
	aud := os.Getenv("SALESFORCE_AUD")

	if clientID == "" {
		return nil, fmt.Errorf("SALESFORCE_CLIENT_ID is required")
	}
	if username == "" {
		return nil, fmt.Errorf("SALESFORCE_USERNAME is required")
	}
	if aud == "" {
		return nil, fmt.Errorf("SALESFORCE_AUD is required (e.g., 'https://login.salesforce.com' or 'https://test.salesforce.com')")
	}

	// Load private key
	privateKey, err := loadPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	// Optional fields with defaults
	apiVersion := getEnvOrDefault("SALESFORCE_API_VERSION", "v58.0")
	httpTimeout := getEnvAsDuration("SALESFORCE_HTTP_TIMEOUT_MS", 15000) * time.Millisecond
	maxRetries := getEnvAsInt("SALESFORCE_MAX_RETRIES", 3)
	backoffBase := getEnvAsDuration("SALESFORCE_BACKOFF_BASE_MS", 200) * time.Millisecond
	enableCache := getEnvAsBool("SALESFORCE_ENABLE_CACHE", false)
	cacheTTL := getEnvAsDuration("SALESFORCE_CACHE_TTL_SECONDS", 300) * time.Second

	// Determine login URL based on environment or explicit aud
	loginURL := aud
	if !strings.HasPrefix(aud, "http") {
		// If aud is just "production" or "sandbox", convert to URL
		if strings.ToLower(aud) == "sandbox" || strings.ToLower(aud) == "test" {
			loginURL = "https://test.salesforce.com"
		} else {
			loginURL = "https://login.salesforce.com"
		}
	}

	log.Printf("✓ Configuration loaded: ClientID=%s, Username=%s, LoginURL=%s, APIVersion=%s",
		maskClientID(clientID), username, loginURL, apiVersion)

	return &Config{
		ClientID:    clientID,
		Username:    username,
		Aud:         loginURL,
		PrivateKey:  privateKey,
		APIVersion:  apiVersion,
		HTTPTimeout: httpTimeout,
		MaxRetries:  maxRetries,
		BackoffBase: backoffBase,
		EnableCache: enableCache,
		CacheTTL:    cacheTTL,
		LoginURL:    loginURL,
	}, nil
}

// loadPrivateKey loads the private key from various sources
func loadPrivateKey() (*rsa.PrivateKey, error) {
	var pemData []byte

	// Try SALESFORCE_PRIVATE_KEY (direct PEM string)
	if keyStr := os.Getenv("SALESFORCE_PRIVATE_KEY"); keyStr != "" {
		log.Println("✓ Loading private key from SALESFORCE_PRIVATE_KEY environment variable")
		pemData = []byte(keyStr)
	} else if keyPath := os.Getenv("SALESFORCE_PRIVATE_KEY_PATH"); keyPath != "" {
		// Try SALESFORCE_PRIVATE_KEY_PATH (file path)
		log.Printf("✓ Loading private key from file: %s", keyPath)
		data, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file %s: %w", keyPath, err)
		}
		pemData = data
	} else if secretName := os.Getenv("SALESFORCE_PRIVATE_KEY_SECRET_NAME"); secretName != "" {
		// Try loading from secret manager (this is a placeholder - implement based on your secret manager)
		return nil, fmt.Errorf("SALESFORCE_PRIVATE_KEY_SECRET_NAME is set but secret manager integration is not yet implemented. Use SALESFORCE_PRIVATE_KEY or SALESFORCE_PRIVATE_KEY_PATH instead")
	} else {
		return nil, fmt.Errorf("no private key source configured. Set one of: SALESFORCE_PRIVATE_KEY, SALESFORCE_PRIVATE_KEY_PATH, or SALESFORCE_PRIVATE_KEY_SECRET_NAME")
	}

	// Parse PEM
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block - ensure the key is in PEM format")
	}

	// Try parsing as PKCS#1
	if privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		log.Println("✓ Private key loaded successfully (PKCS#1 format)")
		return privateKey, nil
	}

	// Try parsing as PKCS#8
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		if rsaKey, ok := key.(*rsa.PrivateKey); ok {
			log.Println("✓ Private key loaded successfully (PKCS#8 format)")
			return rsaKey, nil
		}
		return nil, fmt.Errorf("PKCS#8 key is not an RSA key")
	}

	return nil, fmt.Errorf("failed to parse private key - ensure it's in PKCS#1 or PKCS#8 PEM format")
}

// NewSalesforceClient creates a new Salesforce client
func NewSalesforceClient(config *Config) *SalesforceClient {
	return &SalesforceClient{
		config: config,
		httpClient: &http.Client{
			Timeout: config.HTTPTimeout,
		},
		cache: make(map[string]*cacheEntry),
	}
}

// createJWT creates a signed JWT assertion for Salesforce
func (sf *SalesforceClient) createJWT() (string, error) {
	// JWT Header
	header := map[string]string{
		"alg": "RS256",
		"typ": "JWT",
	}
	headerJSON, _ := json.Marshal(header)
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	// JWT Claims
	claims := JWTClaims{
		Iss: sf.config.ClientID,
		Sub: sf.config.Username,
		Aud: sf.config.Aud,
		Exp: time.Now().Add(5 * time.Minute).Unix(), // Token valid for 5 minutes
	}
	claimsJSON, _ := json.Marshal(claims)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create signature
	message := headerB64 + "." + claimsB64
	hash := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(rand.Reader, sf.config.PrivateKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}
	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	jwt := message + "." + signatureB64
	return jwt, nil
}

// getAccessToken obtains a new access token using JWT bearer flow
func (sf *SalesforceClient) getAccessToken() error {
	log.Println("📝 Creating JWT assertion...")
	jwt, err := sf.createJWT()
	if err != nil {
		return fmt.Errorf("failed to create JWT: %w", err)
	}

	tokenURL := sf.config.LoginURL + "/services/oauth2/token"
	log.Printf("🔑 Requesting access token from %s", tokenURL)

	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	data.Set("assertion", jwt)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := sf.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Parse error response for better diagnostics
		var errorResp map[string]interface{}
		if json.Unmarshal(body, &errorResp) == nil {
			if errDesc, ok := errorResp["error_description"].(string); ok {
				return sf.interpretTokenError(resp.StatusCode, errDesc)
			}
		}
		return fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	sf.tokenMutex.Lock()
	sf.accessToken = tokenResp.AccessToken
	sf.instanceURL = tokenResp.InstanceURL
	// Salesforce access tokens are typically valid for 2 hours, we'll refresh after 1.5 hours
	sf.tokenExpiry = time.Now().Add(90 * time.Minute)
	sf.tokenMutex.Unlock()

	log.Printf("✅ Successfully obtained access token (expires at %s)", sf.tokenExpiry.Format(time.RFC3339))
	log.Printf("   Instance URL: %s", sf.instanceURL)

	return nil
}

// interpretTokenError provides actionable error messages
func (sf *SalesforceClient) interpretTokenError(statusCode int, errorDesc string) error {
	baseMsg := fmt.Sprintf("Authentication failed (HTTP %d): %s", statusCode, errorDesc)

	// Common issues and how to fix them
	troubleshooting := "\n\nTroubleshooting:\n"

	if strings.Contains(strings.ToLower(errorDesc), "invalid_grant") {
		troubleshooting += "• INVALID_GRANT typically means:\n"
		troubleshooting += "  1. The service user is not authorized for this Connected App\n"
		troubleshooting += "     → Ask IT to pre-approve the user or enable 'Admin approved users are pre-authorized'\n"
		troubleshooting += "  2. The public certificate uploaded to the Connected App doesn't match your private key\n"
		troubleshooting += "     → Verify the certificate was uploaded correctly\n"
		troubleshooting += "  3. The SALESFORCE_USERNAME is incorrect or doesn't exist\n"
		troubleshooting += "     → Verify the username matches exactly (case-sensitive)\n"
	} else if strings.Contains(strings.ToLower(errorDesc), "invalid_client_id") {
		troubleshooting += "• INVALID_CLIENT_ID means:\n"
		troubleshooting += "  → The SALESFORCE_CLIENT_ID (Consumer Key) is incorrect\n"
		troubleshooting += "  → Verify the Client ID from the Connected App settings\n"
	} else if strings.Contains(strings.ToLower(errorDesc), "user hasn't approved") {
		troubleshooting += "• User hasn't approved this Connected App:\n"
		troubleshooting += "  → In the Connected App, enable 'Admin approved users are pre-authorized'\n"
		troubleshooting += "  → OR create a Permission Set/Profile and assign the service user\n"
	}

	troubleshooting += "\nEnvironment check:\n"
	troubleshooting += fmt.Sprintf("• SALESFORCE_CLIENT_ID: %s\n", maskClientID(sf.config.ClientID))
	troubleshooting += fmt.Sprintf("• SALESFORCE_USERNAME: %s\n", sf.config.Username)
	troubleshooting += fmt.Sprintf("• SALESFORCE_AUD (Login URL): %s\n", sf.config.Aud)
	troubleshooting += fmt.Sprintf("• Are you using the correct environment? (production vs sandbox)\n")

	return fmt.Errorf("%s%s", baseMsg, troubleshooting)
}

// ensureValidToken ensures we have a valid access token
func (sf *SalesforceClient) ensureValidToken() error {
	sf.tokenMutex.RLock()
	hasToken := sf.accessToken != ""
	tokenExpired := time.Now().After(sf.tokenExpiry.Add(-5 * time.Minute)) // Refresh 5 min before expiry
	sf.tokenMutex.RUnlock()

	if !hasToken || tokenExpired {
		if tokenExpired {
			log.Println("⏰ Access token is expiring soon, refreshing...")
		}
		return sf.getAccessToken()
	}

	return nil
}

// doWithRetry executes an HTTP request with retry logic and exponential backoff
func (sf *SalesforceClient) doWithRetry(req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= sf.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := sf.config.BackoffBase * time.Duration(math.Pow(2, float64(attempt-1)))
			log.Printf("🔄 Retry attempt %d/%d after %v...", attempt, sf.config.MaxRetries, backoff)
			time.Sleep(backoff)
		}

		resp, err := sf.httpClient.Do(req)
		if err != nil {
			lastErr = err
			log.Printf("⚠️  Request failed (attempt %d/%d): %v", attempt+1, sf.config.MaxRetries+1, err)
			continue
		}

		// Success
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		// Unauthorized - token might be expired
		if resp.StatusCode == http.StatusUnauthorized {
			log.Println("🔐 Received 401 Unauthorized - token may be expired, refreshing...")
			resp.Body.Close()
			if err := sf.getAccessToken(); err != nil {
				return nil, fmt.Errorf("failed to refresh token: %w", err)
			}
			// Update authorization header with new token
			sf.tokenMutex.RLock()
			req.Header.Set("Authorization", "Bearer "+sf.accessToken)
			sf.tokenMutex.RUnlock()
			continue
		}

		// Rate limit - Salesforce returns 429 or sometimes uses specific headers
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == 503 {
			resp.Body.Close()
			log.Printf("⏳ Rate limited (HTTP %d), backing off...", resp.StatusCode)
			lastErr = fmt.Errorf("rate limited (HTTP %d)", resp.StatusCode)
			continue
		}

		// Other 4xx errors are not retryable
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return resp, nil
		}

		// 5xx errors are retryable
		if resp.StatusCode >= 500 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			lastErr = fmt.Errorf("server error (HTTP %d): %s", resp.StatusCode, string(body))
			log.Printf("⚠️  Server error (attempt %d/%d): HTTP %d", attempt+1, sf.config.MaxRetries+1, resp.StatusCode)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("request failed after %d retries: %w", sf.config.MaxRetries, lastErr)
}

// Query executes a SOQL query with retry logic
func (sf *SalesforceClient) Query(soql string) (*QueryResponse, error) {
	// Check cache first
	if sf.config.EnableCache {
		if cached := sf.getFromCache(soql); cached != nil {
			if qr, ok := cached.(*QueryResponse); ok {
				log.Println("📦 Returning cached query result")
				return qr, nil
			}
		}
	}

	// Ensure valid token
	if err := sf.ensureValidToken(); err != nil {
		return nil, fmt.Errorf("failed to obtain access token: %w", err)
	}

	queryURL := fmt.Sprintf("%s/services/data/%s/query", sf.instanceURL, sf.config.APIVersion)

	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create query request: %w", err)
	}

	q := req.URL.Query()
	q.Add("q", soql)
	req.URL.RawQuery = q.Encode()

	sf.tokenMutex.RLock()
	req.Header.Set("Authorization", "Bearer "+sf.accessToken)
	sf.tokenMutex.RUnlock()
	req.Header.Set("Content-Type", "application/json")

	resp, err := sf.doWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query failed with status %d: %s", resp.StatusCode, string(body))
	}

	var queryResp QueryResponse
	if err := json.Unmarshal(body, &queryResp); err != nil {
		return nil, fmt.Errorf("failed to parse query response: %w", err)
	}

	// Cache result
	if sf.config.EnableCache {
		sf.putInCache(soql, &queryResp)
	}

	return &queryResp, nil
}

// GetAccountByID retrieves a single Account by ID
func (sf *SalesforceClient) GetAccountByID(accountID string) (*SalesforceAccount, error) {
	// Check cache first
	cacheKey := "account:" + accountID
	if sf.config.EnableCache {
		if cached := sf.getFromCache(cacheKey); cached != nil {
			if account, ok := cached.(*SalesforceAccount); ok {
				log.Println("📦 Returning cached account")
				return account, nil
			}
		}
	}

	// Ensure valid token
	if err := sf.ensureValidToken(); err != nil {
		return nil, fmt.Errorf("failed to obtain access token: %w", err)
	}

	url := fmt.Sprintf("%s/services/data/%s/sobjects/Account/%s", sf.instanceURL, sf.config.APIVersion, accountID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	sf.tokenMutex.RLock()
	req.Header.Set("Authorization", "Bearer "+sf.accessToken)
	sf.tokenMutex.RUnlock()
	req.Header.Set("Content-Type", "application/json")

	resp, err := sf.doWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get account failed with status %d: %s", resp.StatusCode, string(body))
	}

	var account SalesforceAccount
	if err := json.Unmarshal(body, &account); err != nil {
		return nil, fmt.Errorf("failed to parse account response: %w", err)
	}

	// Cache result
	if sf.config.EnableCache {
		sf.putInCache(cacheKey, &account)
	}

	return &account, nil
}

// Cache helpers
func (sf *SalesforceClient) getFromCache(key string) interface{} {
	sf.cacheMutex.RLock()
	defer sf.cacheMutex.RUnlock()

	entry, exists := sf.cache[key]
	if !exists {
		return nil
	}

	if time.Now().After(entry.expiresAt) {
		return nil
	}

	return entry.data
}

func (sf *SalesforceClient) putInCache(key string, data interface{}) {
	sf.cacheMutex.Lock()
	defer sf.cacheMutex.Unlock()

	sf.cache[key] = &cacheEntry{
		data:      data,
		expiresAt: time.Now().Add(sf.config.CacheTTL),
	}
}

// HealthCheck verifies the integration is working
func (sf *SalesforceClient) HealthCheck() error {
	log.Println("🏥 Running health check...")

	// Try to get an access token
	if err := sf.ensureValidToken(); err != nil {
		return fmt.Errorf("❌ Health check failed - cannot obtain access token: %w", err)
	}

	// Try a simple query to verify API access
	query := "SELECT Id, Name FROM Account LIMIT 1"
	log.Printf("   Testing API access with query: %s", query)

	result, err := sf.Query(query)
	if err != nil {
		return fmt.Errorf("❌ Health check failed - API query failed: %w", err)
	}

	log.Printf("✅ Health check passed! Successfully queried Salesforce.")
	log.Printf("   Found %d records", result.TotalSize)
	log.Printf("   Instance URL: %s", sf.instanceURL)
	log.Printf("   API Version: %s", sf.config.APIVersion)
	log.Printf("   Token expires at: %s", sf.tokenExpiry.Format(time.RFC3339))

	return nil
}

// Utility functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue int64) time.Duration {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return time.Duration(intVal)
		}
	}
	return time.Duration(defaultValue)
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true" || value == "1"
	}
	return defaultValue
}

func maskClientID(clientID string) string {
	if len(clientID) <= 8 {
		return "****"
	}
	return clientID[:4] + "****" + clientID[len(clientID)-4:]
}

// mapAccountToClient converts Salesforce Account to Client format
func mapAccountToClient(account SalesforceAccount) Client {
	tags := []string{}
	if account.Tags != "" {
		tags = strings.Split(account.Tags, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	return Client{
		ID:                     account.ID,
		Name:                   account.Name,
		Status:                 account.CustomerStatus,
		Owner:                  account.Owner.Name,
		CreatedAt:              account.CreatedDate,
		LastActivity:           account.LastActivityDate,
		Tags:                   tags,
		Summary:                account.Description,
		TotalARR:               fmt.Sprintf("$%.0f", account.AnnualRevenue),
		NextRenewalDate:        account.RenewalDate,
		DaysUntilRenewal:       account.DaysUntilRenewal,
		NumberOfUnits:          account.NumberOfUnits,
		CurrentAccountsCreated: account.AccountsCreated,
		CurrentMAU:             account.MonthlyActiveUsers,
		InstallType:            account.InstallType,
		Region:                 account.Region,
		SAOwner:                account.SAOwner,
	}
}

// HTTP Handlers
func readyHandler(w http.ResponseWriter, r *http.Request) {
	// Simple readiness check - just confirm the service is running
	// This doesn't require Salesforce to be configured
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ready",
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if sfClient == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  "Salesforce client not initialized",
		})
		return
	}

	err := sfClient.HealthCheck()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":      "healthy",
		"instance":    sfClient.instanceURL,
		"api_version": sfClient.config.APIVersion,
	})
}

func clientsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := `SELECT Id, Name, Description, CreatedDate, LastModifiedDate, LastActivityDate,
		Owner.Name, Customer_Status__c, Tags__c, Annual_Revenue__c, Renewal_Date__c,
		Days_Until_Renewal__c, Number_of_Units__c, Accounts_Created__c,
		Monthly_Active_Users__c, Install_Type__c, Region__c, SA_Owner__c
		FROM Account
		ORDER BY LastModifiedDate DESC
		LIMIT 100`

	result, err := sfClient.Query(query)
	if err != nil {
		log.Printf("Error querying accounts: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Failed to query accounts: %v", err),
		})
		return
	}

	clients := make([]Client, 0, len(result.Records))
	for _, account := range result.Records {
		clients = append(clients, mapAccountToClient(account))
	}

	json.NewEncoder(w).Encode(clients)
}

func clientHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract account ID from URL path
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/clients/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Account ID is required",
		})
		return
	}

	accountID := parts[0]
	account, err := sfClient.GetAccountByID(accountID)
	if err != nil {
		log.Printf("Error fetching account %s: %v", accountID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": fmt.Sprintf("Failed to fetch account: %v", err),
		})
		return
	}

	client := mapAccountToClient(*account)
	json.NewEncoder(w).Encode(client)
}

func main() {
	log.Println("🚀 Starting Salesforce Integration Service...")

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("❌ Configuration error: %v", err)
	}

	// Initialize Salesforce client
	sfClient = NewSalesforceClient(config)

	// Perform initial health check
	log.Println("🏥 Performing initial health check...")
	if err := sfClient.HealthCheck(); err != nil {
		log.Printf("⚠️  Initial health check failed: %v", err)
		log.Println("   The service will start anyway, but you should fix the configuration.")
		log.Println("   Check the error message above for troubleshooting steps.")
	}

	// Setup HTTP routes
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/ready", readyHandler)
	http.HandleFunc("/clients", clientsHandler)
	http.HandleFunc("/clients/", clientHandler)

	port := getEnvOrDefault("PORT", "8080")
	log.Printf("✅ Salesforce Integration Service started on port %s", port)
	log.Printf("   Ready endpoint: http://localhost:%s/ready", port)
	log.Printf("   Health endpoint: http://localhost:%s/health", port)
	log.Printf("   Clients endpoint: http://localhost:%s/clients", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("❌ Server error: %v", err)
	}
}
