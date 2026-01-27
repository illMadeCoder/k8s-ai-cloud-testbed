#!/bin/bash
# OpenBao Unseal and Seed Script
# Reads credentials from ~/.secrets/hub-secrets.env and configures OpenBao

set -e

SECRETS_FILE="${HOME}/.secrets/hub-secrets.env"

if [[ ! -f "$SECRETS_FILE" ]]; then
    echo "Error: Secrets file not found at $SECRETS_FILE"
    echo "Create it with OPENBAO_UNSEAL_KEY and OPENBAO_ROOT_TOKEN"
    exit 1
fi

source "$SECRETS_FILE"

# Check required vars
if [[ -z "$OPENBAO_UNSEAL_KEY" || -z "$OPENBAO_ROOT_TOKEN" ]]; then
    echo "Error: OPENBAO_UNSEAL_KEY and OPENBAO_ROOT_TOKEN must be set"
    exit 1
fi

echo "==> Waiting for OpenBao pod..."
kubectl wait --for=condition=Ready pod/openbao-0 -n openbao --timeout=300s 2>/dev/null || {
    echo "OpenBao pod not ready. Is the cluster running?"
    exit 1
}

echo "==> Checking seal status..."
SEALED=$(kubectl exec -n openbao openbao-0 -- bao status -format=json 2>/dev/null | jq -r '.sealed' || echo "true")

if [[ "$SEALED" == "true" ]]; then
    echo "==> Unsealing OpenBao..."
    kubectl exec -n openbao openbao-0 -- bao operator unseal "$OPENBAO_UNSEAL_KEY"
else
    echo "==> OpenBao already unsealed"
fi

echo "==> Creating ExternalSecrets token..."
kubectl create secret generic openbao-token -n external-secrets \
    --from-literal=token="$OPENBAO_ROOT_TOKEN" \
    --dry-run=client -o yaml | kubectl apply -f -

echo "==> Enabling KV secrets engine..."
kubectl exec -n openbao openbao-0 -- sh -c "
    export BAO_TOKEN='$OPENBAO_ROOT_TOKEN'
    bao secrets enable -path=secret -version=2 kv 2>/dev/null || echo 'KV engine already enabled'
"

echo "==> Seeding secrets..."

# Tailscale
if [[ -n "$TAILSCALE_CLIENT_ID" && -n "$TAILSCALE_CLIENT_SECRET" ]]; then
    echo "    - tailscale"
    kubectl exec -n openbao openbao-0 -- sh -c "
        export BAO_TOKEN='$OPENBAO_ROOT_TOKEN'
        bao kv put secret/tailscale client_id='$TAILSCALE_CLIENT_ID' client_secret='$TAILSCALE_CLIENT_SECRET'
    " >/dev/null
fi

# Cloudflare
if [[ -n "$CLOUDFLARE_API_TOKEN" ]]; then
    echo "    - cloudflare"
    kubectl exec -n openbao openbao-0 -- sh -c "
        export BAO_TOKEN='$OPENBAO_ROOT_TOKEN'
        bao kv put secret/cloudflare api_token='$CLOUDFLARE_API_TOKEN'
    " >/dev/null
fi

# AWS
if [[ -n "$AWS_ACCESS_KEY_ID" && -n "$AWS_SECRET_ACCESS_KEY" ]]; then
    echo "    - cloud/aws"
    kubectl exec -n openbao openbao-0 -- sh -c "
        export BAO_TOKEN='$OPENBAO_ROOT_TOKEN'
        bao kv put secret/cloud/aws access_key_id='$AWS_ACCESS_KEY_ID' secret_access_key='$AWS_SECRET_ACCESS_KEY'
    " >/dev/null
fi

# Azure
if [[ -n "$AZURE_CLIENT_ID" && -n "$AZURE_CLIENT_SECRET" && -n "$AZURE_TENANT_ID" && -n "$AZURE_SUBSCRIPTION_ID" ]]; then
    echo "    - cloud/azure"
    AZURE_CREDS=$(cat <<EOF
{
  "clientId": "$AZURE_CLIENT_ID",
  "clientSecret": "$AZURE_CLIENT_SECRET",
  "subscriptionId": "$AZURE_SUBSCRIPTION_ID",
  "tenantId": "$AZURE_TENANT_ID"
}
EOF
)
    kubectl exec -n openbao openbao-0 -- sh -c "
        export BAO_TOKEN='$OPENBAO_ROOT_TOKEN'
        bao kv put secret/cloud/azure credentials='$AZURE_CREDS'
    " >/dev/null
fi

# GCP
if [[ -n "$GCP_CREDENTIALS_JSON" ]]; then
    echo "    - cloud/gcp"
    kubectl exec -n openbao openbao-0 -- sh -c "
        export BAO_TOKEN='$OPENBAO_ROOT_TOKEN'
        bao kv put secret/cloud/gcp credentials='$GCP_CREDENTIALS_JSON'
    " >/dev/null
fi

echo "==> Triggering ExternalSecrets refresh..."
for es in $(kubectl get externalsecret -A -o jsonpath='{range .items[*]}{.metadata.namespace}/{.metadata.name}{"\n"}{end}'); do
    ns=$(echo "$es" | cut -d/ -f1)
    name=$(echo "$es" | cut -d/ -f2)
    kubectl annotate externalsecret -n "$ns" "$name" force-sync="$(date +%s)" --overwrite >/dev/null 2>&1 || true
done

echo "==> Done! Check ExternalSecrets status:"
kubectl get externalsecret -A
