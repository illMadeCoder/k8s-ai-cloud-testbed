#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Installing Prometheus & Grafana (kube-prometheus-stack)..."

# Add Prometheus community Helm repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts 2>/dev/null || true
helm repo update

# Install kube-prometheus-stack
helm upgrade --install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace observability \
  --create-namespace \
  --values "$SCRIPT_DIR/values.yaml" \
  --wait --timeout 5m

echo ""
echo "Prometheus & Grafana installed!"
echo ""
echo "Access Grafana:"
echo "  kubectl port-forward svc/kube-prometheus-stack-grafana -n observability 3000:80"
echo "  http://localhost:3000"
echo "  Username: admin"
echo "  Password: admin"
echo ""
echo "Access Prometheus:"
echo "  kubectl port-forward svc/kube-prometheus-stack-prometheus -n observability 9090:9090"
echo "  http://localhost:9090"
echo ""
echo "Pre-configured dashboards:"
echo "  - Kubernetes cluster monitoring"
echo "  - Node exporter metrics"
echo "  - Pod/container metrics"
echo ""
