#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

usage() {
  echo "Usage: $0 [--help]"
  echo ""
  echo "Bootstrap the Kubernetes lab environment with ArgoCD."
  echo "All workloads are managed by ArgoCD via GitOps from the git repository."
  echo ""
  echo "This script:"
  echo "  1. Installs ArgoCD"
  echo "  2. Creates the app-of-apps Application to bootstrap all other services"
  echo ""
  echo "The following components will be automatically deployed by ArgoCD:"
  echo "  - cert-manager (TLS certificate management)"
  echo "  - Gateway API with Envoy Gateway (ingress controller)"
  echo "  - Prometheus & Grafana (observability)"
  echo "  - Backstage (developer portal)"
  echo "  - HashiCorp Vault (secrets management)"
  echo "  - Demo application"
  echo ""
  echo "For more information, see manifests/applications/ and manifests/workloads/"
}

install_argocd() {
  echo "========================================"
  echo "Installing ArgoCD..."
  echo "========================================"

  helm repo add argo https://argoproj.github.io/argo-helm 2>/dev/null || true
  helm repo update

  helm install argocd argo/argo-cd \
    --namespace argocd \
    --create-namespace \
    --values "$SCRIPT_DIR/argocd/values.yaml" \
    --wait

  echo ""
  echo "ArgoCD installed!"
  echo "Password: $(kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d)"
  echo ""
  echo "To access:"
  echo "  kubectl port-forward svc/argocd-server -n argocd 8080:443"
  echo "  https://localhost:8080 (admin)"
  echo ""
}

bootstrap_applications() {
  echo "========================================"
  echo "Bootstrapping ArgoCD Applications..."
  echo "========================================"
  echo ""
  echo "Creating app-of-apps Application to manage all workloads..."

  kubectl apply -f "$SCRIPT_DIR/../manifests/applications/app-of-apps.yaml"

  echo ""
  echo "Waiting for ArgoCD to discover and sync applications..."
  sleep 5

  echo ""
  echo "Applications will sync automatically. Monitor progress with:"
  echo "  kubectl get applications -n argocd"
  echo "  kubectl get applications -n argocd -o wide"
  echo ""
}

# Parse arguments
case "${1:-}" in
  -h|--help)
    usage
    exit 0
    ;;
  "")
    # No arguments - proceed with install
    ;;
  *)
    echo "Unknown option: $1"
    usage
    exit 1
    ;;
esac

install_argocd
bootstrap_applications

echo ""
echo "========================================"
echo "Bootstrap complete!"
echo "========================================"
echo ""
echo "All infrastructure is now managed by ArgoCD."
echo "Monitor application status with:"
echo "  kubectl get applications -n argocd"
echo ""
