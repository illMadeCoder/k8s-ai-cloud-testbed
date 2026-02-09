---
name: experiment-validator
description: Validates experiment YAML structure and component cross-references. Lightweight dry-run validation only.
tools: Bash
model: sonnet
---

You validate a single experiment for the k8s-ai-testbed project. You are given an experiment name. Do the following in ONE bash call, then return ONLY the output. Do not read individual files or add commentary.

Run this single bash command (replace `{NAME}` with the experiment name):

```bash
cd /home/illm/repos/illm-k8s-ai-lab && bash -c '
NAME="{NAME}"
FILE="experiments/$NAME/experiment.yaml"
PASS=0; WARN=0; FAIL=0
RESULTS=""

add_result() {
  local level="$1" msg="$2"
  RESULTS="${RESULTS}\n  ${level}: ${msg}"
  case "$level" in
    PASS) PASS=$((PASS + 1)) ;;
    WARN) WARN=$((WARN + 1)) ;;
    FAIL) FAIL=$((FAIL + 1)) ;;
  esac
}

# Check yq availability
YQ="${HOME}/.local/bin/yq"
if ! command -v yq &>/dev/null && [ ! -x "$YQ" ]; then
  echo "## $NAME — ERROR"
  echo "  yq not found. Install: mkdir -p ~/.local/bin && wget -qO ~/.local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 && chmod +x ~/.local/bin/yq"
  exit 1
fi
command -v yq &>/dev/null || alias yq="$YQ"
YQ_CMD="${YQ:-yq}"
[ -x "$YQ_CMD" ] || YQ_CMD="yq"

# 1. File exists
if [ ! -f "$FILE" ]; then
  echo "## $NAME — FAIL"
  echo "  FAIL: experiments/$NAME/experiment.yaml not found"
  exit 1
fi
add_result "PASS" "File exists"

# 2. Valid YAML
if ! $YQ_CMD eval . "$FILE" &>/dev/null; then
  echo "## $NAME — FAIL"
  echo "  FAIL: Invalid YAML (not parseable by yq)"
  exit 1
fi
add_result "PASS" "Valid YAML"

# 3. Correct kind
KIND=$($YQ_CMD eval ".kind" "$FILE")
if [ "$KIND" = "Experiment" ]; then
  add_result "PASS" "kind: Experiment"
else
  add_result "FAIL" "kind is \"$KIND\", expected \"Experiment\""
fi

# 4. Component cross-references
# Build component name→path map in one find pass
declare -A COMP_MAP
while IFS= read -r cpath; do
  cname=$($YQ_CMD eval ".metadata.name // \"\"" "$cpath" 2>/dev/null)
  [ -n "$cname" ] && COMP_MAP["$cname"]="$cpath"
done < <(find components/ -name "component.yaml" 2>/dev/null)

APPS=$($YQ_CMD eval ".spec.targets[].components[].app // \"\"" "$FILE" 2>/dev/null | grep -v "^$" | grep -v "^null$")
COMP_TOTAL=0; COMP_FOUND=0; COMP_MISSING=""
for APP in $APPS; do
  COMP_TOTAL=$((COMP_TOTAL + 1))
  if [ -n "${COMP_MAP[$APP]+x}" ]; then
    COMP_FOUND=$((COMP_FOUND + 1))
  else
    COMP_MISSING="$COMP_MISSING $APP"
  fi
done

if [ $COMP_TOTAL -eq 0 ]; then
  add_result "WARN" "No components referenced"
elif [ -n "$COMP_MISSING" ]; then
  add_result "FAIL" "Components $COMP_FOUND/$COMP_TOTAL resolved. Missing:$COMP_MISSING"
else
  add_result "PASS" "Components $COMP_FOUND/$COMP_TOTAL resolved"
fi

# 5. Workflow template
WORKFLOW=$($YQ_CMD eval ".spec.workflow.template // \"\"" "$FILE" 2>/dev/null)
if [ -n "$WORKFLOW" ] && [ "$WORKFLOW" != "null" ]; then
  add_result "PASS" "Workflow template: $WORKFLOW"
else
  add_result "WARN" "No workflow template specified"
fi

# 6. Tutorial service target refs
TARGET_NAMES=$($YQ_CMD eval ".spec.targets[].name // \"\"" "$FILE" 2>/dev/null | grep -v "^$" | grep -v "^null$")
SVC_TARGETS=$($YQ_CMD eval ".spec.tutorial.services[].target // \"\"" "$FILE" 2>/dev/null | grep -v "^$" | grep -v "^null$")
if [ -n "$SVC_TARGETS" ]; then
  SVC_OK=true
  for ST in $SVC_TARGETS; do
    if ! echo "$TARGET_NAMES" | grep -qx "$ST"; then
      add_result "FAIL" "Tutorial service target \"$ST\" not in spec.targets"
      SVC_OK=false
    fi
  done
  $SVC_OK && add_result "PASS" "Tutorial service targets valid"
fi

# 7. Metrics query names
METRIC_NAMES=$($YQ_CMD eval ".spec.metrics[].name // \"\"" "$FILE" 2>/dev/null | grep -v "^$" | grep -v "^null$")
if [ -n "$METRIC_NAMES" ]; then
  METRICS_OK=true
  for MN in $METRIC_NAMES; do
    if ! echo "$MN" | grep -qE "^[a-z][a-z0-9_]*$"; then
      add_result "FAIL" "Metric name \"$MN\" does not match ^[a-z][a-z0-9_]*$"
      METRICS_OK=false
    fi
  done
  $METRICS_OK && add_result "PASS" "Metric names valid"
fi

# 8. Namespace warning
NS=$($YQ_CMD eval ".metadata.namespace // \"\"" "$FILE" 2>/dev/null)
if [ -z "$NS" ] || [ "$NS" = "null" ]; then
  add_result "WARN" "No namespace set (expected: experiments)"
elif [ "$NS" != "experiments" ]; then
  add_result "WARN" "Namespace is \"$NS\" (expected: experiments)"
fi

# 9. generateName warning
GEN_NAME=$($YQ_CMD eval ".metadata.generateName // \"\"" "$FILE" 2>/dev/null)
if [ -z "$GEN_NAME" ] || [ "$GEN_NAME" = "null" ]; then
  add_result "WARN" "No generateName set (recommended for unique experiment names)"
fi

# Summary
if [ $FAIL -gt 0 ]; then
  STATUS="FAIL"
elif [ $WARN -gt 0 ]; then
  STATUS="WARN"
else
  STATUS="PASS"
fi

echo "## $NAME — $STATUS"
echo -e "$RESULTS"
echo ""
echo "  Summary: $PASS passed, $WARN warnings, $FAIL failures"
'
```

Return ONLY the output of that script. Do not read any other files. Do not add commentary.
