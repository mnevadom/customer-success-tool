# Salesforce Integration Service - Deployment Guide

## Overview

The Salesforce integration service is now deployed and running at:
**https://salesforce-sync-agent-87kkzgw4nmq4.demo.okteto.dev**

## Current Status

✅ Service is deployed and healthy
✅ Running in **mock mode** (no Salesforce credentials configured)
✅ All endpoints are accessible and responding correctly

## Available Endpoints

### Health Check
```bash
GET /health
```
Returns service health status and Salesforce connection state.

**Example:**
```bash
curl https://salesforce-sync-agent-87kkzgw4nmq4.demo.okteto.dev/health
```

### Test Salesforce Connection
```bash
GET /api/test-connection
```
Tests connectivity to Salesforce API.

**Example:**
```bash
curl https://salesforce-sync-agent-87kkzgw4nmq4.demo.okteto.dev/api/test-connection
```

### Get All Clients
```bash
GET /api/clients
```
Fetches all Salesforce Accounts mapped to Client format.

**Example:**
```bash
curl https://salesforce-sync-agent-87kkzgw4nmq4.demo.okteto.dev/api/clients
```

### Get Single Client
```bash
GET /api/clients/:id
```
Fetches a specific Salesforce Account by ID.

**Example:**
```bash
curl https://salesforce-sync-agent-87kkzgw4nmq4.demo.okteto.dev/api/clients/0012x000001ABC123
```

### Refresh Connection
```bash
POST /api/sync
```
Forces a reconnection to Salesforce API (useful when credentials change).

## Configuring Salesforce Credentials

To connect to your actual Salesforce instance, you need to set the following environment variables:

### Required Credentials:
- `SF_USERNAME` - Your Salesforce username (email)
- `SF_PASSWORD` - Your Salesforce password
- `SF_SECURITY_TOKEN` - Your Salesforce security token

### Optional:
- `SF_LOGIN_URL` - Salesforce login URL (defaults to https://login.salesforce.com)
  - Use https://test.salesforce.com for sandbox instances

### How to Set Credentials in Okteto:

You can set these as Okteto Secrets or environment variables:

**Option 1: Using Okteto Secrets (Recommended)**
```bash
okteto context use demo.okteto.dev
okteto secret create SF_USERNAME your-email@company.com
okteto secret create SF_PASSWORD your-salesforce-password
okteto secret create SF_SECURITY_TOKEN your-security-token
```

Then update the Helm deployment to use secrets instead of environment variables.

**Option 2: Set as Environment Variables**

Update the `okteto.yml` deploy command to include actual values:
```yaml
--set env[1].value="your-email@company.com" \
--set env[2].value="your-salesforce-password" \
--set env[3].value="your-security-token" \
```

**Important:** Never commit actual credentials to version control!

### Getting Your Salesforce Security Token

1. Log in to Salesforce
2. Go to **Setup** → **Personal Setup** → **My Personal Information** → **Reset My Security Token**
3. Click "Reset Security Token"
4. Check your email for the new security token

## Salesforce Field Mapping

The service automatically maps Salesforce Account fields to your Client data model:

| Salesforce Field | Client Field | Notes |
|-----------------|--------------|-------|
| Id | id | Account ID |
| Name | name | Account name |
| Customer_Status__c | status | "Active" or "At risk" |
| Annual_Revenue__c | totalARR | Formatted as currency |
| Renewal_Date__c | nextRenewalDate | ISO date string |
| Days_Until_Renewal__c | daysUntilRenewal | Number of days |
| Number_of_Units__c | numberOfUnits | Integer |
| Accounts_Created__c | currentAccountsCreated | Integer |
| Monthly_Active_Users__c | currentMAU | Integer |
| Install_Type__c | installType | e.g., "Self-Hosted" |
| Region__c | region | e.g., "Europe", "US" |
| SA_Owner__c | saOwner | Solutions Architect owner |
| Owner.Name | owner | Account owner name |
| Tags__c | tags | Comma-separated tags |
| Description | summary | Account description |
| CreatedDate | createdAt | ISO date string |
| LastActivityDate | lastActivity | ISO date string |

## Testing the Service

### 1. Health Check
```bash
curl -s https://salesforce-sync-agent-87kkzgw4nmq4.demo.okteto.dev/health | jq
```

Expected output (without credentials):
```json
{
  "status": "healthy",
  "service": "salesforce-sync",
  "time": "2026-01-14T11:42:07.948Z",
  "salesforceConnected": false
}
```

### 2. Test Connection
```bash
curl -s https://salesforce-sync-agent-87kkzgw4nmq4.demo.okteto.dev/api/test-connection | jq
```

Expected output (without credentials):
```json
{
  "success": false,
  "message": "Salesforce credentials not configured",
  "mode": "mock"
}
```

### 3. Fetch Clients
```bash
curl -s https://salesforce-sync-agent-87kkzgw4nmq4.demo.okteto.dev/api/clients | jq
```

## Next Steps

1. **Configure Salesforce Credentials**: Add your actual Salesforce credentials to connect to the live API
2. **Test Connection**: Verify the service can connect to Salesforce
3. **Fetch Real Data**: Use the `/api/clients` endpoint to retrieve actual Salesforce Accounts
4. **Integration**: Once verified, integrate the service with your backend (per user request: "we will do it once we check the API works fine")

## Future Enhancements

- **Streaming API / CometD**: Implement push notifications for real-time updates when Salesforce data changes
- **Caching**: Add Redis or in-memory caching to reduce API calls
- **Webhook Support**: Add webhooks to notify your application of changes
- **Field Customization**: Make field mapping configurable per organization

## Troubleshooting

### Service Won't Start
Check pod logs:
```bash
kubectl logs -n ${OKTETO_NAMESPACE} -l app=salesforce-sync
```

### Connection Fails
1. Verify credentials are correct
2. Check security token is current (they expire when password changes)
3. Verify IP restrictions in Salesforce (Okteto IPs need to be whitelisted)
4. For sandbox, ensure SF_LOGIN_URL is set to https://test.salesforce.com

### No Data Returned
1. Verify your Salesforce Accounts have the custom fields configured
2. Check the SOQL query in `server.js` matches your Salesforce schema
3. Review pod logs for API errors

## Architecture

```
Frontend (React) → Backend (Go/GraphQL) → [Future: Real-time client data]
                                              ↑
                                              |
                                        Salesforce-Sync (Node.js)
                                              ↓
                                        Salesforce API (REST)
                                              ↓
                                        [Future: Streaming API]
```

The service is designed to be standalone and testable independently before backend integration.
