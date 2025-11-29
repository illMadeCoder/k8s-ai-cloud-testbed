#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KEYS_FILE="$SCRIPT_DIR/vault-keys.json"

echo "Adding HashiCorp Helm repo..."
helm repo add hashicorp https://helm.releases.hashicorp.com
helm repo update

echo "Installing Vault..."
helm install vault hashicorp/vault \
  --namespace vault \
  --create-namespace \
  --values "$SCRIPT_DIR/values.yaml"

echo "Waiting for Vault pod to be ready..."
kubectl wait --for=condition=Ready pod/vault-0 -n vault --timeout=120s || true

# Check if Vault is already initialized
INIT_STATUS=$(kubectl exec -n vault vault-0 -- vault status -format=json 2>/dev/null | jq -r '.initialized' || echo "false")

if [ "$INIT_STATUS" = "true" ]; then
  echo "Vault is already initialized."
  if [ -f "$KEYS_FILE" ]; then
    echo "Found existing keys file, attempting unseal..."
    UNSEAL_KEY=$(jq -r '.unseal_keys_b64[0]' "$KEYS_FILE")
    kubectl exec -n vault vault-0 -- vault operator unseal "$UNSEAL_KEY" || true
  fi
else
  echo "Initializing Vault..."
  # Initialize with 1 key share and 1 threshold for lab simplicity
  # Production would use 5 shares with 3 threshold
  kubectl exec -n vault vault-0 -- vault operator init \
    -key-shares=1 \
    -key-threshold=1 \
    -format=json > "$KEYS_FILE"

  echo ""
  echo "=============================================="
  echo "IMPORTANT: Vault keys saved to:"
  echo "  $KEYS_FILE"
  echo ""
  echo "DO NOT commit this file to git!"
  echo "=============================================="
  echo ""

  # Unseal Vault
  UNSEAL_KEY=$(jq -r '.unseal_keys_b64[0]' "$KEYS_FILE")
  kubectl exec -n vault vault-0 -- vault operator unseal "$UNSEAL_KEY"
fi

echo ""
echo "Vault status:"
kubectl exec -n vault vault-0 -- vault status || true

ROOT_TOKEN=$(jq -r '.root_token' "$KEYS_FILE" 2>/dev/null || echo "check $KEYS_FILE")

echo ""
echo "=============================================="
echo "Vault is ready!"
echo ""
echo "Root Token: $ROOT_TOKEN"
echo ""
echo "To access the UI:"
echo "  kubectl port-forward svc/vault -n vault 8200:8200"
echo "  Then open: http://localhost:8200"
echo ""
echo "To use the CLI:"
echo "  export VAULT_ADDR='http://127.0.0.1:8200'"
echo "  export VAULT_TOKEN='$ROOT_TOKEN'"
echo "=============================================="
