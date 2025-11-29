#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Installing Backstage..."

# Add Backstage Helm repo
helm repo add backstage https://backstage.github.io/charts 2>/dev/null || true
helm repo update

# Install Backstage
helm upgrade --install backstage backstage/backstage \
  --namespace backstage \
  --create-namespace \
  --values "$SCRIPT_DIR/values.yaml" \
  --wait --timeout 5m

echo ""
echo "Backstage installed!"
echo ""
echo "To access Backstage:"
echo "  kubectl port-forward svc/backstage -n backstage 7007:7007"
echo "  http://localhost:7007"
echo ""
echo "Note: This is a vanilla demo instance. For production use,"
echo "build a custom image with your required plugins."
echo ""
