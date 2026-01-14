# Salesforce Sync Service

Real-time Salesforce integration service for the Customer Success application.

## Features

- **REST API Integration**: Fetches Account data from Salesforce in real-time
- **Automatic Mapping**: Maps Salesforce Account fields to Customer Success Client format
- **Connection Pooling**: Maintains persistent Salesforce connection
- **Health Checks**: Built-in health and connection testing endpoints
- **Future-Ready**: Structured to add Streaming API/CometD for push notifications

## API Endpoints

### Health & Testing
- `GET /health` - Service health check
- `GET /` - Service information and available endpoints
- `GET /api/test-connection` - Test Salesforce connectivity

### Data Endpoints
- `GET /api/clients` - Fetch all clients from Salesforce Accounts
- `GET /api/clients/:id` - Fetch specific client by Salesforce Account ID
- `POST /api/sync` - Force refresh Salesforce connection

### Future (Placeholder)
- `POST /api/streaming/subscribe` - Subscribe to Salesforce push events (coming soon)

## Salesforce Field Mapping

The service maps Salesforce Account fields to our Client format:

| Salesforce Field | Client Field | Notes |
|------------------|--------------|-------|
| `Id` | `id` | Salesforce Account ID |
| `Name` | `name` | Account name |
| `Customer_Status__c` | `status` | Custom field (Active/At risk) |
| `Owner.Name` | `owner` | Account owner name |
| `CreatedDate` | `createdAt` | Creation timestamp |
| `LastActivityDate` | `lastActivity` | Last activity timestamp |
| `Tags__c` | `tags` | Semicolon-separated tags |
| `Description` | `summary` | Account description |
| `Annual_Revenue__c` | `totalARR` | Annual recurring revenue |
| `Renewal_Date__c` | `nextRenewalDate` | Contract renewal date |
| `Days_Until_Renewal__c` | `daysUntilRenewal` | Days until renewal |
| `Number_of_Units__c` | `numberOfUnits` | License units |
| `Accounts_Created__c` | `currentAccountsCreated` | Accounts created |
| `Monthly_Active_Users__c` | `currentMAU` | Monthly active users |
| `Install_Type__c` | `installType` | Installation type |
| `Region__c` | `region` | Geographic region |
| `SA_Owner__c` | `saOwner` | Solutions Architect owner |

**Note**: Custom fields (ending in `__c`) need to be created in your Salesforce org.

## Environment Variables

### Required (for Salesforce connection)
```bash
SF_USERNAME=your-salesforce-username@example.com
SF_PASSWORD=your-salesforce-password
SF_SECURITY_TOKEN=your-security-token
```

### Optional
```bash
SF_LOGIN_URL=https://login.salesforce.com  # Use https://test.salesforce.com for sandbox
PORT=9000
NODE_ENV=production
```

## Getting Salesforce Credentials

### 1. Username & Password
- Your Salesforce login credentials

### 2. Security Token
- In Salesforce, go to: **Settings** → **Reset My Security Token**
- Token will be emailed to you
- Append the token to your password when authenticating via API

### 3. Connected App (Optional, for OAuth)
For production, consider using OAuth instead of username/password:
- Setup → App Manager → New Connected App
- Enable OAuth Settings
- Use OAuth 2.0 flow for better security

## Local Development

```bash
# Install dependencies
npm install

# Copy environment template
cp .env.example .env

# Edit .env with your Salesforce credentials
nano .env

# Start development server
npm run dev
```

## Testing

```bash
# Test health
curl http://localhost:9000/health

# Test Salesforce connection
curl http://localhost:9000/api/test-connection

# Fetch clients
curl http://localhost:9000/api/clients

# Fetch specific client
curl http://localhost:9000/api/clients/SALESFORCE_ACCOUNT_ID
```

## Deployment with Okteto

The service is deployed via Helm chart and configured in `okteto.yml`:

```bash
okteto deploy --wait
```

Environment variables are set via Helm:
```yaml
--set env[0].name="SF_USERNAME" \
--set env[0].value="${SF_USERNAME}"
```

## Future: Streaming API / Push Notifications

The service is structured to add Salesforce Streaming API support:

### Options:
1. **Platform Events**: Custom events published by Salesforce
2. **Change Data Capture (CDC)**: Real-time notifications on data changes
3. **PushTopic**: Subscribe to SOQL query results changes

### Implementation (Coming Soon):
```javascript
// Subscribe to Account changes
const channel = '/data/AccountChangeEvent';
conn.streaming.topic(channel).subscribe((message) => {
  console.log('Account changed:', message);
  // Push to frontend via WebSocket
});
```

## Architecture

```
┌─────────────┐         ┌──────────────────┐         ┌────────────┐
│   Frontend  │────────▶│ Salesforce Sync  │────────▶│ Salesforce │
│             │         │    Service       │         │     API    │
│             │◀────────│  (REST + Soon    │◀────────│            │
│             │         │   Streaming)     │         │            │
└─────────────┘         └──────────────────┘         └────────────┘
                               │
                               │ (Future)
                               ▼
                        ┌──────────────┐
                        │  WebSocket   │
                        │  to Frontend │
                        └──────────────┘
```

## Dependencies

- **express**: Web server framework
- **jsforce**: Salesforce API client (supports REST, Bulk, Streaming)
- **cors**: CORS middleware
- **dotenv**: Environment variable management

## Error Handling

- **503**: Salesforce not connected (credentials missing)
- **404**: Client/Account not found
- **500**: Salesforce API error or internal error

Service runs in **mock mode** if credentials are not configured.

## Security Notes

⚠️ **Important**:
- Never commit credentials to Git
- Use Okteto Secrets for production deployments
- Consider OAuth 2.0 for production (more secure than username/password)
- Rotate security tokens regularly
- Use sandbox (`test.salesforce.com`) for testing

## Support

For JSforce documentation: https://jsforce.github.io/
For Salesforce API: https://developer.salesforce.com/docs/apis
