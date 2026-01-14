# Thena Integration - Setup Guide

## ✅ Service Successfully Deployed!

Your Thena integration service is now running at:
**https://thena-sync-agent-87kkzgw4nmq4.demo.okteto.dev**

## 🔑 Getting Your Thena API Key

1. Log in to **Thena** at https://app.thena.ai/
2. Go to **Dashboard** → **Organization Settings** → **Security and Access**
3. Click **Generate API Key**
4. Copy your API key securely

**Important Notes:**
- API keys are tied to individual users
- Rate limit: 60 requests per minute per user/org/IP
- Store your key securely - never commit it to version control

## 🚀 Option 1: Quick Setup (Environment Variable)

```bash
# Set your Thena API key
export THENA_API_KEY="your-thena-api-key-here"

# Redeploy the service
okteto deploy --wait
```

## 🔒 Option 2: Production Setup (Okteto Secrets)

```bash
# Create Okteto secret
okteto secret create THENA_API_KEY your-thena-api-key-here

# Then update the Helm chart to use the secret
# (Requires modifying thena-sync/chart/templates/deployment.yaml)
```

## 🧪 Testing the API

### 1. Health Check
```bash
curl https://thena-sync-agent-87kkzgw4nmq4.demo.okteto.dev/health | jq
```

**Expected (without API key):**
```json
{
  "status": "healthy",
  "service": "thena-sync",
  "time": "2026-01-14T12:00:00Z",
  "thenaConnected": false
}
```

**Expected (with API key):**
```json
{
  "status": "healthy",
  "service": "thena-sync",
  "time": "2026-01-14T12:00:00Z",
  "thenaConnected": true
}
```

### 2. Test Connection
```bash
curl https://thena-sync-agent-87kkzgw4nmq4.demo.okteto.dev/api/test-connection | jq
```

**Expected (with API key):**
```json
{
  "success": true,
  "message": "Successfully connected to Thena API",
  "baseURL": "https://bolt.thena.ai"
}
```

### 3. Get All Requests
```bash
curl https://thena-sync-agent-87kkzgw4nmq4.demo.okteto.dev/api/requests | jq
```

**Expected response:**
```json
{
  "success": true,
  "totalCount": 42,
  "page": 1,
  "totalPages": 1,
  "requests": [
    {
      "id": "req_abc123",
      "title": "Feature request from customer",
      "description": "Customer wants dark mode...",
      "status": "open",
      "customerName": "Acme Corp",
      "assignedTo": "Mario",
      "createdAt": "2026-01-14T10:00:00Z",
      "priority": "high"
    }
  ]
}
```

### 4. Get Cards Grouped by Status
```bash
curl https://thena-sync-agent-87kkzgw4nmq4.demo.okteto.dev/api/cards-by-status | jq
```

**Expected response:**
```json
{
  "success": true,
  "cards": {
    "open": [...],
    "in_progress": [...],
    "waiting": [...],
    "done": [...]
  }
}
```

### 5. Get Specific Request
```bash
curl https://thena-sync-agent-87kkzgw4nmq4.demo.okteto.dev/api/requests/req_abc123 | jq
```

## 📊 Available API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Service health check |
| `/api/test-connection` | GET | Test Thena API connectivity |
| `/api/requests` | GET | Get all requests from REQUESTS dashboard |
| `/api/requests/{id}` | GET | Get specific request by ID |
| `/api/cards-by-status` | GET | Get requests grouped by status column |

## 📋 Data Structure

Each request/card includes:

```typescript
{
  id: string           // Request ID
  title: string        // Request title
  description: string  // Request description
  status: string       // Status column (open, in_progress, etc.)
  customerName: string // Customer name
  assignedTo: string   // Assigned person name
  createdAt: string    // ISO 8601 timestamp
  priority: string     // Priority level
}
```

## 🔍 View Logs

```bash
# Watch service logs
kubectl logs -n ${OKTETO_NAMESPACE} -l app=thena-sync -f

# Check pod status
kubectl get pods -n ${OKTETO_NAMESPACE} | grep thena
```

## 🐛 Troubleshooting

### "Thena API key not configured"
Set the `THENA_API_KEY` environment variable and redeploy:
```bash
export THENA_API_KEY="your-key"
okteto deploy --wait
```

### "Failed to connect" / API errors
- Verify your API key is correct
- Check that you have the right permissions in Thena
- Ensure your API key hasn't been revoked

### Rate limit errors (429)
- You're exceeding 60 requests per minute
- Implement caching or reduce request frequency
- Wait a minute and try again

### No data returned
- Verify you have requests in your Thena REQUESTS dashboard
- Check the API response for error messages
- Review pod logs for detailed error information

## 📚 Resources

- **Thena API Documentation**: https://docs.thena.ai/api-reference/introduction
- **Get All Requests API**: https://help.thena.ai/reference/get-all-requests
- **Thena Help Center**: https://help.thena.ai/docs/getting-started

## 🔮 Next Steps

1. **Get your API key** from Thena dashboard
2. **Set the environment variable** and redeploy
3. **Test the connection** with `/api/test-connection`
4. **Fetch your requests** with `/api/requests`
5. **Integrate with your frontend** to display Thena data

---

**Service URL**: https://thena-sync-agent-87kkzgw4nmq4.demo.okteto.dev

Ready to connect to Thena! 🚀
