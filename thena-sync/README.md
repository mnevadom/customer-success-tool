# Thena API Integration Service

Go-based REST API service for integrating with [Thena.ai](https://thena.ai) to fetch customer support requests and dashboard data.

## Features

- ✅ Fetch all requests from Thena REQUESTS dashboard
- ✅ Get request details by ID
- ✅ Group cards by status column
- ✅ RESTful API endpoints
- ✅ CORS enabled for frontend integration
- ✅ Health check endpoint
- ✅ Connection testing

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `THENA_API_KEY` | Your Thena API key | - | Yes |
| `THENA_BASE_URL` | Thena API base URL | `https://bolt.thena.ai` | No |
| `PORT` | Service port | `9100` | No |

## Getting Your Thena API Key

1. Log in to Thena at https://app.thena.ai/
2. Go to **Dashboard** → **Organization Settings** → **Security and Access**
3. Click **Generate API Key**
4. Copy and securely store your API key

**Note**: API keys are tied to individual users and have a rate limit of 60 requests per minute.

## API Endpoints

### Health Check
```bash
GET /health
```
Returns service health status and Thena connection state.

**Response:**
```json
{
  "status": "healthy",
  "service": "thena-sync",
  "time": "2026-01-14T12:00:00Z",
  "thenaConnected": true
}
```

### Test Connection
```bash
GET /api/test-connection
```
Tests connectivity to Thena API.

**Response:**
```json
{
  "success": true,
  "message": "Successfully connected to Thena API",
  "baseURL": "https://bolt.thena.ai"
}
```

### Get All Requests
```bash
GET /api/requests
```
Fetches all requests from Thena REQUESTS dashboard.

**Response:**
```json
{
  "success": true,
  "totalCount": 42,
  "page": 1,
  "totalPages": 1,
  "requests": [
    {
      "id": "req_123",
      "title": "Feature request: Dark mode",
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

### Get Request by ID
```bash
GET /api/requests/{id}
```
Fetches a specific request by ID.

**Response:**
```json
{
  "success": true,
  "request": {
    "id": "req_123",
    "title": "Feature request: Dark mode",
    "status": "open",
    ...
  }
}
```

### Get Cards Grouped by Status
```bash
GET /api/cards-by-status
```
Fetches all requests grouped by their status column.

**Response:**
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

## Data Mapping

The service maps Thena Request fields to a Card format:

| Thena Field | Card Field | Description |
|-------------|------------|-------------|
| `id` | `id` | Request ID |
| `title` | `title` | Request title |
| `description` | `description` | Request description |
| `status` | `status` | Status column value |
| `customer.name` | `customerName` | Customer name |
| `assignee.name` | `assignedTo` | Assigned person |
| `created_at` | `createdAt` | Creation timestamp (ISO 8601) |
| `priority` | `priority` | Request priority |

## Local Development

```bash
# Set environment variables
export THENA_API_KEY="your-api-key"
export PORT=9100

# Run the service
go run main.go
```

## Docker

```bash
# Build
docker build -t thena-sync .

# Run
docker run -p 9100:9100 \
  -e THENA_API_KEY="your-api-key" \
  thena-sync
```

## Okteto Deployment

The service is configured to deploy automatically with the rest of the application stack.

Set your API key:
```bash
export THENA_API_KEY="your-api-key"
okteto deploy --wait
```

## Rate Limits

Thena API has the following rate limits:
- **60 requests per minute** per user, organization, and IP address

The service automatically handles rate limit errors and returns appropriate error messages.

## Testing

```bash
# Test connection
curl http://localhost:9100/api/test-connection

# Get all requests
curl http://localhost:9100/api/requests

# Get requests by status
curl http://localhost:9100/api/cards-by-status

# Get specific request
curl http://localhost:9100/api/requests/req_123
```

## Security Notes

- Never commit your API key to version control
- API keys are sensitive - store them securely
- Use Okteto Secrets for production deployments
- Keys are tied to individual users with specific permissions

## Troubleshooting

### "Thena API key not configured"
Set the `THENA_API_KEY` environment variable with your API key from Thena.

### "Failed to connect" / Authentication errors
- Verify your API key is correct
- Check that your key hasn't been revoked
- Ensure your user has proper permissions

### Rate limit errors
- You're making more than 60 requests per minute
- Implement caching or reduce request frequency
- Consider batching operations

## References

- [Thena API Documentation](https://docs.thena.ai/api-reference/introduction)
- [Get All Requests API](https://help.thena.ai/reference/get-all-requests)
- [Thena Help Center](https://help.thena.ai/docs/getting-started)
