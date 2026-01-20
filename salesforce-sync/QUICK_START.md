# Salesforce JWT Integration - Quick Start

## ✅ Current Status

The Salesforce JWT Bearer integration is **fully deployed and configured** - it just needs credentials!

**Service Status**: 🔴 CrashLoopBackOff (expected - waiting for credentials)

**Error Message**:
```
❌ Configuration error: failed to load private key:
   failed to read private key file /secrets/private_key.pem:
   no such file or directory
```

This is **correct behavior** - the service won't start until you provide the private key via admin variables.

---

## 🚀 How to Enable the Integration (3 Steps)

### Step 1: Generate RSA Key Pair

Run the key generation script:

```bash
cd salesforce-sync
./generate-keypair.sh
```

This creates:
- `private_key.pem` - Keep secret!
- `public_cert.pem` - Give to IT for Salesforce

### Step 2: Configure Salesforce Connected App (IT Task)

IT needs to upload `public_cert.pem` to Salesforce:

1. Go to **Salesforce Setup → App Manager**
2. Edit your Connected App (or create one)
3. Under **Use digital signatures**, upload `public_cert.pem`
4. Set **Permitted Users** to `Admin approved users are pre-authorized`
5. **Save** and wait 2-10 minutes for propagation
6. Get the **Consumer Key** (this becomes `SALESFORCE_CLIENT_ID`)

### Step 3: Set Okteto Admin Variables

In Okteto UI → **Settings → Admin Variables**, set:

| Variable | Value | Where to Get It |
|----------|-------|-----------------|
| **SALESFORCE_PRIVATE_KEY** | Full content of `private_key.pem` | Copy entire file content including `-----BEGIN` and `-----END` lines |
| **SALESFORCE_CLIENT_ID** | Consumer Key | From Connected App → Manage Consumer Details |
| **SALESFORCE_USERNAME** | Service user email | The Salesforce user account the service will act as |
| **SALESFORCE_AUD** | `https://login.salesforce.com` or `https://test.salesforce.com` | Use `test.salesforce.com` for sandbox |
| **SALESFORCE_API_VERSION** | `v58.0` | ⚪ Optional (has default) |

**That's it!** On the next deployment:
1. The secret will be automatically created from `SALESFORCE_PRIVATE_KEY`
2. The pod will restart with the mounted private key
3. The service will authenticate and start successfully

---

## 🧪 Verify It's Working

After setting admin variables, redeploy:

```bash
okteto deploy --wait
```

### Check Deployment Logs

You should see:
```
✅ Secret salesforce-jwt-key created/updated
```

### Check Pod Status

```bash
kubectl get pods -n ${OKTETO_NAMESPACE} -l app=salesforce-sync
```

Should show `Running` (not CrashLoopBackOff).

### Check Service Logs

```bash
kubectl logs -n ${OKTETO_NAMESPACE} deployment/salesforce-sync --tail=50
```

Success looks like:
```
🚀 Starting Salesforce Integration Service...
✓ Configuration loaded: ClientID=3MVG****, Username=..., LoginURL=...
✓ Private key loaded successfully (PKCS#1 format)
🏥 Running health check...
📝 Creating JWT assertion...
🔑 Requesting access token from https://login.salesforce.com/services/oauth2/token
✅ Successfully obtained access token (expires at ...)
   Instance URL: https://your-instance.salesforce.com
   Testing API access with query: SELECT Id, Name FROM Account LIMIT 1
✅ Health check passed! Successfully queried Salesforce.
   Found X records
✅ Salesforce Integration Service started on port 9000
```

### Test Health Endpoint

```bash
kubectl run -n ${OKTETO_NAMESPACE} -it --rm debug \
  --image=curlimages/curl --restart=Never \
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

---

## 🆘 Troubleshooting

### Still CrashLoopBackOff after setting variables?

**Check deployment logs**:
```bash
okteto deploy --wait 2>&1 | grep -A5 "Create Salesforce JWT secret"
```

Should show:
```
✅ Secret salesforce-jwt-key created/updated
```

If it shows the warning instead, the `SALESFORCE_PRIVATE_KEY` admin variable is not set correctly.

### "invalid_grant" error in logs?

This means authentication failed. Common causes:

1. **Service user not authorized**
   - IT needs to enable "Admin approved users are pre-authorized" in Connected App

2. **Public cert doesn't match private key**
   - Verify IT uploaded the `public_cert.pem` that matches your `private_key.pem`
   - If you regenerated keys, give IT the new `public_cert.pem`

3. **Wrong username**
   - `SALESFORCE_USERNAME` must exactly match a Salesforce user (case-sensitive)

4. **Wrong environment**
   - Sandbox: `SALESFORCE_AUD=https://test.salesforce.com`
   - Production: `SALESFORCE_AUD=https://login.salesforce.com`

### "invalid_client_id" error?

The `SALESFORCE_CLIENT_ID` (Consumer Key) is wrong.
- Get it from Connected App → Manage Consumer Details

### Secret not mounting?

**Verify secret exists**:
```bash
kubectl get secret salesforce-jwt-key -n ${OKTETO_NAMESPACE}
```

If missing, the `SALESFORCE_PRIVATE_KEY` admin variable wasn't set when you deployed.

---

## 📖 Full Documentation

- **`README_SALESFORCE_JWT.md`** - Complete technical guide
- **`DEPLOYMENT_GUIDE.md`** - Detailed deployment instructions
- **`generate-keypair.sh`** - RSA key pair generation script

---

## 🔐 Security Reminder

- ✅ Private key stored in Okteto admin variables (encrypted)
- ✅ Automatically mounted as Kubernetes secret (not in git)
- ✅ Service never logs private keys or access tokens
- ✅ `private_key.pem` is in `.gitignore` (can't be committed)

**Never commit or share the private key!**

---

## 📝 Admin Variables Summary

Copy this checklist when configuring admin variables:

```
☐ SALESFORCE_PRIVATE_KEY (entire private_key.pem content)
☐ SALESFORCE_CLIENT_ID (Consumer Key from Salesforce)
☐ SALESFORCE_USERNAME (service user email)
☐ SALESFORCE_AUD (login.salesforce.com or test.salesforce.com)
☐ SALESFORCE_API_VERSION (optional, defaults to v58.0)
```

Once all are set, run `okteto deploy --wait` and the service will start! 🚀
