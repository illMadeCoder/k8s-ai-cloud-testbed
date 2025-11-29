#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Installing Gateway API CRDs..."
# Install the Gateway API CRDs (standard across all implementations)
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.0/standard-install.yaml

echo ""
echo "Installing Envoy Gateway..."
# Envoy Gateway uses OCI registry for Helm chart
helm install eg oci://docker.io/envoyproxy/gateway-helm \
  --version v1.2.4 \
  --namespace envoy-gateway-system \
  --create-namespace \
  --values "$SCRIPT_DIR/values.yaml" \
  --wait

echo ""
echo "Waiting for Envoy Gateway to be ready..."
kubectl wait --for=condition=Available deployment/envoy-gateway -n envoy-gateway-system --timeout=120s

echo ""
echo "Creating default Gateway..."
kubectl apply -f "$SCRIPT_DIR/default-gateway.yaml"

echo ""
echo "=============================================="
echo "Gateway API with Envoy Gateway is ready!"
echo ""
echo "GatewayClass:"
kubectl get gatewayclass
echo ""
echo "Gateway:"
kubectl get gateway -n default
echo ""
echo "To expose a service, create an HTTPRoute:"
echo ""
echo "  apiVersion: gateway.networking.k8s.io/v1"
echo "  kind: HTTPRoute"
echo "  metadata:"
echo "    name: my-app"
echo "  spec:"
echo "    parentRefs:"
echo "      - name: lab-gateway"
echo "    hostnames:"
echo "      - \"my-app.local\""
echo "    rules:"
echo "      - matches:"
echo "          - path:"
echo "              type: PathPrefix"
echo "              value: /"
echo "        backendRefs:"
echo "          - name: my-app-service"
echo "            port: 80"
echo ""
echo "=============================================="
