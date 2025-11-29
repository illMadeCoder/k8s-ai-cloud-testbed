#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Installing cert-manager..."

# Add Jetstack Helm repo
helm repo add jetstack https://charts.jetstack.io 2>/dev/null || true
helm repo update

# Install cert-manager
helm upgrade --install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --values "$SCRIPT_DIR/values.yaml" \
  --wait --timeout 5m

echo ""
echo "Waiting for cert-manager to be ready..."
kubectl wait --for=condition=available --timeout=300s \
  deployment/cert-manager -n cert-manager
kubectl wait --for=condition=available --timeout=300s \
  deployment/cert-manager-webhook -n cert-manager
kubectl wait --for=condition=available --timeout=300s \
  deployment/cert-manager-cainjector -n cert-manager

echo ""
echo "Creating certificate issuers..."
kubectl apply -f "$SCRIPT_DIR/issuers.yaml"

echo ""
echo "=========================================="
echo "cert-manager is ready!"
echo ""
echo "Available ClusterIssuers:"
kubectl get clusterissuers
echo ""
echo "To request a certificate, create a Certificate resource:"
echo ""
cat <<'EOF'
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: my-app-tls
  namespace: default
spec:
  secretName: my-app-tls-secret
  issuerRef:
    name: selfsigned-issuer
    kind: ClusterIssuer
  dnsNames:
    - my-app.local
EOF
echo ""
echo "Or add annotations to your Ingress:"
echo "  cert-manager.io/cluster-issuer: selfsigned-issuer"
echo ""
echo "=========================================="
