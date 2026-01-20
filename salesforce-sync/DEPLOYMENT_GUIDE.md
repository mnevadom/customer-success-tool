# Salesforce Integration Deployment Guide

## Current Status

✅ **Code Deployed**: The new JWT Bearer integration is deployed and ready
⚠️ **Service Status**: CrashLoopBackOff (expected - waiting for credentials)

The service is **designed to fail fast** until proper credentials are configured. This is a security feature.

## What IT Needs to Provide

### 1. Generate Key Pair

IT should generate an RSA key pair using their secure key management system.

For testing/dev purposes only, you can use:

```bash
cd salesforce-sync
./generate-keypair.sh
```

This creates:
- `private_key.pem` - Keep secret! (for deployment)
- `public_cert.pem` - Upload to Salesforce Connected App

### 2. Configure Salesforce Connected App

IT needs to:

1. **Create/Update Connected App** in Salesforce Setup → App Manager
2. **Enable OAuth Settings**
3. **Upload Public Certificate** (`public_cert.pem`) under "Use digital signatures"
4. **Add OAuth Scopes**:
   - Access and manage your data (api)
   - Perform requests on your behalf at any time (refresh_token, offline_access)
5. **Configure User Authorization**:
   - Set "Permitted Users" to `Admin approved users are pre-authorized`
   - OR assign specific users via Permission Set
6. **Get Consumer Key** - This becomes `SALESFORCE_CLIENT_ID`

**Full instructions**: See `README_SALESFORCE_JWT.md` → "Configure Salesforce Connected App"

### 3. Deploy Credentials to Okteto

#### Method A: Using Kubernetes Secrets (Recommended)

```bash
# 1. Create secret for private key
kubectl create secret generic salesforce-jwt-key \
  --from-file=private_key.pem=./private_key.pem \
  -n agent-87kkzgw4nmq4

# 2. Update Helm deployment to use secret
# Edit okteto.yml or use Helm values to set:
#   - SALESFORCE_CLIENT_ID=<Consumer Key from Connected App>
#   - SALESFORCE_USERNAME=<service user email>
#   - SALESFORCE_AUD=https://login.salesforce.com  (or https://test.salesforce.com for sandbox)
#   - SALESFORCE_PRIVATE_KEY_PATH=/secrets/private_key.pem
#
# And mount the secret as a volume at /secrets
```

#### Method B: Using Environment Variable (Less secure, for testing only)

```bash
# Set environment variables directly (not recommended for production):
# - SALESFORCE_CLIENT_ID
# - SALESFORCE_USERNAME
# - SALESFORCE_AUD
# - SALESFORCE_PRIVATE_KEY (entire PEM content)
```

### 4. Update okteto.yml

Current salesforce-sync deployment in `okteto.yml` uses old OAuth variables. Update to:

```yaml
- name: Deploy salesforce-sync with Helm
  command: |
    helm upgrade --install salesforce-sync ./salesforce-sync/chart \
      --set image.repository=${OKTETO_BUILD_SALESFORCE_SYNC_IMAGE%:*} \
      --set image.tag=${OKTETO_BUILD_SALESFORCE_SYNC_IMAGE##*:} \
      --set env[0].name="PORT" \
      --set env[0].value="9000" \
      --set env[1].name="SALESFORCE_CLIENT_ID" \
      --set env[1].value="${SALESFORCE_CLIENT_ID:-}" \
      --set env[2].name="SALESFORCE_USERNAME" \
      --set env[2].value="${SALESFORCE_USERNAME:-}" \
      --set env[3].name="SALESFORCE_AUD" \
      --set env[3].value="${SALESFORCE_AUD:-https://login.salesforce.com}" \
      --set env[4].name="SALESFORCE_PRIVATE_KEY_PATH" \
      --set env[4].value="/secrets/private_key.pem" \
      --set env[5].name="SALESFORCE_API_VERSION" \
      --set env[5].value="${SALESFORCE_API_VERSION:-v58.0}" \
      --set volumes[0].name="salesforce-key" \
      --set volumes[0].secret.secretName="salesforce-jwt-key" \
      --set volumeMounts[0].name="salesforce-key" \
      --set volumeMounts[0].mountPath="/secrets" \
      --set volumeMounts[0].readOnly="true" \
      --wait
```

Or update the Helm chart directly in `salesforce-sync/chart/values.yaml` and `templates/deployment.yaml`.

## Testing the Integration

Once credentials are configured:

### 1. Check Pod Status

```bash
kubectl get pods -n agent-87kkzgw4nmq4 -l app=salesforce-sync
```

Should show `Running` instead of `CrashLoopBackOff`.

### 2. Check Logs

```bash
kubectl logs -n agent-87kkzgw4nmq4 deployment/salesforce-sync --tail=50
```

You should see:
```
✓ Configuration loaded: ClientID=3MVG****, Username=..., LoginURL=...
✓ Private key loaded successfully (PKCS#1 format)
🏥 Running health check...
📝 Creating JWT assertion...
🔑 Requesting access token from https://login.salesforce.com/services/oauth2/token
✅ Successfully obtained access token (expires at ...)
   Testing API access with query: SELECT Id, Name FROM Account LIMIT 1
✅ Health check passed! Successfully queried Salesforce.
✅ Salesforce Integration Service started on port 9000
```

### 3. Test Health Endpoint

```bash
# From inside cluster
kubectl run -n agent-87kkzgw4nmq4 -it --rm debug --image=curlimages/curl --restart=Never \
  -- curl http://salesforce-sync:9000/health
```

Expected response:
```json
{
  "status": "healthy",
  "instance": "https://your-instance.salesforce.com",
  "api_version": "v58.0"
}
```

### 4. Test API Endpoint

```bash
# Get all accounts
kubectl run -n agent-87kkzgw4nmq4 -it --rm debug --image=curlimages/curl --restart=Never \
  -- curl http://salesforce-sync:9000/clients
```

## Troubleshooting

### Service still in CrashLoopBackOff after configuring credentials

**Check logs**:
```bash
kubectl logs -n agent-87kkzgw4nmq4 deployment/salesforce-sync --tail=100
```

Common errors and solutions are documented in `README_SALESFORCE_JWT.md` → "Troubleshooting" section.

### "invalid_grant" error

This usually means:
1. Service user not authorized in Connected App → Enable "Admin approved users are pre-authorized"
2. Public certificate doesn't match private key → Re-upload correct certificate
3. Wrong username → Verify exact username (case-sensitive)

### "invalid_client_id" error

Consumer Key is wrong → Get it from Connected App → Manage Consumer Details

## Environment Variables Reference

| Required | Variable | Example | Where to Get It |
|----------|----------|---------|-----------------|
| ✅ | `SALESFORCE_CLIENT_ID` | `3MVG9riCAn8...` | Salesforce Connected App → Consumer Key |
| ✅ | `SALESFORCE_USERNAME` | `integration@company.com` | Salesforce user that service will act as |
| ✅ | `SALESFORCE_AUD` | `https://login.salesforce.com` | `https://login.salesforce.com` (prod) or `https://test.salesforce.com` (sandbox) |
| ✅ | `SALESFORCE_PRIVATE_KEY_PATH` | `/secrets/private_key.pem` | Path to mounted private key file |
| ⚪ | `SALESFORCE_API_VERSION` | `v58.0` | Optional (default: v58.0) |
| ⚪ | `SALESFORCE_ENABLE_CACHE` | `true` | Optional (default: false) |
| ⚪ | `SALESFORCE_CACHE_TTL_SECONDS` | `600` | Optional (default: 300) |

## Security Checklist

- [ ] Private key stored in Kubernetes secret (not in git)
- [ ] Private key never appears in logs
- [ ] Public certificate uploaded to Salesforce
- [ ] Connected App configured with correct scopes
- [ ] Service user has necessary Salesforce permissions
- [ ] IP restrictions configured (if applicable)
- [ ] Environment is correct (production vs sandbox)

## Next Steps After Deployment

Once the service is running successfully:

1. **Verify Integration**: Check that `/clients` endpoint returns Salesforce accounts
2. **Monitor Performance**: Watch logs for retry patterns, rate limits
3. **Enable Caching** (optional): Set `SALESFORCE_ENABLE_CACHE=true` to reduce API calls
4. **Set Up Alerts**: Monitor for authentication failures, rate limiting
5. **Plan Key Rotation**: Schedule periodic key rotation (every 90-180 days)

## Support

For detailed documentation, see:
- `README_SALESFORCE_JWT.md` - Complete setup guide
- `generate-keypair.sh` - Key generation script
- Salesforce JWT Bearer Flow docs: https://help.salesforce.com/s/articleView?id=sf.remoteaccess_oauth_jwt_flow.htm

For issues, check the pod logs - they include detailed troubleshooting steps!
