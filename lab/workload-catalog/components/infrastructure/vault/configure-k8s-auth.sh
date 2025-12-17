#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
KEYS_FILE="$SCRIPT_DIR/vault-keys.json"

if [ ! -f "$KEYS_FILE" ]; then
  echo "Error: $KEYS_FILE not found. Run install.sh first."
  exit 1
fi

ROOT_TOKEN=$(jq -r '.root_token' "$KEYS_FILE")

echo "Configuring Kubernetes auth in Vault..."

# Enable Kubernetes auth method
kubectl exec -n vault vault-0 -- sh -c "
  export VAULT_TOKEN='$ROOT_TOKEN'

  # Enable k8s auth if not already enabled
  vault auth enable kubernetes 2>/dev/null || echo 'Kubernetes auth already enabled'

  # Configure it to talk to the Kubernetes API
  vault write auth/kubernetes/config \
    kubernetes_host=\"https://\$KUBERNETES_PORT_443_TCP_ADDR:443\"
"

echo ""
echo "Kubernetes auth enabled!"
echo ""
echo "Next steps to use Vault with your workloads:"
echo ""
echo "1. Create a policy (e.g., for an app called 'myapp'):"
echo "   vault policy write myapp - <<EOF"
echo "   path \"secret/data/myapp/*\" {"
echo "     capabilities = [\"read\"]"
echo "   }"
echo "   EOF"
echo ""
echo "2. Create a role binding ServiceAccounts to the policy:"
echo "   vault write auth/kubernetes/role/myapp \\"
echo "     bound_service_account_names=myapp \\"
echo "     bound_service_account_namespaces=default \\"
echo "     policies=myapp \\"
echo "     ttl=1h"
echo ""
echo "3. Annotate your pods for Vault Agent injection:"
echo "   annotations:"
echo "     vault.hashicorp.com/agent-inject: 'true'"
echo "     vault.hashicorp.com/role: 'myapp'"
echo "     vault.hashicorp.com/agent-inject-secret-config: 'secret/data/myapp/config'"
