#!/bin/bash

#
# Salesforce JWT Key Pair Generator
#
# This script generates an RSA key pair for Salesforce JWT Bearer authentication.
#
# ⚠️  FOR DEVELOPMENT USE ONLY ⚠️
#
# For production, your IT/Security team should generate and manage keys
# using your organization's key management system.
#
# Usage:
#   ./generate-keypair.sh
#
# Output:
#   - private_key.pem     (Keep secret! Not committed to git)
#   - public_cert.pem     (Give to IT to upload to Salesforce Connected App)
#

set -e

PRIVATE_KEY_FILE="private_key.pem"
PUBLIC_CERT_FILE="public_cert.pem"
PUBLIC_KEY_FILE="public_key.pem"
CSR_FILE="cert_request.csr"

echo "🔐 Salesforce JWT Key Pair Generator"
echo "===================================="
echo ""

# Check if keys already exist
if [ -f "$PRIVATE_KEY_FILE" ] || [ -f "$PUBLIC_CERT_FILE" ]; then
    echo "⚠️  Warning: Key files already exist!"
    echo ""
    read -p "Do you want to overwrite existing keys? (yes/no): " -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        echo "❌ Aborted. Existing keys preserved."
        exit 1
    fi
    echo "🗑️  Removing existing keys..."
    rm -f "$PRIVATE_KEY_FILE" "$PUBLIC_CERT_FILE" "$PUBLIC_KEY_FILE" "$CSR_FILE"
fi

echo "🔑 Generating 2048-bit RSA private key..."
openssl genrsa -out "$PRIVATE_KEY_FILE" 2048

echo "📄 Extracting public key..."
openssl rsa -in "$PRIVATE_KEY_FILE" -pubout -out "$PUBLIC_KEY_FILE"

echo "📋 Creating certificate signing request..."
openssl req -new -key "$PRIVATE_KEY_FILE" -out "$CSR_FILE" \
    -subj "/C=US/ST=State/L=City/O=Organization/OU=IT/CN=SalesforceIntegration"

echo "📜 Generating self-signed certificate (valid for 10 years)..."
openssl x509 -req -days 3650 -in "$CSR_FILE" -signkey "$PRIVATE_KEY_FILE" -out "$PUBLIC_CERT_FILE"

echo "🧹 Cleaning up temporary files..."
rm -f "$CSR_FILE" "$PUBLIC_KEY_FILE"

echo ""
echo "✅ Success! Key pair generated."
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📁 Generated Files:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "  🔒 $PRIVATE_KEY_FILE"
echo "     ├─ KEEP THIS SECRET!"
echo "     ├─ Never commit to git (already in .gitignore)"
echo "     ├─ Use for SALESFORCE_PRIVATE_KEY_PATH"
echo "     └─ For dev only - production keys should be managed by IT"
echo ""
echo "  📤 $PUBLIC_CERT_FILE"
echo "     ├─ Give this to IT"
echo "     ├─ IT uploads this to Salesforce Connected App"
echo "     └─ This is safe to share (it's the public certificate)"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📋 Next Steps:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "1. 📧 Send $PUBLIC_CERT_FILE to IT"
echo ""
echo "2. 🔧 IT should:"
echo "   • Go to Salesforce Setup → App Manager"
echo "   • Edit the Connected App"
echo "   • Under 'Use digital signatures', upload $PUBLIC_CERT_FILE"
echo "   • Save and wait 2-10 minutes for changes to propagate"
echo ""
echo "3. 🔐 Set environment variables:"
echo "   export SALESFORCE_CLIENT_ID='...'        # From Connected App"
echo "   export SALESFORCE_USERNAME='...'         # Service user"
echo "   export SALESFORCE_AUD='https://test.salesforce.com'  # or login.salesforce.com"
echo "   export SALESFORCE_PRIVATE_KEY_PATH='./private_key.pem'"
echo ""
echo "4. ▶️  Run the service:"
echo "   go run main.go"
echo ""
echo "5. 🏥 Test the health check:"
echo "   curl http://localhost:8080/health"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📖 For full documentation, see: README_SALESFORCE_JWT.md"
echo ""

# Display public certificate for easy copy-paste
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📄 Public Certificate (for IT to upload):"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat "$PUBLIC_CERT_FILE"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
