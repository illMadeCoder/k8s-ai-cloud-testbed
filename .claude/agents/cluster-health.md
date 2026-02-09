---
name: cluster-health
description: Hub cluster health sweep — ArgoCD sync, pod health, Crossplane, experiments, operator status. No arguments needed.
tools: Bash
model: sonnet
---

You perform a health check sweep of the k8s-ai-cloud-testbed hub cluster. Run the following in ONE bash call, then return ONLY the output with brief recommendations for any issues found.

```bash
bash -c '
ISSUES=0

echo "========================================"
echo " Hub Cluster Health Check"
echo "========================================"
echo ""

# 1. ArgoCD Applications
echo "--- ArgoCD Applications ---"
if command -v kubectl &>/dev/null && kubectl get applications.argoproj.io -A &>/dev/null 2>&1; then
  TOTAL=$(kubectl get applications.argoproj.io -A --no-headers 2>/dev/null | wc -l)
  HEALTHY=$(kubectl get applications.argoproj.io -A --no-headers 2>/dev/null | awk "{if(\$3==\"Synced\" && \$4==\"Healthy\") print}" | wc -l)
  echo "  Total: $TOTAL | Synced+Healthy: $HEALTHY"
  if [ "$TOTAL" -ne "$HEALTHY" ]; then
    echo "  Degraded:"
    kubectl get applications.argoproj.io -A --no-headers 2>/dev/null | awk "{if(\$3!=\"Synced\" || \$4!=\"Healthy\") printf \"    %-40s sync=%-10s health=%s\n\", \$2, \$3, \$4}"
    ISSUES=$((ISSUES + TOTAL - HEALTHY))
  fi
else
  echo "  SKIP: Cannot reach ArgoCD CRDs"
  ISSUES=$((ISSUES + 1))
fi
echo ""

# 2. Pod Health (key namespaces)
echo "--- Pod Health ---"
NAMESPACES="experiment-operator-system argocd argo-workflows seaweedfs crossplane-system kyverno tailscale experiments"
for NS in $NAMESPACES; do
  if kubectl get ns "$NS" &>/dev/null 2>&1; then
    NOT_READY=$(kubectl get pods -n "$NS" --no-headers 2>/dev/null | grep -v -E "Running|Completed|Succeeded" || true)
    POD_COUNT=$(kubectl get pods -n "$NS" --no-headers 2>/dev/null | wc -l)
    BAD_COUNT=$(echo "$NOT_READY" | grep -c -v "^$" || true)
    if [ "$BAD_COUNT" -gt 0 ]; then
      echo "  $NS: $BAD_COUNT/$POD_COUNT pods not ready"
      echo "$NOT_READY" | while IFS= read -r line; do
        [ -n "$line" ] && echo "    $line"
      done
      ISSUES=$((ISSUES + BAD_COUNT))
    else
      echo "  $NS: $POD_COUNT/$POD_COUNT pods OK"
    fi
  else
    echo "  $NS: namespace not found"
  fi
done
echo ""

# 3. Crossplane Resources
echo "--- Crossplane ---"
if kubectl api-resources --api-group=apiextensions.crossplane.io &>/dev/null 2>&1; then
  XR_COUNT=$(kubectl get composite --no-headers -A 2>/dev/null | wc -l)
  CLAIM_COUNT=$(kubectl get claim --no-headers -A 2>/dev/null | wc -l)
  echo "  Composite resources: $XR_COUNT | Claims: $CLAIM_COUNT"
  # Check for non-ready composites
  NOT_READY_XR=$(kubectl get composite -A --no-headers 2>/dev/null | grep -v "True" || true)
  BAD_XR=$(echo "$NOT_READY_XR" | grep -c -v "^$" || true)
  if [ "$BAD_XR" -gt 0 ]; then
    echo "  Non-ready composites:"
    echo "$NOT_READY_XR" | while IFS= read -r line; do
      [ -n "$line" ] && echo "    $line"
    done
    ISSUES=$((ISSUES + BAD_XR))
  fi
else
  echo "  SKIP: Crossplane CRDs not available"
fi
echo ""

# 4. Active Experiments
echo "--- Experiments ---"
if kubectl get experiments -A &>/dev/null 2>&1; then
  kubectl get experiments -A --no-headers 2>/dev/null | while IFS= read -r line; do
    [ -n "$line" ] && echo "  $line"
  done
  EXP_COUNT=$(kubectl get experiments -A --no-headers 2>/dev/null | wc -l)
  [ "$EXP_COUNT" -eq 0 ] && echo "  No active experiments"
  FAILED=$(kubectl get experiments -A --no-headers 2>/dev/null | grep -c "Failed" || true)
  ISSUES=$((ISSUES + FAILED))
else
  echo "  SKIP: Experiment CRDs not available"
fi
echo ""

# 5. Operator Deployment
echo "--- Operator ---"
if kubectl get deployment experiment-operator-controller-manager -n experiment-operator-system &>/dev/null 2>&1; then
  READY=$(kubectl get deployment experiment-operator-controller-manager -n experiment-operator-system -o jsonpath="{.status.readyReplicas}" 2>/dev/null)
  DESIRED=$(kubectl get deployment experiment-operator-controller-manager -n experiment-operator-system -o jsonpath="{.spec.replicas}" 2>/dev/null)
  IMAGE=$(kubectl get deployment experiment-operator-controller-manager -n experiment-operator-system -o jsonpath="{.spec.template.spec.containers[0].image}" 2>/dev/null)
  echo "  Replicas: ${READY:-0}/${DESIRED:-1} ready"
  echo "  Image: $IMAGE"
  if [ "${READY:-0}" -lt "${DESIRED:-1}" ]; then
    ISSUES=$((ISSUES + 1))
  fi
else
  echo "  SKIP: Operator deployment not found"
  ISSUES=$((ISSUES + 1))
fi
echo ""

# Summary
echo "========================================"
if [ "$ISSUES" -eq 0 ]; then
  echo " All systems operational"
else
  echo " Issues found: $ISSUES"
fi
echo "========================================"
'
```

Return ONLY the output of that script. If issues are found, add 1-2 sentences of recommendations after the output.
