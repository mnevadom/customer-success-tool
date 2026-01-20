# Salesforce JWT Bearer Integration

This service implements a production-ready, server-to-server Salesforce integration using the **JWT Bearer OAuth flow**. No interactive login or callback is required.

## Table of Contents

- [Overview](#overview)
- [Authentication Flow](#authentication-flow)
- [Environment Variables](#environment-variables)
- [Setup Instructions](#setup-instructions)
  - [1. Generate RSA Key Pair (Development)](#1-generate-rsa-key-pair-development)
  - [2. Configure Salesforce Connected App](#2-configure-salesforce-connected-app)
  - [3. Set Environment Variables](#3-set-environment-variables)
- [Health Check](#health-check)
- [Troubleshooting](#troubleshooting)
- [Security Best Practices](#security-best-practices)
- [API Endpoints](#api-endpoints)

## Overview

This integration service:

- ✅ Authenticates using **JWT Bearer flow** (server-to-server, no user interaction)
- ✅ Automatically manages access tokens with expiry awareness
- ✅ Implements retry logic with exponential backoff for transient errors
- ✅ Supports both production and sandbox environments
- ✅ Provides optional caching for frequently requested resources
- ✅ Includes comprehensive error messages for troubleshooting
- ✅ Fully configurable via environment variables
- ✅ Never logs or commits secrets

## Authentication Flow

The JWT Bearer flow works as follows:

1. **Service creates a JWT** signed with the private key
2. **JWT contains claims**: Client ID, Username, Audience (login URL), Expiry
3. **Service exchanges JWT for access token** via Salesforce OAuth endpoint
4. **Access token is cached** and automatically refreshed before expiry
5. **API requests use the access token** in Authorization header

**No user interaction or callback is required** - the service authenticates autonomously.

## Environment Variables

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `SALESFORCE_CLIENT_ID` | Consumer Key from Connected App | `3MVG9riCAn8HH...` |
| `SALESFORCE_USERNAME` | Service user username (acts as this user) | `integration@company.com` |
| `SALESFORCE_AUD` | Environment identifier | `https://login.salesforce.com` or `https://test.salesforce.com` or `production` or `sandbox` |
| `SALESFORCE_PRIVATE_KEY` **OR**<br/>`SALESFORCE_PRIVATE_KEY_PATH` **OR**<br/>`SALESFORCE_PRIVATE_KEY_SECRET_NAME` | Private key (choose one method):<br/>• Direct PEM string<br/>• Path to PEM file<br/>• Secret manager identifier | See [Private Key Configuration](#private-key-configuration) |

### Optional Variables (with defaults)

| Variable | Description | Default |
|----------|-------------|---------|
| `SALESFORCE_API_VERSION` | Salesforce API version | `v58.0` |
| `SALESFORCE_HTTP_TIMEOUT_MS` | HTTP request timeout in milliseconds | `15000` (15 seconds) |
| `SALESFORCE_MAX_RETRIES` | Maximum retry attempts for failed requests | `3` |
| `SALESFORCE_BACKOFF_BASE_MS` | Base backoff delay in milliseconds | `200` |
| `SALESFORCE_ENABLE_CACHE` | Enable caching (`true` or `false`) | `false` |
| `SALESFORCE_CACHE_TTL_SECONDS` | Cache time-to-live in seconds | `300` (5 minutes) |

### Private Key Configuration

You **must** configure the private key using one of these methods:

#### Method 1: Environment Variable (Simple, for testing)

```bash
export SALESFORCE_PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA...
...
-----END RSA PRIVATE KEY-----"
```

#### Method 2: File Path (Recommended for containers)

```bash
export SALESFORCE_PRIVATE_KEY_PATH="/secrets/salesforce-key.pem"
```

Mount the key file as a secret in your container.

#### Method 3: Secret Manager (Not yet implemented)

```bash
export SALESFORCE_PRIVATE_KEY_SECRET_NAME="salesforce-jwt-key"
```

**Note**: Secret manager integration is a placeholder - use Method 1 or 2 for now.

## Setup Instructions

### 1. Generate RSA Key Pair (Development)

For development, use the provided script to generate a key pair:

```bash
cd salesforce-sync
./generate-keypair.sh
```

This will create:
- `private_key.pem` - **Keep this secret!** (added to .gitignore)
- `public_cert.pem` - Give this to IT to upload to Salesforce

**For production**: Your IT/Security team should generate and manage the private key using your organization's key management system.

### 2. Configure Salesforce Connected App

IT needs to create or update a Connected App in Salesforce:

#### Step-by-Step for IT:

1. **Navigate to Setup → App Manager → New Connected App**

2. **Basic Information**:
   - Connected App Name: `Customer Success Tool Integration`
   - API Name: `Customer_Success_Tool_Integration`
   - Contact Email: `your-email@company.com`

3. **API (Enable OAuth Settings)**:
   - ✅ Enable OAuth Settings
   - ✅ Enable for Device Flow (not required but harmless)
   - ✅ Use digital signatures
     - Upload the `public_cert.pem` file provided by the dev team

4. **Selected OAuth Scopes**:
   - Add these scopes:
     - `Access and manage your data (api)`
     - `Perform requests on your behalf at any time (refresh_token, offline_access)`
     - `Access your basic information (id, profile, email, address, phone)`

5. **Callback URL**:
   - Not required for JWT Bearer flow, but Salesforce requires one
   - Use: `https://login.salesforce.com/services/oauth2/success`

6. **Save** and wait 2-10 minutes for changes to propagate

7. **Click "Manage Consumer Details"** to get the **Consumer Key**
   - This is your `SALESFORCE_CLIENT_ID`

8. **Configure User Authorization**:

   **Option A: Pre-authorize all admin-approved users** (Recommended)
   - Edit the Connected App policies
   - Under "OAuth Policies", set:
     - Permitted Users: `Admin approved users are pre-authorized`
     - IP Relaxation: `Relax IP restrictions` (or configure as per your security policy)

   **Option B: Assign specific users**
   - Create a Permission Set or Profile
   - Assign the `SALESFORCE_USERNAME` user to this Permission Set
   - In Connected App → Manage Profiles/Permission Sets, add your Permission Set

9. **Verify the service user**:
   - Ensure the username in `SALESFORCE_USERNAME` exists in Salesforce
   - Verify the user has necessary permissions to access Account objects
   - The user should have an active license

### 3. Set Environment Variables

#### For Local Development:

Create a `.env` file (NOT committed to git):

```bash
# Required
SALESFORCE_CLIENT_ID=3MVG9riCAn8HHkYV8Vp.R7uGxqCej...
SALESFORCE_USERNAME=integration@company.com
SALESFORCE_AUD=https://test.salesforce.com
SALESFORCE_PRIVATE_KEY_PATH=./private_key.pem

# Optional
SALESFORCE_API_VERSION=v58.0
SALESFORCE_ENABLE_CACHE=true
SALESFORCE_CACHE_TTL_SECONDS=600
```

Then run:

```bash
export $(cat .env | xargs)
go run main.go
```

#### For Okteto/Kubernetes Deployment:

Create a Kubernetes secret for the private key:

```bash
kubectl create secret generic salesforce-jwt-key \
  --from-file=private_key.pem=./private_key.pem \
  -n your-namespace
```

Then configure your deployment to:

1. **Mount the secret as a file**:

```yaml
volumes:
  - name: salesforce-key
    secret:
      secretName: salesforce-jwt-key
volumeMounts:
  - name: salesforce-key
    mountPath: /secrets
    readOnly: true
```

2. **Set environment variables**:

```yaml
env:
  - name: SALESFORCE_CLIENT_ID
    value: "3MVG9riCAn8HHkYV8Vp.R7uGxqCej..."
  - name: SALESFORCE_USERNAME
    value: "integration@company.com"
  - name: SALESFORCE_AUD
    value: "https://login.salesforce.com"
  - name: SALESFORCE_PRIVATE_KEY_PATH
    value: "/secrets/private_key.pem"
  - name: SALESFORCE_API_VERSION
    value: "v58.0"
```

## Health Check

The service provides a health check endpoint to verify the integration:

### Internal Health Check (on startup)

The service automatically runs a health check on startup. If it fails, you'll see detailed error messages in the logs.

### HTTP Health Endpoint

```bash
curl http://localhost:8080/health
```

**Success Response** (HTTP 200):

```json
{
  "status": "healthy",
  "instance": "https://your-instance.salesforce.com",
  "api_version": "v58.0"
}
```

**Failure Response** (HTTP 503):

```json
{
  "status": "error",
  "error": "Authentication failed (HTTP 400): invalid_grant\n\nTroubleshooting:\n• INVALID_GRANT typically means:\n  1. The service user is not authorized for this Connected App\n     → Ask IT to pre-approve the user or enable 'Admin approved users are pre-authorized'\n  ..."
}
```

## Troubleshooting

### Common Errors and Solutions

#### Error: `SALESFORCE_CLIENT_ID is required`

**Cause**: Environment variable not set.

**Solution**: Set all required environment variables (see [Environment Variables](#environment-variables)).

---

#### Error: `no private key source configured`

**Cause**: None of the private key environment variables are set.

**Solution**: Set one of:
- `SALESFORCE_PRIVATE_KEY`
- `SALESFORCE_PRIVATE_KEY_PATH`
- `SALESFORCE_PRIVATE_KEY_SECRET_NAME`

---

#### Error: `Authentication failed (HTTP 400): invalid_grant`

**Cause**: This is the most common error. Possible reasons:

1. **Service user not authorized**
   - Solution: IT needs to pre-authorize users in the Connected App OR assign the user to a Permission Set/Profile

2. **Public certificate doesn't match private key**
   - Solution: Verify IT uploaded the correct `public_cert.pem` that matches your `private_key.pem`
   - Regenerate both keys if needed and re-upload the public cert

3. **Username is incorrect**
   - Solution: Verify `SALESFORCE_USERNAME` exactly matches a user in Salesforce (case-sensitive)

4. **Wrong environment**
   - Solution: If using sandbox, ensure `SALESFORCE_AUD=https://test.salesforce.com`
   - If using production, ensure `SALESFORCE_AUD=https://login.salesforce.com`

---

#### Error: `Authentication failed (HTTP 400): invalid_client_id`

**Cause**: The `SALESFORCE_CLIENT_ID` (Consumer Key) is incorrect.

**Solution**:
- Go to Salesforce → Setup → App Manager → Your Connected App → View
- Click "Manage Consumer Details"
- Copy the Consumer Key and update `SALESFORCE_CLIENT_ID`

---

#### Error: `User hasn't approved this application`

**Cause**: The Connected App requires user approval, but JWT Bearer flow is server-to-server.

**Solution**: In the Connected App:
- Edit → OAuth Policies
- Set "Permitted Users" to `Admin approved users are pre-authorized`
- OR assign the user to a Permission Set/Profile that has access to this Connected App

---

#### Error: `failed to parse private key - ensure it's in PKCS#1 or PKCS#8 PEM format`

**Cause**: The private key file is not in the correct format.

**Solution**: Ensure your private key is in PEM format and starts with:
- `-----BEGIN RSA PRIVATE KEY-----` (PKCS#1), or
- `-----BEGIN PRIVATE KEY-----` (PKCS#8)

If you have a different format, convert it:

```bash
# Convert PKCS#8 to PKCS#1
openssl rsa -in private_key.pem -out private_key_pkcs1.pem

# Convert PEM to PKCS#1 (if you have a certificate)
openssl x509 -in cert.pem -inform PEM -out cert.der -outform DER
```

---

#### Error: `query failed with status 401: Unauthorized`

**Cause**: Access token expired or revoked.

**Solution**: The service automatically refreshes tokens, but this can happen if:
- The Connected App was modified or deleted
- The service user's password was changed or user was deactivated
- IP restrictions are blocking the request

Check the service user's status and Connected App configuration.

---

#### Error: `rate limited (HTTP 429)`

**Cause**: Salesforce API rate limit exceeded.

**Solution**: The service automatically retries with exponential backoff. To reduce rate limiting:
- Enable caching: `SALESFORCE_ENABLE_CACHE=true`
- Increase cache TTL: `SALESFORCE_CACHE_TTL_SECONDS=600`
- Reduce query frequency
- Contact Salesforce to increase API limits

## Security Best Practices

### 🔒 Secrets Management

1. **Never commit secrets to git**
   - `private_key.pem` is in `.gitignore`
   - Never commit `.env` files

2. **Use Kubernetes secrets for production**
   - Mount private keys as files from secrets
   - Use secret management tools (HashiCorp Vault, AWS Secrets Manager, etc.)

3. **Rotate keys periodically**
   - Generate new key pairs every 90-180 days
   - Update Salesforce Connected App with new public cert
   - Deploy new private key to production

4. **Restrict access to private keys**
   - Only authorized personnel should have access
   - Use role-based access control (RBAC)

### 🛡️ Network Security

1. **Use IP restrictions** (if applicable)
   - Configure IP ranges in Salesforce Connected App
   - Ensure your deployment IP is whitelisted

2. **Use TLS/HTTPS** for all communications
   - The service always uses HTTPS for Salesforce API calls

### 📝 Logging

- The service **never logs**:
  - Private keys
  - Access tokens
  - Client secrets

- The service **does log**:
  - Masked Client ID (first 4 and last 4 characters)
  - Username
  - Token expiry times
  - API errors and retry attempts

## API Endpoints

### `GET /health`

Health check endpoint - verifies authentication and API access.

**Response**:
```json
{
  "status": "healthy",
  "instance": "https://your-instance.salesforce.com",
  "api_version": "v58.0"
}
```

---

### `GET /clients`

Retrieve all Salesforce accounts (up to 100).

**Response**:
```json
[
  {
    "id": "001...",
    "name": "Acme Corp",
    "status": "Active",
    "owner": "John Doe",
    "createdAt": "2023-01-15T10:30:00.000+0000",
    "lastActivity": "2024-01-10T14:20:00.000+0000",
    "tags": ["Enterprise", "High Priority"],
    "summary": "Large enterprise customer",
    "totalARR": "$250000",
    "nextRenewalDate": "2024-12-31",
    "daysUntilRenewal": 150,
    "numberOfUnits": 500,
    "currentAccountsCreated": 50,
    "currentMAU": 1200,
    "installType": "Cloud",
    "region": "North America",
    "saOwner": "Jane Smith"
  }
]
```

---

### `GET /clients/{accountId}`

Retrieve a specific Salesforce account by ID.

**Response**:
```json
{
  "id": "001...",
  "name": "Acme Corp",
  ...
}
```

## Development

### Running Locally

```bash
# Generate key pair (first time only)
./generate-keypair.sh

# Set environment variables
export SALESFORCE_CLIENT_ID=your_client_id
export SALESFORCE_USERNAME=your_username
export SALESFORCE_AUD=https://test.salesforce.com
export SALESFORCE_PRIVATE_KEY_PATH=./private_key.pem

# Run the service
go run main.go
```

### Testing

```bash
# Health check
curl http://localhost:8080/health

# Get clients
curl http://localhost:8080/clients

# Get specific client
curl http://localhost:8080/clients/001XXXXXXXXXXXXXXX
```

## Support

For issues:

1. Check the logs for detailed error messages
2. Run the health check: `curl http://localhost:8080/health`
3. Verify environment variables are set correctly
4. Review the [Troubleshooting](#troubleshooting) section
5. Contact IT to verify Connected App configuration

## References

- [Salesforce JWT Bearer Flow Documentation](https://help.salesforce.com/s/articleView?id=sf.remoteaccess_oauth_jwt_flow.htm)
- [Connected App Setup](https://help.salesforce.com/s/articleView?id=sf.connected_app_create.htm)
- [OAuth 2.0 JWT Bearer Token Flow](https://datatracker.ietf.org/doc/html/rfc7523)
