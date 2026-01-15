package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/rs/cors"
)

// Client represents a customer in the system
type Client struct {
	ID                    string    `json:"id"`
	Name                  string    `json:"name"`
	Status                string    `json:"status"`
	Owner                 string    `json:"owner"`
	CreatedAt             time.Time `json:"createdAt"`
	LastActivity          time.Time `json:"lastActivity"`
	Tags                  []string  `json:"tags"`
	Summary               string    `json:"summary"`
	TotalARR              string    `json:"totalARR"`
	NextRenewalDate       time.Time `json:"nextRenewalDate"`
	DaysUntilRenewal      int       `json:"daysUntilRenewal"`
	NumberOfUnits         int       `json:"numberOfUnits"`
	CurrentAccountsCreated int       `json:"currentAccountsCreated"`
	CurrentMAU            int       `json:"currentMAU"`
	InstallType           string    `json:"installType"`
	Region                string    `json:"region"`
	SAOwner               string    `json:"saOwner"`
}

// Widget represents a dashboard widget
type Widget struct {
	ID    string                 `json:"id"`
	Title string                 `json:"title"`
	Type  string                 `json:"type"`
	Data  map[string]interface{} `json:"data"`
}

// Dashboard represents a collection of widgets
type Dashboard struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Widgets []Widget `json:"widgets"`
}

// ThenaRequest represents a Thena webhook request
type ThenaRequest struct {
	ID              string    `json:"id"`
	RequestID       int       `json:"requestId"`
	ThenaID         string    `json:"thenaId"`
	EventID         string    `json:"eventId"`
	Status          string    `json:"status"`
	SubStatus       string    `json:"subStatus"`
	SubStatusName   string    `json:"subStatusName"`
	SubStatusDesc   string    `json:"subStatusDesc"`
	CustomerName    string    `json:"customerName"`
	CRMID           string    `json:"crmId"`
	CRMName         string    `json:"crmName"`
	ChannelID       string    `json:"channelId"`
	ChannelName     string    `json:"channelName"`
	Permalink       string    `json:"permalink"`
	RequestLink     string    `json:"requestLink"`
	ThenaURL        string    `json:"thenaUrl"`
	AssignedToID    string    `json:"assignedToId"`
	AssignedToName  string    `json:"assignedToName"`
	AssignedToEmail string    `json:"assignedToEmail"`
	RequestorID     string    `json:"requestorId"`
	RequestorName   string    `json:"requestorName"`
	RequestorEmail  string    `json:"requestorEmail"`
	CreatedAt       string    `json:"createdAt"`
	UpdatedAt       string    `json:"updatedAt"`
	ReplyCount      int       `json:"replyCount"`
	Description     string    `json:"description"`
	ReceivedAt      time.Time `json:"receivedAt"`
}

// Mock data store
var mockClients = []Client{
	{
		ID: "client-1", Name: "iCapital Network", Status: "Active", Owner: "Alice Johnson",
		CreatedAt: time.Date(2023, 3, 15, 10, 0, 0, 0, time.UTC), LastActivity: time.Now().Add(-3 * time.Hour),
		Tags: []string{"enterprise", "finance"}, Summary: "Leading wealth management platform. Strong engagement with platform.",
		TotalARR: "$175,000.00", NextRenewalDate: time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: -11,
		NumberOfUnits: 300, CurrentAccountsCreated: 71, CurrentMAU: 71, InstallType: "Self-Hosted", Region: "Europe", SAOwner: "Mario",
	},
	{
		ID: "client-2", Name: "ServiceTitan Inc", Status: "Active", Owner: "Bob Martinez",
		CreatedAt: time.Date(2022, 6, 20, 14, 30, 0, 0, time.UTC), LastActivity: time.Now().Add(-1 * time.Hour),
		Tags: []string{"enterprise", "SaaS"}, Summary: "Home services software leader. Planning expansion to 3 new regions.",
		TotalARR: "$89,400.00", NextRenewalDate: time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 20,
		NumberOfUnits: 50, CurrentAccountsCreated: 109, CurrentMAU: 10, InstallType: "Self-Hosted", Region: "USA", SAOwner: "Jonathan",
	},
	{
		ID: "client-3", Name: "Mercadona", Status: "Active", Owner: "Carlos Rivera",
		CreatedAt: time.Date(2023, 9, 5, 9, 0, 0, 0, time.UTC), LastActivity: time.Now().Add(-6 * time.Hour),
		Tags: []string{"enterprise", "retail"}, Summary: "Major European retailer. Expanding digital transformation initiatives.",
		TotalARR: "$68,310.00", NextRenewalDate: time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 79,
		NumberOfUnits: 50, CurrentAccountsCreated: 31, CurrentMAU: 31, InstallType: "Self-Hosted", Region: "Europe", SAOwner: "Mario",
	},
	{
		ID: "client-4", Name: "Ruggable", Status: "At risk", Owner: "Diana Foster",
		CreatedAt: time.Date(2024, 1, 10, 11, 0, 0, 0, time.UTC), LastActivity: time.Now().Add(-15 * 24 * time.Hour),
		Tags: []string{"e-commerce", "SMB"}, Summary: "Machine-washable rug company. Support tickets increasing, engagement declining.",
		TotalARR: "$12,414.60", NextRenewalDate: time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 109,
		NumberOfUnits: 10, CurrentAccountsCreated: 6, CurrentMAU: 3, InstallType: "Self-Hosted", Region: "USA", SAOwner: "Jonathan",
	},
	{
		ID: "client-5", Name: "Replicated", Status: "Active", Owner: "Alice Johnson",
		CreatedAt: time.Date(2023, 4, 22, 15, 30, 0, 0, time.UTC), LastActivity: time.Now().Add(-2 * time.Hour),
		Tags: []string{"enterprise", "DevOps"}, Summary: "Kubernetes native platform. Active community contributor and advocate.",
		TotalARR: "$27,324.00", NextRenewalDate: time.Date(2026, 5, 6, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 115,
		NumberOfUnits: 20, CurrentAccountsCreated: 30, CurrentMAU: 25, InstallType: "Managed Service", Region: "USA", SAOwner: "Jonathan",
	},
	{
		ID: "client-6", Name: "Visible Ideas", Status: "Active", Owner: "Emily Thompson",
		CreatedAt: time.Date(2024, 2, 14, 10, 0, 0, 0, time.UTC), LastActivity: time.Now().Add(-5 * time.Hour),
		Tags: []string{"agency", "SMB"}, Summary: "Creative agency focused on brand strategy. Recently upgraded to premium tier.",
		TotalARR: "$34,006.50", NextRenewalDate: time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 140,
		NumberOfUnits: 25, CurrentAccountsCreated: 19, CurrentMAU: 14, InstallType: "Self-Hosted", Region: "Europe", SAOwner: "Mario",
	},
	{
		ID: "client-7", Name: "Quickwork Technologies Private Limited", Status: "Active", Owner: "Raj Patel",
		CreatedAt: time.Date(2023, 7, 8, 9, 15, 0, 0, time.UTC), LastActivity: time.Now().Add(-4 * time.Hour),
		Tags: []string{"enterprise", "automation"}, Summary: "Integration platform provider. Exploring partnership opportunities.",
		TotalARR: "$2,352.00", NextRenewalDate: time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 155,
		NumberOfUnits: 4, CurrentAccountsCreated: 5, CurrentMAU: 5, InstallType: "Self-Hosted", Region: "India", SAOwner: "Mario",
	},
	{
		ID: "client-8", Name: "Fenix Media Limited T/A Pulsar", Status: "Active", Owner: "Oliver Smith",
		CreatedAt: time.Date(2023, 11, 3, 14, 0, 0, 0, time.UTC), LastActivity: time.Now().Add(-8 * time.Hour),
		Tags: []string{"media", "analytics"}, Summary: "Social media intelligence platform. Strong product-market fit.",
		TotalARR: "$52,614.20", NextRenewalDate: time.Date(2026, 6, 17, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 157,
		NumberOfUnits: 11, CurrentAccountsCreated: 24, CurrentMAU: 10, InstallType: "SaaS", Region: "Europe", SAOwner: "Mario",
	},
	{
		ID: "client-9", Name: "CoverWallet, Inc.", Status: "At risk", Owner: "Bob Martinez",
		CreatedAt: time.Date(2024, 3, 25, 11, 30, 0, 0, time.UTC), LastActivity: time.Now().Add(-20 * 24 * time.Hour),
		Tags: []string{"insurance", "fintech"}, Summary: "Small business insurance platform. Key stakeholder left company recently.",
		TotalARR: "$177,012.00", NextRenewalDate: time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 170,
		NumberOfUnits: 90, CurrentAccountsCreated: 70, CurrentMAU: 42, InstallType: "Self-Hosted", Region: "Europe", SAOwner: "Mario",
	},
	{
		ID: "client-10", Name: "Wave Financial USA", Status: "Active", Owner: "Diana Foster",
		CreatedAt: time.Date(2023, 5, 17, 10, 45, 0, 0, time.UTC), LastActivity: time.Now().Add(-1 * time.Hour),
		Tags: []string{"fintech", "SMB"}, Summary: "Accounting software for small businesses. Consistent growth trajectory.",
		TotalARR: "$129,789.00", NextRenewalDate: time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 170,
		NumberOfUnits: 95, CurrentAccountsCreated: 90, CurrentMAU: 90, InstallType: "Self-Hosted", Region: "Canada", SAOwner: "Jonathan",
	},
	{
		ID: "client-11", Name: "TELESCOPE TECHNOLOGY LIMITED", Status: "Active", Owner: "Emily Thompson",
		CreatedAt: time.Date(2024, 1, 30, 13, 0, 0, 0, time.UTC), LastActivity: time.Now().Add(-3 * time.Hour),
		Tags: []string{"enterprise", "observatory"}, Summary: "Advanced telescope control systems. Recent successful deployment.",
		TotalARR: "$11,880.00", NextRenewalDate: time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 170,
		NumberOfUnits: 10, CurrentAccountsCreated: 13, CurrentMAU: 10, InstallType: "Managed Service", Region: "Europe", SAOwner: "Mario",
	},
	{
		ID: "client-12", Name: "Yotpo Ltd.", Status: "Active", Owner: "Alice Johnson",
		CreatedAt: time.Date(2022, 8, 12, 9, 30, 0, 0, time.UTC), LastActivity: time.Now().Add(-2 * time.Hour),
		Tags: []string{"enterprise", "e-commerce"}, Summary: "E-commerce marketing platform. Long-term strategic partner.",
		TotalARR: "$70,000.00", NextRenewalDate: time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 178,
		NumberOfUnits: 80, CurrentAccountsCreated: 137, CurrentMAU: 60, InstallType: "Managed Service", Region: "Israel", SAOwner: "Mario",
	},
	{
		ID: "client-13", Name: "Fiverr International Ltd", Status: "Active", Owner: "Carlos Rivera",
		CreatedAt: time.Date(2023, 2, 28, 15, 0, 0, 0, time.UTC), LastActivity: time.Now().Add(-4 * time.Hour),
		Tags: []string{"enterprise", "marketplace"}, Summary: "Global freelance marketplace. Scaling infrastructure for growth.",
		TotalARR: "$173,257.20", NextRenewalDate: time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 201,
		NumberOfUnits: 170, CurrentAccountsCreated: 125, CurrentMAU: 121, InstallType: "Self-Hosted", Region: "Israel", SAOwner: "Mario",
	},
	{
		ID: "client-14", Name: "World 50", Status: "Active", Owner: "Oliver Smith",
		CreatedAt: time.Date(2023, 10, 19, 11, 15, 0, 0, time.UTC), LastActivity: time.Now().Add(-7 * time.Hour),
		Tags: []string{"enterprise", "community"}, Summary: "Executive networking community. High engagement among C-suite members.",
		TotalARR: "$16,431.00", NextRenewalDate: time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 201,
		NumberOfUnits: 5, CurrentAccountsCreated: 10, CurrentMAU: 10, InstallType: "Managed Service", Region: "USA", SAOwner: "Jonathan",
	},
	{
		ID: "client-15", Name: "Exivity", Status: "Active", Owner: "Raj Patel",
		CreatedAt: time.Date(2024, 4, 5, 10, 0, 0, 0, time.UTC), LastActivity: time.Now().Add(-5 * time.Hour),
		Tags: []string{"enterprise", "analytics"}, Summary: "IT cost management and chargeback platform. Positive ROI demonstrated.",
		TotalARR: "$10,692.00", NextRenewalDate: time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 220,
		NumberOfUnits: 9, CurrentAccountsCreated: 7, CurrentMAU: 8, InstallType: "Self-Hosted", Region: "USA", SAOwner: "Jonathan",
	},
	{
		ID: "client-16", Name: "Flexera Software LLC", Status: "Active", Owner: "Bob Martinez",
		CreatedAt: time.Date(2022, 11, 8, 14, 30, 0, 0, time.UTC), LastActivity: time.Now().Add(-2 * time.Hour),
		Tags: []string{"enterprise", "software"}, Summary: "Software asset management leader. Multi-year agreement in place.",
		TotalARR: "$26,820.00", NextRenewalDate: time.Date(2026, 8, 24, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 225,
		NumberOfUnits: 15, CurrentAccountsCreated: 2, CurrentMAU: 2, InstallType: "Self-Hosted", Region: "Israel", SAOwner: "Mario",
	},
	{
		ID: "client-17", Name: "Sirona Medical", Status: "At risk", Owner: "Diana Foster",
		CreatedAt: time.Date(2024, 5, 12, 9, 0, 0, 0, time.UTC), LastActivity: time.Now().Add(-18 * 24 * time.Hour),
		Tags: []string{"healthcare", "SMB"}, Summary: "Medical imaging platform. Budget constraints reported for next quarter.",
		TotalARR: "$30,124.71", NextRenewalDate: time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 232,
		NumberOfUnits: 20, CurrentAccountsCreated: 20, CurrentMAU: 3, InstallType: "Self-Hosted", Region: "USA", SAOwner: "Jonathan",
	},
	{
		ID: "client-18", Name: "Kudosity", Status: "Active", Owner: "Emily Thompson",
		CreatedAt: time.Date(2023, 12, 1, 11, 30, 0, 0, time.UTC), LastActivity: time.Now().Add(-6 * time.Hour),
		Tags: []string{"HR", "SMB"}, Summary: "Employee recognition platform. Recently featured in HR Tech publication.",
		TotalARR: "$31,384.00", NextRenewalDate: time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 232,
		NumberOfUnits: 8, CurrentAccountsCreated: 15, CurrentMAU: 11, InstallType: "SaaS", Region: "Australia", SAOwner: "Ramiro",
	},
	{
		ID: "client-19", Name: "Lawrence Berkeley National Laboratory", Status: "Active", Owner: "Alice Johnson",
		CreatedAt: time.Date(2023, 6, 15, 10, 0, 0, 0, time.UTC), LastActivity: time.Now().Add(-3 * time.Hour),
		Tags: []string{"government", "research"}, Summary: "National research facility. Long procurement cycles but stable funding.",
		TotalARR: "$35,760.00", NextRenewalDate: time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 276,
		NumberOfUnits: 20, CurrentAccountsCreated: 15, CurrentMAU: 8, InstallType: "Self-Hosted", Region: "USA", SAOwner: "Jonathan",
	},
	{
		ID: "client-20", Name: "Acerta", Status: "Active", Owner: "Carlos Rivera",
		CreatedAt: time.Date(2024, 2, 20, 13, 45, 0, 0, time.UTC), LastActivity: time.Now().Add(-4 * time.Hour),
		Tags: []string{"automotive", "AI"}, Summary: "AI-powered quality control for manufacturing. Expanding to new verticals.",
		TotalARR: "$29,900.00", NextRenewalDate: time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 293,
		NumberOfUnits: 20, CurrentAccountsCreated: 17, CurrentMAU: 17, InstallType: "Self-Hosted", Region: "Europe", SAOwner: "Mario",
	},
	{
		ID: "client-21", Name: "Catamorphic Co. dba LaunchDarkly", Status: "Active", Owner: "Oliver Smith",
		CreatedAt: time.Date(2022, 9, 10, 15, 0, 0, 0, time.UTC), LastActivity: time.Now().Add(-1 * time.Hour),
		Tags: []string{"enterprise", "DevOps"}, Summary: "Feature management platform. Strong developer community engagement.",
		TotalARR: "$70,200.00", NextRenewalDate: time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 323,
		NumberOfUnits: 75, CurrentAccountsCreated: 64, CurrentMAU: 42, InstallType: "Managed Service", Region: "USA", SAOwner: "Jonathan",
	},
	{
		ID: "client-22", Name: "Lema Labs LTD", Status: "Active", Owner: "Raj Patel",
		CreatedAt: time.Date(2024, 3, 8, 10, 30, 0, 0, time.UTC), LastActivity: time.Now().Add(-5 * time.Hour),
		Tags: []string{"startup", "AI"}, Summary: "AI research lab. Early stage but high growth potential.",
		TotalARR: "$9,410.38", NextRenewalDate: time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 323,
		NumberOfUnits: 8, CurrentAccountsCreated: 13, CurrentMAU: 12, InstallType: "Self-Hosted", Region: "Israel", SAOwner: "Mario",
	},
	{
		ID: "client-23", Name: "Hinge Health, Inc.", Status: "Active", Owner: "Diana Foster",
		CreatedAt: time.Date(2023, 8, 25, 9, 15, 0, 0, time.UTC), LastActivity: time.Now().Add(-2 * time.Hour),
		Tags: []string{"healthcare", "enterprise"}, Summary: "Digital musculoskeletal clinic. Rapid user adoption and strong outcomes.",
		TotalARR: "$207,900.00", NextRenewalDate: time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 354,
		NumberOfUnits: 250, CurrentAccountsCreated: 201, CurrentMAU: 201, InstallType: "Managed Service", Region: "India", SAOwner: "Mario",
	},
	{
		ID: "client-24", Name: "Hello Doctor LTD", Status: "Active", Owner: "Emily Thompson",
		CreatedAt: time.Date(2024, 1, 18, 14, 0, 0, 0, time.UTC), LastActivity: time.Now().Add(-7 * time.Hour),
		Tags: []string{"healthcare", "telemedicine"}, Summary: "Telemedicine platform expanding across African markets.",
		TotalARR: "$27,834.84", NextRenewalDate: time.Date(2027, 1, 31, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 385,
		NumberOfUnits: 33, CurrentAccountsCreated: 39, CurrentMAU: 37, InstallType: "Self-Hosted", Region: "Israel", SAOwner: "Mario",
	},
	{
		ID: "client-25", Name: "monday.com", Status: "Active", Owner: "Bob Martinez",
		CreatedAt: time.Date(2022, 7, 5, 11, 0, 0, 0, time.UTC), LastActivity: time.Now().Add(-1 * time.Hour),
		Tags: []string{"enterprise", "productivity"}, Summary: "Work operating system. Strategic partnership with co-marketing initiatives.",
		TotalARR: "$681,100.00", NextRenewalDate: time.Date(2027, 3, 31, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 444,
		NumberOfUnits: 650, CurrentAccountsCreated: 659, CurrentMAU: 659, InstallType: "Managed Service", Region: "Israel", SAOwner: "Mario",
	},
	{
		ID: "client-26", Name: "Upwork", Status: "Active", Owner: "Alice Johnson",
		CreatedAt: time.Date(2023, 3, 22, 10, 30, 0, 0, time.UTC), LastActivity: time.Now().Add(-3 * time.Hour),
		Tags: []string{"enterprise", "marketplace"}, Summary: "Global freelancing platform. Exploring API integration opportunities.",
		TotalARR: "$350,000.00", NextRenewalDate: time.Date(2027, 8, 31, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 597,
		NumberOfUnits: 400, CurrentAccountsCreated: 0, CurrentMAU: 0, InstallType: "Self-Hosted", Region: "USA", SAOwner: "Jonathan",
	},
	{
		ID: "client-27", Name: "SURF", Status: "Active", Owner: "Carlos Rivera",
		CreatedAt: time.Date(2023, 11, 14, 9, 0, 0, 0, time.UTC), LastActivity: time.Now().Add(-4 * time.Hour),
		Tags: []string{"education", "research"}, Summary: "Dutch national research infrastructure. Strong collaboration on innovation projects.",
		TotalARR: "$29,788.00", NextRenewalDate: time.Date(2028, 3, 31, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 810,
		NumberOfUnits: 20, CurrentAccountsCreated: 16, CurrentMAU: 16, InstallType: "Self-Hosted", Region: "Europe", SAOwner: "Mario",
	},
	{
		ID: "client-28", Name: "Sonar Software", Status: "Active", Owner: "Oliver Smith",
		CreatedAt: time.Date(2024, 4, 10, 13, 30, 0, 0, time.UTC), LastActivity: time.Now().Add(-6 * time.Hour),
		Tags: []string{"ISP", "telecom"}, Summary: "ISP management platform. Recently won industry award for innovation.",
		TotalARR: "$45,594.00", NextRenewalDate: time.Date(2028, 12, 2, 0, 0, 0, 0, time.UTC), DaysUntilRenewal: 1056,
		NumberOfUnits: 25, CurrentAccountsCreated: 25, CurrentMAU: 23, InstallType: "Self-Hosted", Region: "Canada", SAOwner: "Jonathan",
	},
}

// In-memory store for Thena requests (with mutex for thread safety)
var (
	thenaRequests      = []ThenaRequest{}
	thenaRequestsMutex = &sync.Mutex{}
)

var mockDashboards = []Dashboard{
	{
		ID:   "dashboard-1",
		Name: "Customer Health Overview",
		Widgets: []Widget{
			{
				ID:    "widget-1",
				Title: "Active Clients",
				Type:  "KPI",
				Data: map[string]interface{}{
					"value": 42,
					"trend": "+5%",
				},
			},
			{
				ID:    "widget-2",
				Title: "At Risk Clients",
				Type:  "KPI",
				Data: map[string]interface{}{
					"value": 3,
					"trend": "-2",
				},
			},
			{
				ID:    "widget-3",
				Title: "Monthly Revenue",
				Type:  "chart",
				Data: map[string]interface{}{
					"chartType": "line",
					"value":     "$125,000",
				},
			},
			{
				ID:    "widget-4",
				Title: "Recent Activity",
				Type:  "text",
				Data: map[string]interface{}{
					"content": "5 client meetings scheduled this week. 2 renewals pending.",
				},
			},
		},
	},
	{
		ID:   "dashboard-2",
		Name: "Sales Pipeline",
		Widgets: []Widget{
			{
				ID:    "widget-5",
				Title: "Pipeline Value",
				Type:  "KPI",
				Data: map[string]interface{}{
					"value": "$450,000",
					"trend": "+12%",
				},
			},
			{
				ID:    "widget-6",
				Title: "Conversion Rate",
				Type:  "KPI",
				Data: map[string]interface{}{
					"value": "28%",
					"trend": "+3%",
				},
			},
		},
	},
	{
		ID:   "dashboard-3",
		Name: "Customer Tasks Board",
		Widgets: []Widget{},
	},
	{
		ID:   "dashboard-thena",
		Name: "Thena Board",
		Widgets: []Widget{},
	},
}

// Helper to find client by ID
func findClientByID(id string) *Client {
	for _, client := range mockClients {
		if client.ID == id {
			return &client
		}
	}
	return nil
}

// Helper to find dashboard by ID
func findDashboardByID(id string) *Dashboard {
	for _, dashboard := range mockDashboards {
		if dashboard.ID == id {
			return &dashboard
		}
	}
	return nil
}

// Handler for receiving webhook data from thena-sync
func thenaWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		log.Printf("Error decoding webhook payload: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Extract fields from payload
	thenaReq := ThenaRequest{
		ID:         fmt.Sprintf("thena-%d", time.Now().UnixNano()),
		ReceivedAt: time.Now(),
	}

	// Helper to safely extract string
	getString := func(key string) string {
		if val, ok := payload[key]; ok && val != nil {
			return fmt.Sprintf("%v", val)
		}
		return ""
	}

	// Helper to safely extract int
	getInt := func(key string) int {
		if val, ok := payload[key]; ok && val != nil {
			switch v := val.(type) {
			case float64:
				return int(v)
			case int:
				return v
			}
		}
		return 0
	}

	// Helper to extract nested object string
	getNestedString := func(obj map[string]interface{}, key string) string {
		if val, ok := obj[key]; ok && val != nil {
			return fmt.Sprintf("%v", val)
		}
		return ""
	}

	thenaReq.RequestID = getInt("requestId")
	thenaReq.ThenaID = getString("thenaId")
	thenaReq.EventID = getString("eventId")
	thenaReq.Status = getString("status")
	thenaReq.SubStatus = getString("subStatus")
	thenaReq.SubStatusName = getString("subStatusName")
	thenaReq.SubStatusDesc = getString("subStatusDesc")
	thenaReq.CustomerName = getString("customerName")
	thenaReq.CRMID = getString("crmId")
	thenaReq.CRMName = getString("crmName")
	thenaReq.ChannelID = getString("channelId")
	thenaReq.ChannelName = getString("channelName")
	thenaReq.Permalink = getString("permalink")
	thenaReq.RequestLink = getString("requestLink")
	thenaReq.ThenaURL = getString("thenaUrl")
	thenaReq.CreatedAt = getString("createdAt")
	thenaReq.UpdatedAt = getString("updatedAt")
	thenaReq.ReplyCount = getInt("replyCount")
	thenaReq.Description = getString("description")

	// Extract assignedTo
	if assignedTo, ok := payload["assignedTo"].(map[string]interface{}); ok {
		thenaReq.AssignedToID = getNestedString(assignedTo, "id")
		thenaReq.AssignedToName = getNestedString(assignedTo, "name")
		thenaReq.AssignedToEmail = getNestedString(assignedTo, "email")
	}

	// Extract requestor
	if requestor, ok := payload["requestor"].(map[string]interface{}); ok {
		thenaReq.RequestorID = getNestedString(requestor, "id")
		thenaReq.RequestorName = getNestedString(requestor, "name")
		thenaReq.RequestorEmail = getNestedString(requestor, "email")
	}

	// Store the request
	thenaRequestsMutex.Lock()
	thenaRequests = append(thenaRequests, thenaReq)
	thenaRequestsMutex.Unlock()

	log.Printf("✅ Stored Thena request: requestId=%d customer=%s status=%s",
		thenaReq.RequestID, thenaReq.CustomerName, thenaReq.Status)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	// Define GraphQL types
	widgetType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Widget",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"title": &graphql.Field{
				Type: graphql.String,
			},
			"type": &graphql.Field{
				Type: graphql.String,
			},
			"data": &graphql.Field{
				Type: graphql.NewScalar(graphql.ScalarConfig{
					Name: "JSON",
					Serialize: func(value interface{}) interface{} {
						return value
					},
				}),
			},
		},
	})

	dashboardType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Dashboard",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"widgets": &graphql.Field{
				Type: graphql.NewList(widgetType),
			},
		},
	})

	clientType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Client",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"status": &graphql.Field{
				Type: graphql.String,
			},
			"owner": &graphql.Field{
				Type: graphql.String,
			},
			"createdAt": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if client, ok := p.Source.(Client); ok {
						return client.CreatedAt.Format(time.RFC3339), nil
					}
					return nil, nil
				},
			},
			"lastActivity": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if client, ok := p.Source.(Client); ok {
						return client.LastActivity.Format(time.RFC3339), nil
					}
					return nil, nil
				},
			},
			"tags": &graphql.Field{
				Type: graphql.NewList(graphql.String),
			},
			"summary": &graphql.Field{
				Type: graphql.String,
			},
			"totalARR": &graphql.Field{
				Type: graphql.String,
			},
			"nextRenewalDate": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if client, ok := p.Source.(Client); ok {
						return client.NextRenewalDate.Format(time.RFC3339), nil
					}
					return nil, nil
				},
			},
			"daysUntilRenewal": &graphql.Field{
				Type: graphql.Int,
			},
			"numberOfUnits": &graphql.Field{
				Type: graphql.Int,
			},
			"currentAccountsCreated": &graphql.Field{
				Type: graphql.Int,
			},
			"currentMAU": &graphql.Field{
				Type: graphql.Int,
			},
			"installType": &graphql.Field{
				Type: graphql.String,
			},
			"region": &graphql.Field{
				Type: graphql.String,
			},
			"saOwner": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

	thenaRequestType := graphql.NewObject(graphql.ObjectConfig{
		Name: "ThenaRequest",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"requestId": &graphql.Field{
				Type: graphql.Int,
			},
			"thenaId": &graphql.Field{
				Type: graphql.String,
			},
			"eventId": &graphql.Field{
				Type: graphql.String,
			},
			"status": &graphql.Field{
				Type: graphql.String,
			},
			"subStatus": &graphql.Field{
				Type: graphql.String,
			},
			"subStatusName": &graphql.Field{
				Type: graphql.String,
			},
			"subStatusDesc": &graphql.Field{
				Type: graphql.String,
			},
			"customerName": &graphql.Field{
				Type: graphql.String,
			},
			"crmId": &graphql.Field{
				Type: graphql.String,
			},
			"crmName": &graphql.Field{
				Type: graphql.String,
			},
			"channelId": &graphql.Field{
				Type: graphql.String,
			},
			"channelName": &graphql.Field{
				Type: graphql.String,
			},
			"permalink": &graphql.Field{
				Type: graphql.String,
			},
			"requestLink": &graphql.Field{
				Type: graphql.String,
			},
			"thenaUrl": &graphql.Field{
				Type: graphql.String,
			},
			"assignedToId": &graphql.Field{
				Type: graphql.String,
			},
			"assignedToName": &graphql.Field{
				Type: graphql.String,
			},
			"assignedToEmail": &graphql.Field{
				Type: graphql.String,
			},
			"requestorId": &graphql.Field{
				Type: graphql.String,
			},
			"requestorName": &graphql.Field{
				Type: graphql.String,
			},
			"requestorEmail": &graphql.Field{
				Type: graphql.String,
			},
			"createdAt": &graphql.Field{
				Type: graphql.String,
			},
			"updatedAt": &graphql.Field{
				Type: graphql.String,
			},
			"replyCount": &graphql.Field{
				Type: graphql.Int,
			},
			"description": &graphql.Field{
				Type: graphql.String,
			},
			"receivedAt": &graphql.Field{
				Type: graphql.String,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if req, ok := p.Source.(ThenaRequest); ok {
						return req.ReceivedAt.Format(time.RFC3339), nil
					}
					return nil, nil
				},
			},
		},
	})

	// Define root query
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"clients": &graphql.Field{
				Type:        graphql.NewList(clientType),
				Description: "Get list of all clients",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					log.Println("Query: clients - returning all clients")
					return mockClients, nil
				},
			},
			"client": &graphql.Field{
				Type:        clientType,
				Description: "Get a single client by ID",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["id"].(string)
					if !ok {
						return nil, fmt.Errorf("invalid id argument")
					}
					log.Printf("Query: client(id=%s)\n", id)
					client := findClientByID(id)
					if client == nil {
						log.Printf("Client not found: %s\n", id)
					}
					return client, nil
				},
			},
			"dashboards": &graphql.Field{
				Type:        graphql.NewList(dashboardType),
				Description: "Get list of all dashboards",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					log.Println("Query: dashboards - returning all dashboards")
					return mockDashboards, nil
				},
			},
			"dashboard": &graphql.Field{
				Type:        dashboardType,
				Description: "Get a single dashboard by ID",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["id"].(string)
					if !ok {
						return nil, fmt.Errorf("invalid id argument")
					}
					log.Printf("Query: dashboard(id=%s)\n", id)
					dashboard := findDashboardByID(id)
					if dashboard == nil {
						log.Printf("Dashboard not found: %s\n", id)
					}
					return dashboard, nil
				},
			},
			"thenaRequests": &graphql.Field{
				Type:        graphql.NewList(thenaRequestType),
				Description: "Get list of all Thena webhook requests",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					thenaRequestsMutex.Lock()
					defer thenaRequestsMutex.Unlock()
					log.Printf("Query: thenaRequests - returning %d requests", len(thenaRequests))
					// Return in reverse order (newest first)
					result := make([]ThenaRequest, len(thenaRequests))
					for i, req := range thenaRequests {
						result[len(thenaRequests)-1-i] = req
					}
					return result, nil
				},
			},
		},
	})

	// Create schema
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: rootQuery,
	})
	if err != nil {
		log.Fatalf("Failed to create GraphQL schema: %v", err)
	}

	// Create GraphQL handler
	graphqlHandler := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	// Create CORS handler
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Setup routes
	mux := http.NewServeMux()

	// GraphQL endpoint
	mux.Handle("/graphql", corsHandler.Handler(graphqlHandler))

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Internal endpoint for thena-sync to forward events
	mux.HandleFunc("/internal/thena/events", thenaWebhookHandler)

	// Public REST API endpoint for frontend
	mux.HandleFunc("/api/requests", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		thenaRequestsMutex.Lock()
		defer thenaRequestsMutex.Unlock()

		w.Header().Set("Content-Type", "application/json")

		// Return in reverse order (newest first)
		result := make([]ThenaRequest, len(thenaRequests))
		for i, req := range thenaRequests {
			result[len(thenaRequests)-1-i] = req
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"count":   len(result),
			"requests": result,
		})
	})

	// Root endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"service": "Customer Success Backend",
			"version": "1.0.0",
			"graphql": "/graphql",
			"health":  "/health",
		})
	})

	// Create server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Backend server starting on port %s", port)
		log.Printf("📊 GraphQL endpoint: http://localhost:%s/graphql", port)
		log.Printf("💚 Health check: http://localhost:%s/health", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
