#!/bin/bash

# Thena API Testing Script
# Usage: ./test-thena.sh [your-api-key]

THENA_API_KEY="$1"
THENA_ENDPOINT="https://thena-sync-agent-87kkzgw4nmq4.demo.okteto.dev"

if [ -z "$THENA_API_KEY" ]; then
    echo "❌ Error: Please provide your Thena API key"
    echo "Usage: ./test-thena.sh YOUR_THENA_API_KEY"
    echo ""
    echo "Get your API key from:"
    echo "https://app.thena.ai/ → Settings → Security and Access → Generate API Key"
    exit 1
fi

echo "🔑 Testing Thena API with provided key..."
echo ""

# Deploy with the API key
echo "📦 Deploying with Thena API key..."
export THENA_API_KEY="$THENA_API_KEY"
okteto deploy --wait

echo ""
echo "✅ Deployment complete!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🧪 Running API Tests"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Test 1: Health Check
echo "1️⃣  Health Check"
echo "────────────────────────────────────────────────────────"
curl -s "$THENA_ENDPOINT/health" | jq
echo ""

# Test 2: Connection Test
echo "2️⃣  Testing Thena Connection"
echo "────────────────────────────────────────────────────────"
curl -s "$THENA_ENDPOINT/api/test-connection" | jq
echo ""

# Test 3: Get All Requests
echo "3️⃣  Fetching All Requests"
echo "────────────────────────────────────────────────────────"
REQUESTS=$(curl -s "$THENA_ENDPOINT/api/requests")
echo "$REQUESTS" | jq

# Save first request ID for next test
FIRST_ID=$(echo "$REQUESTS" | jq -r '.requests[0].id // empty')
echo ""

# Test 4: Get Cards by Status
echo "4️⃣  Fetching Cards Grouped by Status"
echo "────────────────────────────────────────────────────────"
curl -s "$THENA_ENDPOINT/api/cards-by-status" | jq
echo ""

# Test 5: Get specific request (if we have an ID)
if [ -n "$FIRST_ID" ]; then
    echo "5️⃣  Fetching Specific Request (ID: $FIRST_ID)"
    echo "────────────────────────────────────────────────────────"
    curl -s "$THENA_ENDPOINT/api/requests/$FIRST_ID" | jq
    echo ""
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ All tests complete!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📝 Summary:"
curl -s "$THENA_ENDPOINT/api/requests" | jq -r '"Total Requests: \(.totalCount // 0)"'
echo ""
echo "🔗 Thena Sync Endpoint: $THENA_ENDPOINT"
echo "📖 API Documentation: thena-sync/README.md"
echo ""
