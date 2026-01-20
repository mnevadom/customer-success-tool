package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// SalesforceClient handles authentication and API requests
type SalesforceClient struct {
	InstanceURL   string
	AccessToken   string
	RefreshToken  string
	Username      string
	Password      string
	Token         string
	LoginURL      string
	Connected     bool
	ClientID      string
	ClientSecret  string
}

// LoginResponse represents the OAuth response from Salesforce
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	InstanceURL  string `json:"instance_url"`
	ID           string `json:"id"`
	TokenType    string `json:"token_type"`
	IssuedAt     string `json:"issued_at"`
	Signature    string `json:"signature"`
}

// SalesforceAccount represents a Salesforce Account object
type SalesforceAccount struct {
	ID                     string      `json:"Id"`
	Name                   string      `json:"Name"`
	Description            string      `json:"Description"`
	CreatedDate            string      `json:"CreatedDate"`
	LastModifiedDate       string      `json:"LastModifiedDate"`
	LastActivityDate       string      `json:"LastActivityDate"`
	Owner                  Owner       `json:"Owner"`
	CustomerStatus         string      `json:"Customer_Status__c"`
	Tags                   string      `json:"Tags__c"`
	AnnualRevenue          float64     `json:"Annual_Revenue__c"`
	RenewalDate            string      `json:"Renewal_Date__c"`
	DaysUntilRenewal       int         `json:"Days_Until_Renewal__c"`
	NumberOfUnits          int         `json:"Number_of_Units__c"`
	AccountsCreated        int         `json:"Accounts_Created__c"`
	MonthlyActiveUsers     int         `json:"Monthly_Active_Users__c"`
	InstallType            string      `json:"Install_Type__c"`
	Region                 string      `json:"Region__c"`
	SAOwner                string      `json:"SA_Owner__c"`
}

type Owner struct {
	Name string `json:"Name"`
}

// QueryResponse represents a Salesforce SOQL query response
type QueryResponse struct {
	TotalSize int                  `json:"totalSize"`
	Done      bool                 `json:"done"`
	Records   []SalesforceAccount  `json:"records"`
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

// Initialize Salesforce client
func NewSalesforceClient() *SalesforceClient {
	return &SalesforceClient{
		Username:     os.Getenv("SF_USERNAME"),
		Password:     os.Getenv("SF_PASSWORD"),
		Token:        os.Getenv("SF_SECURITY_TOKEN"),
		LoginURL:     getEnvOrDefault("SF_LOGIN_URL", "https://login.salesforce.com"),
		ClientID:     os.Getenv("SF_CLIENT_ID"),
		ClientSecret: os.Getenv("SF_CLIENT_SECRET"),
		Connected:    false,
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Login authenticates with Salesforce using OAuth Username-Password flow
func (sf *SalesforceClient) Login() error {
	if sf.Username == "" || sf.Password == "" {
		return fmt.Errorf("Salesforce credentials not configured")
	}

	loginURL := sf.LoginURL + "/services/oauth2/token"

	clientID := os.Getenv("SF_CLIENT_ID")
	clientSecret := os.Getenv("SF_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("SF_CLIENT_ID and SF_CLIENT_SECRET must be configured")
	}

	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("username", sf.Username)
	data.Set("password", sf.Password+sf.Token)

	req, err := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create login request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return fmt.Errorf("failed to parse login response: %v", err)
	}

	sf.AccessToken = loginResp.AccessToken
	sf.InstanceURL = loginResp.InstanceURL
	sf.Connected = true

	log.Printf("Successfully authenticated with Salesforce: %s", sf.InstanceURL)
	return nil
}

// Query executes a SOQL query
func (sf *SalesforceClient) Query(soql string) (*QueryResponse, error) {
	if !sf.Connected {
		if err := sf.Login(); err != nil {
			return nil, err
		}
	}

	queryURL := fmt.Sprintf("%s/services/data/v58.0/query", sf.InstanceURL)

	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create query request: %v", err)
	}

	q := req.URL.Query()
	q.Add("q", soql)
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", "Bearer "+sf.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		// Token expired, try to refresh or re-authenticate
		log.Println("Access token expired, attempting to refresh...")
		if sf.RefreshToken != "" {
			// Try refresh token first
			if err := sf.RefreshAccessToken(); err == nil {
				log.Println("Successfully refreshed token, retrying query...")
				return sf.Query(soql)
			}
			log.Printf("Refresh token failed: %v", err)
		}
		// Fall back to password flow if available
		sf.Connected = false
		return sf.Query(soql)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query failed with status %d: %s", resp.StatusCode, string(body))
	}

	var queryResp QueryResponse
	if err := json.Unmarshal(body, &queryResp); err != nil {
		return nil, fmt.Errorf("failed to parse query response: %v", err)
	}

	return &queryResp, nil
}

// GetAccountByID retrieves a single Account by ID
func (sf *SalesforceClient) GetAccountByID(accountID string) (*SalesforceAccount, error) {
	if !sf.Connected {
		if err := sf.Login(); err != nil {
			return nil, err
		}
	}

	url := fmt.Sprintf("%s/services/data/v58.0/sobjects/Account/%s", sf.InstanceURL, accountID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+sf.AccessToken)
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

	if resp.StatusCode == http.StatusUnauthorized {
		sf.Connected = false
		return sf.GetAccountByID(accountID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var account SalesforceAccount
	if err := json.Unmarshal(body, &account); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &account, nil
}

// MapAccountToClient converts a Salesforce Account to Client format
func MapAccountToClient(account SalesforceAccount) Client {
	// Parse tags
	tags := []string{}
	if account.Tags != "" {
		tags = strings.Split(account.Tags, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
	}

	// Format ARR
	totalARR := fmt.Sprintf("$%.2f", account.AnnualRevenue)
	if account.AnnualRevenue == 0 {
		totalARR = "$0.00"
	}

	// Default status
	status := "Active"
	if account.CustomerStatus != "" {
		status = account.CustomerStatus
	}

	return Client{
		ID:                     account.ID,
		Name:                   account.Name,
		Status:                 status,
		Owner:                  account.Owner.Name,
		CreatedAt:              account.CreatedDate,
		LastActivity:           account.LastActivityDate,
		Tags:                   tags,
		Summary:                account.Description,
		TotalARR:               totalARR,
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

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"status":              "healthy",
		"service":             "salesforce-sync",
		"time":                time.Now().Format(time.RFC3339),
		"salesforceConnected": sfClient.Connected,
	}

	json.NewEncoder(w).Encode(response)
}

func testConnectionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if sfClient.Username == "" || sfClient.Password == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Salesforce credentials not configured",
			"mode":    "mock",
		})
		return
	}

	err := sfClient.Login()
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
		"message": "Successfully connected to Salesforce",
		"instanceURL": sfClient.InstanceURL,
	})
}

func getClientsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if sfClient.Username == "" || sfClient.Password == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Salesforce not connected. Please configure SF_USERNAME and SF_PASSWORD.",
			"mode":    "mock",
			"clients": []Client{},
		})
		return
	}

	// SOQL query to get Accounts
	soql := `SELECT Id, Name, Description, CreatedDate, LastModifiedDate, LastActivityDate,
		Owner.Name, Customer_Status__c, Tags__c, Annual_Revenue__c,
		Renewal_Date__c, Days_Until_Renewal__c, Number_of_Units__c,
		Accounts_Created__c, Monthly_Active_Users__c, Install_Type__c,
		Region__c, SA_Owner__c
		FROM Account
		WHERE Customer_Status__c IN ('Active', 'At risk')
		ORDER BY Name
		LIMIT 100`

	queryResp, err := sfClient.Query(soql)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Query failed: %v", err),
			"clients": []Client{},
		})
		return
	}

	// Map to client format
	clients := make([]Client, 0, len(queryResp.Records))
	for _, account := range queryResp.Records {
		clients = append(clients, MapAccountToClient(account))
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"totalCount": queryResp.TotalSize,
		"clients":    clients,
	})
}

func getClientByIDHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract ID from path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Account ID required"})
		return
	}
	accountID := pathParts[3]

	if sfClient.Username == "" || sfClient.Password == "" {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Salesforce not connected",
		})
		return
	}

	account, err := sfClient.GetAccountByID(accountID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to get account: %v", err),
		})
		return
	}

	client := MapAccountToClient(*account)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"client":  client,
	})
}

func syncHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	sfClient.Connected = false
	err := sfClient.Login()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to reconnect: %v", err),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Successfully reconnected to Salesforce",
	})
}

// OAuth 2.0 Web Server Flow - Initiate Authorization
func oauthInitHandler(w http.ResponseWriter, r *http.Request) {
	if sfClient.ClientID == "" {
		http.Error(w, "SF_CLIENT_ID not configured", http.StatusInternalServerError)
		return
	}

	// Get the callback URL from environment or construct it
	callbackURL := os.Getenv("SF_CALLBACK_URL")
	if callbackURL == "" {
		// Construct default callback URL from request
		scheme := "https"
		if r.TLS == nil {
			scheme = "http"
		}
		callbackURL = fmt.Sprintf("%s://%s/oauth/callback", scheme, r.Host)
	}

	// Build authorization URL
	authURL := fmt.Sprintf("%s/services/oauth2/authorize?response_type=code&client_id=%s&redirect_uri=%s",
		sfClient.LoginURL,
		url.QueryEscape(sfClient.ClientID),
		url.QueryEscape(callbackURL),
	)

	log.Printf("Initiating OAuth flow, redirecting to: %s", authURL)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// OAuth 2.0 Web Server Flow - Handle Callback
func oauthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	errorParam := r.URL.Query().Get("error")
	errorDesc := r.URL.Query().Get("error_description")

	if errorParam != "" {
		log.Printf("OAuth error: %s - %s", errorParam, errorDesc)
		http.Error(w, fmt.Sprintf("OAuth error: %s - %s", errorParam, errorDesc), http.StatusBadRequest)
		return
	}

	if code == "" {
		http.Error(w, "No authorization code received", http.StatusBadRequest)
		return
	}

	// Exchange code for access token
	if err := exchangeCodeForToken(code, r.Host); err != nil {
		log.Printf("Failed to exchange code for token: %v", err)
		http.Error(w, fmt.Sprintf("Failed to exchange code for token: %v", err), http.StatusInternalServerError)
		return
	}

	// Success page
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Salesforce Connected</title>
			<style>
				body { font-family: Arial, sans-serif; text-align: center; padding: 50px; background: #f0f0f0; }
				.success { background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); max-width: 500px; margin: 0 auto; }
				.success h1 { color: #00a86b; }
				.btn { display: inline-block; padding: 10px 20px; background: #0070d2; color: white; text-decoration: none; border-radius: 5px; margin-top: 20px; }
			</style>
		</head>
		<body>
			<div class="success">
				<h1>✅ Successfully Connected to Salesforce!</h1>
				<p>Your Salesforce integration is now active.</p>
				<p>Instance URL: <strong>%s</strong></p>
				<a href="/" class="btn">Return to Dashboard</a>
			</div>
		</body>
		</html>
	`, sfClient.InstanceURL)
}

// Exchange authorization code for access token
func exchangeCodeForToken(code, host string) error {
	if sfClient.ClientID == "" || sfClient.ClientSecret == "" {
		return fmt.Errorf("SF_CLIENT_ID and SF_CLIENT_SECRET must be configured")
	}

	// Get callback URL
	callbackURL := os.Getenv("SF_CALLBACK_URL")
	if callbackURL == "" {
		callbackURL = fmt.Sprintf("https://%s/oauth/callback", host)
	}

	tokenURL := sfClient.LoginURL + "/services/oauth2/token"

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", sfClient.ClientID)
	data.Set("client_secret", sfClient.ClientSecret)
	data.Set("redirect_uri", callbackURL)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("token request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return fmt.Errorf("failed to parse token response: %v", err)
	}

	sfClient.AccessToken = loginResp.AccessToken
	sfClient.RefreshToken = loginResp.RefreshToken
	sfClient.InstanceURL = loginResp.InstanceURL
	sfClient.Connected = true

	log.Printf("✅ Successfully authenticated with Salesforce via OAuth: %s", sfClient.InstanceURL)
	log.Printf("Access token obtained, refresh token: %s", sfClient.RefreshToken[:10]+"...")

	return nil
}

// Refresh access token using refresh token
func (sf *SalesforceClient) RefreshAccessToken() error {
	if sf.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	tokenURL := sf.LoginURL + "/services/oauth2/token"

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", sf.RefreshToken)
	data.Set("client_id", sf.ClientID)
	data.Set("client_secret", sf.ClientSecret)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create refresh request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("refresh request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return fmt.Errorf("failed to parse refresh response: %v", err)
	}

	sf.AccessToken = loginResp.AccessToken
	// Refresh token might be updated, update if provided
	if loginResp.RefreshToken != "" {
		sf.RefreshToken = loginResp.RefreshToken
	}

	log.Printf("✅ Successfully refreshed Salesforce access token")
	return nil
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
	// Initialize Salesforce client
	sfClient = NewSalesforceClient()

	// Try to connect on startup if credentials are available
	if sfClient.Username != "" && sfClient.Password != "" {
		if err := sfClient.Login(); err != nil {
			log.Printf("Warning: Failed to connect to Salesforce on startup: %v", err)
			log.Println("Service will run in mock mode until credentials are configured")
		}
	} else {
		log.Println("Salesforce credentials not configured. Running in mock mode.")
	}

	// Setup routes
	http.HandleFunc("/health", enableCORS(healthHandler))
	http.HandleFunc("/api/test-connection", enableCORS(testConnectionHandler))
	http.HandleFunc("/api/clients", enableCORS(getClientsHandler))
	http.HandleFunc("/api/clients/", enableCORS(getClientByIDHandler))
	http.HandleFunc("/api/sync", enableCORS(syncHandler))

	// OAuth 2.0 Web Server Flow routes
	http.HandleFunc("/oauth/authorize", oauthInitHandler)
	http.HandleFunc("/oauth/callback", oauthCallbackHandler)

	port := getEnvOrDefault("PORT", "9000")
	log.Printf("Starting Salesforce Sync service on port %s", port)
	log.Printf("Login URL: %s", sfClient.LoginURL)

	if sfClient.ClientID != "" {
		log.Printf("OAuth 2.0 Web Server Flow enabled")
		log.Printf("To connect, visit: https://salesforce-sync-agent-87kkzgw4nmq4.demo.okteto.dev/oauth/authorize")
		log.Printf("Callback URL for Salesforce Connected App: https://salesforce-sync-agent-87kkzgw4nmq4.demo.okteto.dev/oauth/callback")
	}

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
