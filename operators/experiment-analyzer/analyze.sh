#!/usr/bin/env bash
set -euo pipefail

# Multi-pass AI experiment analyzer
#
# Produces research-paper quality analysis via 5 focused claude -p calls:
#   Pass 1: Analysis plan (technologies, focus areas, domain context)
#   Pass 2: Core analysis (abstract, targetAnalysis, performanceAnalysis, metricInsights, architectureDiagram)
#   Pass 3: FinOps + SecOps analysis
#   Pass 4: Capabilities matrix + feedback
#   Pass 5: Body synthesis (rich narrative with typed content blocks)
#
# Required environment variables:
#   EXPERIMENT_NAME  - experiment name (used for S3 key and GitHub path)
#   S3_ENDPOINT      - SeaweedFS S3 endpoint (e.g. seaweedfs-s3.seaweedfs.svc.cluster.local:8333)
#   GITHUB_TOKEN     - GitHub PAT for committing results
#   GITHUB_REPO      - GitHub repo (e.g. illMadeCoder/k8s-ai-cloud-testbed)
#
# Required file mount:
#   ~/.claude/.credentials.json - Claude Code credentials file (with refresh token),
#     mounted from the claude-auth K8s Secret. Enables auto-refresh of expired access tokens.
#
# Optional:
#   GITHUB_BRANCH    - branch to commit to (REQUIRED — no default, refuses to commit to main)
#   GITHUB_RESULTS_PATH - path in repo for result JSONs (default: site/data)

: "${EXPERIMENT_NAME:?EXPERIMENT_NAME is required}"
: "${S3_ENDPOINT:?S3_ENDPOINT is required}"

# Verify Claude Code credentials file is mounted
CLAUDE_CREDS="${HOME}/.claude/.credentials.json"
if [ ! -f "${CLAUDE_CREDS}" ]; then
  echo "ERROR: Claude Code credentials file not found at ${CLAUDE_CREDS}"
  echo "The claude-auth secret must be mounted as a volume with credentials.json"
  exit 1
fi
echo "==> Claude Code credentials file found at ${CLAUDE_CREDS}"

# Pre-flight auth check — triggers token refresh if access token is expired
echo "==> Pre-flight auth check (triggers token refresh if needed)..."
if ! claude -p "respond with ok" --output-format json > /dev/null 2>&1; then
  echo "ERROR: Claude CLI auth failed — credentials may need manual refresh in OpenBao"
  echo "Run: bao kv put secret/experiment-operator/claude-auth credentials=\"\$(cat ~/.claude/.credentials.json)\""
  exit 1
fi
echo "==> Auth check passed"

# Safety: GITHUB_BRANCH must be explicitly set by the operator (to experiment/{name}).
# Refusing to default to main prevents accidental commits to the production branch.
SKIP_GITHUB_COMMIT=""
if [ -z "${GITHUB_BRANCH:-}" ]; then
  echo "WARNING: GITHUB_BRANCH not set — refusing to default to main"
  echo "The operator must set GITHUB_BRANCH to the experiment branch"
  echo "Analysis will still run; results go to S3 only (no GitHub commit)"
  SKIP_GITHUB_COMMIT=true
fi
GITHUB_RESULTS_PATH="${GITHUB_RESULTS_PATH:-site/data}"

S3_URL="http://${S3_ENDPOINT}/experiment-results/${EXPERIMENT_NAME}/summary.json"
WORK_DIR="$(mktemp -d)"
SUMMARY_FILE="${WORK_DIR}/summary.json"
ENRICHED_FILE="${WORK_DIR}/enriched.json"

cleanup() { rm -rf "${WORK_DIR}"; }
trap cleanup EXIT

echo "==> Fetching summary.json from S3: ${S3_URL}"
if ! curl -sf "${S3_URL}" -o "${SUMMARY_FILE}"; then
  echo "ERROR: Failed to fetch summary.json from S3"
  exit 1
fi

echo "==> Summary fetched ($(wc -c < "${SUMMARY_FILE}") bytes)"

# --- Extract requested analysis sections from summary.json ---
REQUESTED_SECTIONS=""
# Try new field first, fall back to old
if jq -e '.analyzerConfig.sections' "${SUMMARY_FILE}" > /dev/null 2>&1; then
  REQUESTED_SECTIONS=$(jq -r '.analyzerConfig.sections | join(",")' "${SUMMARY_FILE}")
  echo "==> Requested analysis sections: ${REQUESTED_SECTIONS}"
elif jq -e '.analysisConfig.sections' "${SUMMARY_FILE}" > /dev/null 2>&1; then
  REQUESTED_SECTIONS=$(jq -r '.analysisConfig.sections | join(",")' "${SUMMARY_FILE}")
  echo "==> Requested analysis sections: ${REQUESTED_SECTIONS} (legacy analysisConfig field)"
else
  echo "==> No analyzerConfig.sections in summary.json — skipping analysis"
  echo "==> To enable analysis, add spec.analyzerConfig.sections to the Experiment CR"
  exit 0
fi

# --- Helper: check if a section is requested ---
section_requested() {
  local section="$1"
  echo ",${REQUESTED_SECTIONS}," | grep -q ",${section},"
}

# --- Helper: check if ANY of the given sections are requested ---
any_section_requested() {
  for section in "$@"; do
    if section_requested "${section}"; then
      return 0
    fi
  done
  return 1
}

# --- Helper: extract JSON from claude --output-format json (JSONL) response ---
extract_json() {
  local raw_file="$1"
  local out_file="$2"

  # Claude --output-format json outputs JSONL. The final "result" object has the text.
  if RESULT_TEXT=$(jq -rs '[.[] | select(.type == "result")] | last | .result' "${raw_file}" 2>/dev/null) \
     && [ -n "${RESULT_TEXT}" ] && [ "${RESULT_TEXT}" != "null" ]; then
    printf '%s\n' "${RESULT_TEXT}" > "${out_file}"
  elif jq -e '.[0].text' "${raw_file}" > /dev/null 2>&1; then
    jq -r '.[0].text' "${raw_file}" > "${out_file}"
  else
    cp "${raw_file}" "${out_file}"
  fi

  # Strip markdown fences if present (handles leading blank lines, ```json tag, trailing whitespace)
  # Remove leading blank lines, then check for opening fence
  sed -i '/\S/,$!d' "${out_file}"
  if head -1 "${out_file}" | grep -qE '^\s*```'; then
    sed -i '1d' "${out_file}"
  fi
  # Remove trailing blank lines, then check for closing fence
  sed -i -e :a -e '/^\s*$/{ $d; N; ba; }' "${out_file}"
  if tail -1 "${out_file}" | grep -qE '^\s*```'; then
    sed -i '$d' "${out_file}"
  fi

  # Validate JSON — if invalid, try stripping preamble text before first {
  if ! jq empty "${out_file}" 2>/dev/null; then
    if grep -qn '^{' "${out_file}"; then
      echo "WARNING: Stripping preamble text before JSON object"
      sed -i '1,/^{/{/^{/!d}' "${out_file}"
      # Also strip any trailing text after the JSON (after last })
      local last_brace
      last_brace=$(grep -n '^}' "${out_file}" | tail -1 | cut -d: -f1)
      if [ -n "${last_brace}" ]; then
        sed -i "$((last_brace + 1)),\$d" "${out_file}"
      fi
    fi
    if ! jq empty "${out_file}" 2>/dev/null; then
      echo "ERROR: Output is not valid JSON"
      cat "${out_file}" >&2
      return 1
    fi
  fi
}

# --- Helper: run a single analysis pass with retry ---
# Args: pass_name, prompt_file (path to file containing the prompt), out_file
run_pass() {
  local pass_name="$1"
  local prompt_file="$2"
  local out_file="$3"
  local raw_file="${WORK_DIR}/${pass_name}_raw.json"
  local stderr_file="${WORK_DIR}/${pass_name}_stderr.log"

  echo "==> Running ${pass_name}..."

  local attempt
  for attempt in 1 2; do
    if claude -p --output-format json < "${prompt_file}" > "${raw_file}" 2>"${stderr_file}"; then
      if extract_json "${raw_file}" "${out_file}"; then
        echo "==> ${pass_name} complete ($(wc -c < "${out_file}") bytes)"
        return 0
      fi
    fi
    if grep -q "expired\|authentication_error" "${stderr_file}" 2>/dev/null || \
       grep -q "expired\|authentication_error" "${raw_file}" 2>/dev/null; then
      echo "ERROR: Authentication failure in ${pass_name} — token may be expired"
    fi
    if [ "${attempt}" -eq 1 ]; then
      echo "WARNING: ${pass_name} attempt 1 failed, retrying..."
      sleep 2
    fi
  done

  echo "WARNING: ${pass_name} failed after 2 attempts — section will be null"
  echo '{}' > "${out_file}"
  return 1
}

# --- Helper: validate Mermaid diagram structure ---
# Returns non-empty string describing issues if validation fails
validate_diagram() {
  local diagram="$1"
  local issues=""

  # Count nodes: lines matching id["..."] or id(...) or id[...] patterns (excluding subgraph/end lines)
  local node_count
  node_count=$(printf '%s\n' "${diagram}" | grep -cE '^\s*[a-zA-Z_][a-zA-Z0-9_]*(\["|(\("|{")|\["|\[)' 2>/dev/null || echo 0)
  if [ "${node_count}" -gt 10 ]; then
    issues="${issues}Has ${node_count} nodes (max 10). "
  elif [ "${node_count}" -gt 8 ]; then
    issues="${issues}Has ${node_count} nodes (recommended max 8). "
  fi

  # Count subgraphs
  local subgraph_count
  subgraph_count=$(printf '%s\n' "${diagram}" | grep -cE '^\s*subgraph\b' 2>/dev/null || echo 0)
  if [ "${subgraph_count}" -gt 4 ]; then
    issues="${issues}Has ${subgraph_count} subgraphs (max 4). "
  fi

  # Check for nested subgraphs (subgraph before its corresponding end)
  local depth=0 max_depth=0
  while IFS= read -r line; do
    if echo "${line}" | grep -qE '^\s*subgraph\b'; then
      depth=$((depth + 1))
      [ "${depth}" -gt "${max_depth}" ] && max_depth=${depth}
    elif echo "${line}" | grep -qE '^\s*end\s*$'; then
      [ "${depth}" -gt 0 ] && depth=$((depth - 1))
    fi
  done <<< "${diagram}"
  if [ "${max_depth}" -gt 1 ]; then
    issues="${issues}Nested subgraphs (depth ${max_depth}, max 1). "
  fi

  # Check longest label (text inside ["..."] or ["...<br/>..."])
  local longest_label
  longest_label=$(printf '%s\n' "${diagram}" | grep -oE '\["[^"]*"\]' | sed 's/\["//;s/"\]//' | sed 's/<br\/>/\n/g' | awk '{ if (length > max) max = length } END { print max+0 }')
  if [ "${longest_label}" -gt 30 ]; then
    issues="${issues}Longest label is ${longest_label} chars (max 30). "
  fi

  printf '%s' "${issues}"
}

# --- Helper: check diagram rendered width via mmdc ---
check_diagram_width() {
  local diagram="$1" max_width="${2:-900}"

  # Skip if mmdc not available
  if ! command -v mmdc > /dev/null 2>&1; then
    return 0
  fi

  printf '%s\n' "${diagram}" > "${WORK_DIR}/diagram_check.mmd"
  if ! mmdc -i "${WORK_DIR}/diagram_check.mmd" -o "${WORK_DIR}/diagram_check.svg" \
       -t dark -b transparent --puppeteerConfigFile /dev/null 2>/dev/null; then
    echo "WARNING: mmdc render failed — skipping width check" >&2
    return 0
  fi

  local svg_width
  svg_width=$(grep -oP 'viewBox="[0-9.]+ [0-9.]+ \K[0-9.]+' "${WORK_DIR}/diagram_check.svg" 2>/dev/null | head -1)
  if [ -n "${svg_width}" ] && [ "${svg_width%.*}" -gt "${max_width}" ]; then
    printf '%s' "${svg_width%.*}"
    return 1
  fi
  return 0
}

# --- Helper: build a prompt file from strings and files (avoids ARG_MAX limits) ---
# Args: output_path, then pairs of (type, content) where type is "str" or "file"
build_prompt_file() {
  local output="$1"; shift
  > "${output}"
  while [ $# -ge 2 ]; do
    local type="$1" content="$2"; shift 2
    if [ "${type}" = "file" ]; then
      cat "${content}" >> "${output}"
    else
      printf '%s\n' "${content}" >> "${output}"
    fi
  done
}

# --- Extract available code snippet keys from summary.json ---
CODE_SNIPPET_KEYS=""
if jq -e '.codeSnippets' "${SUMMARY_FILE}" > /dev/null 2>&1; then
  CODE_SNIPPET_KEYS=$(jq -r '.codeSnippets | keys | join(", ")' "${SUMMARY_FILE}")
  echo "==> Code snippets found: ${CODE_SNIPPET_KEYS}"
fi

# Build conditional code placement hint for Pass 5
CODE_PLACEMENT_HINT=""
if [ -n "${CODE_SNIPPET_KEYS}" ]; then
  CODE_PLACEMENT_HINT="CODE PLACEMENT: You MUST include a \"code\" block for each of these snippet keys: ${CODE_SNIPPET_KEYS}. Place each one in the topic where its code is most relevant to the discussion. For each code block, add an \"annotations\" array (max 3) identifying the most performance-critical lines — syscalls, branching points, hot paths — with category-specific callouts that reference actual metric values. If no lines are annotation-worthy, use \"insight\" instead."
fi

# Extract hypothesis context if present (claim, questions, focus from experiment spec)
# Try new 'hypothesis' field first, fall back to legacy 'study' field
STUDY_CONTEXT=""
MACHINE_VERDICT=""
if jq -e '.hypothesis' "${SUMMARY_FILE}" > /dev/null 2>&1; then
  STUDY_HYPOTHESIS=$(jq -r '.hypothesis.claim // empty' "${SUMMARY_FILE}")
  STUDY_QUESTIONS=$(jq -r '.hypothesis.questions // [] | join("; ")' "${SUMMARY_FILE}")
  STUDY_FOCUS=$(jq -r '.hypothesis.focus // [] | join(", ")' "${SUMMARY_FILE}")
  MACHINE_VERDICT=$(jq -r '.hypothesis.machineVerdict // empty' "${SUMMARY_FILE}")
elif jq -e '.study' "${SUMMARY_FILE}" > /dev/null 2>&1; then
  STUDY_HYPOTHESIS=$(jq -r '.study.hypothesis // empty' "${SUMMARY_FILE}")
  STUDY_QUESTIONS=$(jq -r '.study.questions // [] | join("; ")' "${SUMMARY_FILE}")
  STUDY_FOCUS=$(jq -r '.study.focus // [] | join(", ")' "${SUMMARY_FILE}")
  MACHINE_VERDICT=""
fi

# Extract iteration metadata if present (quality gate re-collection)
ITERATION_CONTEXT=""
if jq -e '.iterationStatus' "${SUMMARY_FILE}" > /dev/null 2>&1; then
  ITER_CURRENT=$(jq -r '.iterationStatus.currentIteration // 0' "${SUMMARY_FILE}")
  ITER_MAX=$(jq -r '.iterationStatus.maxIterations // 0' "${SUMMARY_FILE}")
  ITER_PHASE=$(jq -r '.iterationStatus.phase // "unknown"' "${SUMMARY_FILE}")
  if [ "${ITER_CURRENT}" -gt 0 ]; then
    ITERATION_CONTEXT="
NOTE: This experiment used the metrics quality gate. Metrics were re-collected ${ITER_CURRENT} time(s)
(max ${ITER_MAX}) with progressively shorter query windows to improve data coverage.
Quality gate phase: ${ITER_PHASE}. Earlier iterations may have had empty metrics due to
rate() dilution from the full experiment duration covering pre-benchmark idle time.
"
    echo "==> Quality gate iteration context: iteration=${ITER_CURRENT}, phase=${ITER_PHASE}"
  fi
fi

if [ -n "${STUDY_HYPOTHESIS}${STUDY_QUESTIONS}${STUDY_FOCUS}" ]; then
  STUDY_CONTEXT="
STUDY CONTEXT (from the experimenter — use this to guide your analysis):
"
  [ -n "${STUDY_HYPOTHESIS}" ] && STUDY_CONTEXT="${STUDY_CONTEXT}Hypothesis: ${STUDY_HYPOTHESIS}
"
  [ -n "${STUDY_QUESTIONS}" ] && STUDY_CONTEXT="${STUDY_CONTEXT}Questions to answer: ${STUDY_QUESTIONS}
"
  [ -n "${STUDY_FOCUS}" ] && STUDY_CONTEXT="${STUDY_CONTEXT}Focus areas: ${STUDY_FOCUS}
"
  [ -n "${MACHINE_VERDICT}" ] && STUDY_CONTEXT="${STUDY_CONTEXT}Machine verdict (from success criteria evaluation): ${MACHINE_VERDICT}
"
  [ -n "${ITERATION_CONTEXT}" ] && STUDY_CONTEXT="${STUDY_CONTEXT}${ITERATION_CONTEXT}"
  echo "==> Hypothesis context found: claim=$(echo "${STUDY_HYPOTHESIS}" | head -c 80)..."
else
  echo "==> No hypothesis context in experiment spec — analyzer will infer intent"
fi

# ============================================================================
# Pass 1: Analysis Plan
# ============================================================================
PASS1_PROMPT=$(cat <<'EOF'
You are analyzing Kubernetes experiment benchmark results. Examine the experiment data and produce a JSON analysis plan.

Identify:
- The technologies being compared (if any)
- Whether this is a comparison experiment (multiple targets with different components)
- Key focus areas for deep analysis (e.g. resource efficiency, query languages, storage backends)
- Relevant domain knowledge about the technologies (e.g. "Loki uses LogQL and indexes only labels, while Elasticsearch uses Lucene and full-text indexes documents")
- The experiment domain (observability, networking, storage, cicd)

IMPORTANT: If a "STUDY CONTEXT" section is provided below the data, the experimenter has stated
their hypothesis and questions. Your analysis plan MUST prioritize these. The focusAreas should
align with the study's focus, and your domainContext should include knowledge relevant to
evaluating the hypothesis.

Output ONLY a JSON object with this structure:
{
  "technologies": ["Tech1", "Tech2"],
  "isComparison": true,
  "focusAreas": ["area1", "area2", "area3"],
  "domainContext": "Brief domain knowledge paragraph relevant to interpreting results",
  "domain": "observability"
}

Rules:
- If not a comparison, set technologies to a single-element array and isComparison to false
- focusAreas should have 2-5 entries, specific to this experiment
- domainContext should include knowledge an expert would use to interpret the metrics
- Output ONLY the JSON object, no markdown fences or extra text

Here is the experiment data:
EOF
)

PASS1_FILE="${WORK_DIR}/pass_1.json"
PASS1_PROMPT_FILE="${WORK_DIR}/pass_1_prompt.txt"
build_prompt_file "${PASS1_PROMPT_FILE}" \
  "str" "${PASS1_PROMPT}" \
  "file" "${SUMMARY_FILE}" \
  "str" "${STUDY_CONTEXT}"
run_pass "pass_1_plan" "${PASS1_PROMPT_FILE}" "${PASS1_FILE}" || true

echo "==> Analysis plan: $(jq -c '{technologies, isComparison, focusAreas}' "${PASS1_FILE}" 2>/dev/null || echo '{}')"

# ============================================================================
# Pass 2: Core Analysis (abstract, targetAnalysis, performanceAnalysis, metricInsights)
# ============================================================================
PASS2_FILE="${WORK_DIR}/pass_2.json"
if any_section_requested "abstract" "targetAnalysis" "performanceAnalysis" "metricInsights" "architectureDiagram" "vocabulary"; then

PASS2_PROMPT=$(cat <<'EOF'
You are writing research-paper quality analysis of Kubernetes experiment benchmark results.
You have an analysis plan and the full experiment data. Generate the core analysis sections.

Your analysis will be published on the Testbed Benchmarks site. Each section appears as a
styled card on the experiment detail page. Be specific with numbers from the data.

Output ONLY a JSON object with these sections:

{
  "hypothesisVerdict": "<EXACTLY one of: validated | invalidated | insufficient>",

  "abstract": "<4-6 sentence abstract. Start by stating whether the experiment conclusively validates, partially validates, or invalidates the hypothesis, and WHY. Summarize the key evidence. If the experiment was insufficient to evaluate the hypothesis (e.g. missing metrics, wrong granularity, too short), say so explicitly and what would be needed. End with the most actionable finding.>",

  "targetAnalysis": {
    "overview": "<How infrastructure choices (machine type, node count, preemptible) affect the results>",
    "perTarget": {
      "<target_name>": "<Analysis of this specific target's configuration and performance>"
    },
    "comparisonToBaseline": "<If comparison: how targets compare to each other or to baseline expectations. Null if not comparison.>"
  },

  "performanceAnalysis": {
    "overview": "<High-level performance assessment>",
    "findings": [
      "<Numbered finding 1 with specific data values>",
      "<Numbered finding 2>",
      "<Numbered finding 3>"
    ],
    "bottlenecks": ["<Identified bottleneck or limitation>"]
  },

  "metricInsights": {
    "<exact_metric_key>": "<1-2 sentence insight referencing actual values from the data. Each insight appears below its Vega-Lite chart.>"
  },

  "codeInsights": {
    "<exact_code_snippet_key>": "<1-2 sentence analysis of this code snippet's implementation and its relationship to observed metrics. Explain what the code does and why it matters for the results.>"
  },

  "architectureDiagram": "<ASCII architecture diagram string with \\n for newlines>",

  "vocabulary": [
    {"term": "<domain term>", "definition": "<1-sentence definition>"}
  ]
}

Rules for vocabulary:
- 5-10 entries covering the most important domain-specific terms used in the analysis
- Include technologies, protocols, metrics concepts, and infrastructure terms
- Definitions should be concise (1 sentence) and accessible to someone new to the domain
- Order from most fundamental to most specialized

Rules:
- "hypothesisVerdict" MUST be exactly one of the three allowed values (validated, invalidated, insufficient) — it is displayed as a status badge in the experiment header
- "abstract" is the most important section — it appears directly below the hypothesis on the experiment page
- The abstract MUST open with a verdict on the hypothesis matching hypothesisVerdict
- Explain the causal reasoning: does the data confirm WHY the hypothesis predicted this outcome?
- If the experiment design was insufficient (wrong metrics, missing isolation, too short), state what specifically was missing
- "targetAnalysis.perTarget" must have one entry per target in the experiment
- "performanceAnalysis.findings" should have 3-6 numbered findings with actual data
- If study questions exist, findings should directly answer as many as possible
- "metricInsights" must have one entry per metric key in metrics.queries, using exact key names
- "codeInsights" must have one entry per code snippet key in codeSnippets, using exact key names. If no codeSnippets exist, omit codeInsights entirely
- Reference specific numbers from the data (CPU cores, memory bytes, durations)
- Be technical and concise — this is for infrastructure engineers
- "architectureDiagram": Mermaid flowchart for an 800px-wide container.
  SYNTAX: 'flowchart TD' only. 'subgraph' for boundaries (max 3, NO nesting).
  Node: id["Short Label"]. Max 3 words per label, NO resource metrics in labels.
  Edge: source -->|label| target. Edge labels 1-2 words max.
  Max 8 nodes total. Fewer is better.
  CONTENT: Only target cluster workloads under test — omit hub cluster, operators,
  kube-state-metrics, node-exporter, ConfigMaps, Secrets, RBAC.
  Group by role (e.g. "Ingestion", "Storage", "Query").
  LAYOUT: Keep nodes vertical. Max 2 nodes at the same horizontal level per subgraph.
  For comparisons with parallel stacks, use one subgraph per stack connected from
  a shared source at top.
  Single JSON string with \n for line breaks.
- "architectureDiagramFormat": "mermaid"
- Output ONLY the JSON object, no markdown fences or extra text

ANALYSIS PLAN:
EOF
)

PASS2_PROMPT_FILE="${WORK_DIR}/pass_2_prompt.txt"
build_prompt_file "${PASS2_PROMPT_FILE}" \
  "str" "${PASS2_PROMPT}" \
  "file" "${PASS1_FILE}" \
  "str" "${STUDY_CONTEXT}" \
  "str" "EXPERIMENT DATA:" \
  "file" "${SUMMARY_FILE}"
run_pass "pass_2_core" "${PASS2_PROMPT_FILE}" "${PASS2_FILE}" || true

# --- Diagram validation + retry ---
if section_requested "architectureDiagram"; then
  DIAGRAM_TEXT=$(jq -r '.architectureDiagram // ""' "${PASS2_FILE}" 2>/dev/null)
  if [ -n "${DIAGRAM_TEXT}" ] && [ "${DIAGRAM_TEXT}" != "null" ]; then
    DIAGRAM_DECODED=$(printf '%b' "${DIAGRAM_TEXT}")

    # Layer B: Structural validation
    STRUCT_ISSUES=$(validate_diagram "${DIAGRAM_DECODED}" || true)

    # Layer C: SVG width check (only if structural validation passed)
    WIDTH_ISSUE=""
    if [ -z "${STRUCT_ISSUES}" ]; then
      RENDERED_WIDTH=$(check_diagram_width "${DIAGRAM_DECODED}" 900 || true)
      if [ -n "${RENDERED_WIDTH}" ] && [ "${RENDERED_WIDTH}" != "0" ]; then
        WIDTH_ISSUE="Diagram rendered to ${RENDERED_WIDTH}px wide (must fit 800px). "
      fi
    fi

    ALL_ISSUES="${STRUCT_ISSUES}${WIDTH_ISSUE}"
    if [ -n "${ALL_ISSUES}" ]; then
      echo "==> Diagram validation failed: ${ALL_ISSUES}"
      echo "==> Retrying diagram generation with tighter constraints..."

      RETRY_FILE="${WORK_DIR}/diagram_retry.json"
      RETRY_PROMPT_FILE="${WORK_DIR}/diagram_retry_prompt.txt"
      build_prompt_file "${RETRY_PROMPT_FILE}" \
        "str" "You previously generated a Mermaid architecture diagram that has issues:
${ALL_ISSUES}

Fix the diagram. Output ONLY a JSON object with a single key:
{\"architectureDiagram\": \"<fixed mermaid flowchart>\"}

Rules:
- flowchart TD only, max 6 nodes, max 2 subgraphs (NO nesting)
- Labels: max 2 words, no resource metrics
- Edge labels: 1-2 words max
- Keep it simple and vertical — must fit an 800px container
- Single JSON string with \\n for line breaks
- Output ONLY the JSON object, no markdown fences

Original diagram:
${DIAGRAM_TEXT}"

      if run_pass "diagram_retry" "${RETRY_PROMPT_FILE}" "${RETRY_FILE}"; then
        # Merge retry diagram back into pass 2 output
        jq -s '.[0] * .[1]' "${PASS2_FILE}" "${RETRY_FILE}" > "${PASS2_FILE}.tmp" \
          && mv "${PASS2_FILE}.tmp" "${PASS2_FILE}"
        echo "==> Diagram retry merged into pass 2 output"
      else
        echo "WARNING: Diagram retry failed — using original diagram"
      fi
    else
      echo "==> Diagram validation passed"
    fi
  fi
fi

else
  echo "==> Skipping pass 2 (core) — no relevant sections requested"
  echo '{}' > "${PASS2_FILE}"
fi

# ============================================================================
# Pass 3: FinOps + SecOps Analysis
# ============================================================================
PASS3_FILE="${WORK_DIR}/pass_3.json"
if any_section_requested "finopsAnalysis" "secopsAnalysis"; then

PASS3_PROMPT=$(cat <<'EOF'
You are writing financial and security analysis of Kubernetes experiment benchmark results.
You have an analysis plan and the full experiment data.

Output ONLY a JSON object with these sections:

{
  "finopsAnalysis": {
    "overview": "<High-level cost assessment of the experiment>",
    "costDrivers": [
      "<Primary cost driver with explanation>",
      "<Secondary cost driver>"
    ],
    "projection": "<Production projection: What would this cost running 24/7 on production-grade nodes (not preemptible)? Calculate monthly cost for a realistic multi-node setup. Include the math.>",
    "optimizations": [
      "<Specific cost optimization suggestion with expected savings>"
    ]
  },

  "secopsAnalysis": {
    "overview": "<Security posture assessment of the deployed components>",
    "findings": [
      "<Security observation about the deployment>",
      "<Network policy or RBAC observation>",
      "<Another security finding>"
    ],
    "supplyChain": "<Assessment of image provenance, signing, SBOM status for the components used>"
  }
}

Rules:
- finopsAnalysis.projection must include actual dollar amounts for 24/7 production operation
- finopsAnalysis.costDrivers should reference the actual cost estimate from the data
- secopsAnalysis should assess the components actually deployed (check the targets/components)
- Consider: network policies, RBAC, secrets management, image trust, resource limits
- Be specific and actionable — infrastructure engineers will act on these findings
- Output ONLY the JSON object, no markdown fences or extra text

ANALYSIS PLAN:
EOF
)

PASS3_PROMPT_FILE="${WORK_DIR}/pass_3_prompt.txt"
build_prompt_file "${PASS3_PROMPT_FILE}" \
  "str" "${PASS3_PROMPT}" \
  "file" "${PASS1_FILE}" \
  "str" "" \
  "str" "EXPERIMENT DATA:" \
  "file" "${SUMMARY_FILE}"
run_pass "pass_3_finops_secops" "${PASS3_PROMPT_FILE}" "${PASS3_FILE}" || true

else
  echo "==> Skipping pass 3 (finops/secops) — no relevant sections requested"
  echo '{}' > "${PASS3_FILE}"
fi

# ============================================================================
# Pass 4: Capabilities Matrix + Feedback
# ============================================================================
PASS4_FILE="${WORK_DIR}/pass_4.json"
if any_section_requested "capabilitiesMatrix" "feedback"; then

IS_COMPARISON=$(jq -r '.isComparison // false' "${PASS1_FILE}" 2>/dev/null || echo "false")

if [ "${IS_COMPARISON}" = "true" ]; then
  CAP_MATRIX_INSTRUCTION=$(cat <<'EOF'
  "capabilitiesMatrix": {
    "technologies": ["Tech1", "Tech2"],
    "categories": [
      {
        "name": "<Category name, e.g. 'Query Language', 'Storage', 'Resource Efficiency'>",
        "capabilities": [
          {
            "name": "<Capability name, e.g. 'Full-text search'>",
            "values": {"Tech1": "<Assessment>", "Tech2": "<Assessment>"}
          }
        ]
      }
    ],
    "summary": "<2-3 sentence critical verdict: which technology wins, under what conditions, and the key trade-off the reader must weigh>"
  },
EOF
)
else
  CAP_MATRIX_INSTRUCTION='"capabilitiesMatrix": null,'
fi

PASS4_PROMPT=$(cat <<EOF
You are writing the capabilities assessment and feedback for a Kubernetes experiment.
You have an analysis plan and the full experiment data.

Output ONLY a JSON object with these sections:

{
  ${CAP_MATRIX_INSTRUCTION}

  "feedback": {
    "recommendations": [
      "<Actionable next-iteration suggestion>",
      "<Another recommendation>"
    ],
    "experimentDesign": [
      "<How to improve this experiment's methodology>",
      "<Additional metrics or tests to add>"
    ]
  }
}

Rules:
- capabilitiesMatrix (if comparison): 3-5 categories with 2-4 capabilities each. Values should be concise assessments (e.g., "Limited (LogQL)", "Full Lucene syntax", "~0.1 cores avg"). summary: a direct critical verdict — which technology wins and why, with the key trade-off
- feedback.recommendations: 2-4 actionable items for the next experiment iteration
- feedback.experimentDesign: 1-3 suggestions for improving the benchmark methodology
- Output ONLY the JSON object, no markdown fences or extra text

ANALYSIS PLAN:
EOF
)

PASS4_PROMPT_FILE="${WORK_DIR}/pass_4_prompt.txt"
build_prompt_file "${PASS4_PROMPT_FILE}" \
  "str" "${PASS4_PROMPT}" \
  "file" "${PASS1_FILE}" \
  "str" "" \
  "str" "EXPERIMENT DATA:" \
  "file" "${SUMMARY_FILE}"
run_pass "pass_4_capabilities" "${PASS4_PROMPT_FILE}" "${PASS4_FILE}" || true

else
  echo "==> Skipping pass 4 (capabilities) — no relevant sections requested"
  echo '{}' > "${PASS4_FILE}"
fi

# ============================================================================
# Pass 5: Body Synthesis (rich narrative with typed content blocks)
# ============================================================================
PASS5_FILE="${WORK_DIR}/pass_5.json"
if section_requested "body"; then

# Build prior-sections context file from passes 2-4
PRIOR_SECTIONS_FILE="${WORK_DIR}/prior_sections.txt"
> "${PRIOR_SECTIONS_FILE}"
for prior_file in "${PASS2_FILE}" "${PASS3_FILE}" "${PASS4_FILE}"; do
  if [ -f "${prior_file}" ] && [ "$(cat "${prior_file}")" != "{}" ]; then
    cat "${prior_file}" >> "${PRIOR_SECTIONS_FILE}"
    printf '\n' >> "${PRIOR_SECTIONS_FILE}"
  fi
done

# Extract available metric keys from summary.json
METRIC_KEYS=""
if jq -e '.metrics.queries' "${SUMMARY_FILE}" > /dev/null 2>&1; then
  METRIC_KEYS=$(jq -r '.metrics.queries | keys | join(", ")' "${SUMMARY_FILE}")
fi

PASS5_PROMPT_STATIC=$(cat <<EOF
You are writing the main analysis for a Kubernetes experiment benchmark page.

OUTPUT FORMAT: A JSON object with a single "body" key containing a "blocks" array.
Each block has a "type" field. Available types:

  text       — Prose paragraph. Keep SHORT (2-3 sentences max). Let visuals speak.
               Fields: type, content
  topic      — Collapsible subsection. Has "title" and nested "blocks" array.
               Fields: type, title, blocks
  metric     — Inline chart. "key" must match a metric key listed below.
               "size": "large" (full width) or "small" (compact, ~50% width).
               Optional "insight" annotation below the chart.
               Fields: type, key, size (optional), insight (optional)
  comparison — Side-by-side value cards. "items" array with label + value + description.
               Fields: type, items (array of {label, value, description?})
  capabilityRow — Single capability row. "capability" name + "values" (tech → assessment).
               Fields: type, capability, values
  table      — Data table. "headers" array + "rows" (array of string arrays) + "caption".
               Fields: type, headers, rows, caption (optional)
  architecture — Mermaid flowchart diagram. Same rules as top-level architectureDiagram:
               max 8 nodes, max 3 subgraphs (no nesting), short labels (max 3 words).
               "diagram" is 'flowchart TD' syntax with \\n.
               "format" must be "mermaid". "caption" optional.
               Only include if the body needs a DIFFERENT view (e.g. data flow detail).
               Do NOT duplicate the top-level architectureDiagram.
               Fields: type, diagram, format, caption (optional)
  code       — Reference a code snippet by key. "key" must match a codeSnippets key.
               Optional "insight" provides contextual explanation shown below the code
               (only shown when no annotations are present).
               Optional "annotations" array highlights specific lines with categorized callouts.
               UI renders syntax-highlighted source code with annotated line backgrounds and badges.
               Fields: type, key, insight (optional), annotations (optional array)

               Each annotation object:
                 fromLine: number  — 1-based offset within the snippet (1 = first line of snippet)
                 toLine?: number   — end line inclusive (omit for single-line annotation)
                 category: string  — one of: syscall, algorithm, hot-path, config, branching, io, general
                 label: string     — 2-5 word header (e.g. "fsync(2) durability barrier")
                 content: string   — 1-2 sentence explanation that MUST reference specific metric data

               Annotation rules:
               - Max 3 annotations per code block — highlight the most important lines only
               - fromLine/toLine are offsets within the snippet (1 = first line), NOT file line numbers
               - Multi-line annotations span at most 5 contiguous lines
               - Annotation content must reference specific metric values from the experiment data
               - Do NOT create overlapping line ranges between annotations
               - If no specific lines are interesting enough to annotate, use "insight" instead

               Example:
               {"type":"code","key":"fsync-store","annotations":[
                 {"fromLine":5,"toLine":8,"category":"branching","label":"Conditional sync path",
                  "content":"This branch determines whether fsync is called. In fsync mode, adds 4.5ms mean latency per write."},
                 {"fromLine":7,"category":"syscall","label":"fsync(2) durability barrier",
                  "content":"file.sync_all() maps to fsync(2). On pd-standard this takes 5.2ms vs 3.1ms on pd-ssd."}
               ]}
  callout    — Emphasis box. "variant" (info|warning|success|finding) + "title" + "content".
               Fields: type, variant, title, content
  recommendation — Action item. "priority" (p0-p3) + "title" + "description" + "effort" (low|medium|high).
               Fields: type, priority, title, description, effort (optional)
  row        — Horizontal layout group. Children render side-by-side (2-3 items).
               Use to pair small metrics, or place a metric next to explanatory text.
               On mobile, items stack vertically.
               Fields: type, blocks

STRUCTURE: Create 3-5 topic blocks, each covering one theme of the experiment
(e.g., "Resource Efficiency", "Query Capabilities", "Production Readiness").
Add brief intro text before the first topic, and a closing verdict after the last.
Within topics, interleave metric blocks with code blocks where the code directly
explains the observed behavior. If code snippets are available, distribute them
across the most relevant topics rather than clustering them all in one place.

PHILOSOPHY:
- VISUAL-FIRST: Prefer showing a chart or comparison block over describing numbers in prose.
  A metric block + one sentence of context > a paragraph restating metric values.
- CHUNKED: Each topic should be self-contained. A reader can expand just one topic
  to understand that aspect of the experiment.
- MINIMAL PROSE: Text blocks should be 2-3 sentences connecting visuals, not data dumps.
  The components tell the story — text blocks provide the connective tissue.
- PAIR VISUALS: When using size: "small" metrics, wrap them in a "row" block so they
  sit side-by-side. A metric + text also works well in a row. Never put more than 3
  items in a row. Full-width blocks (tables, architecture, large metrics) stay outside rows.
- INTERLEAVE: Alternate between visual blocks and brief text. Never stack >2 text blocks.
- CODE AS EVIDENCE: When code snippets are available, use them as evidence alongside the metrics
  they influence. Place a "code" block adjacent to the metric or discussion it explains — e.g.,
  show the fsync implementation next to fsync latency metrics. Every available code snippet
  should appear at least once in the body. The "insight" field on code blocks should provide
  context specific to that topic — not just repeat the snippet's description.

AVAILABLE METRIC KEYS:
${METRIC_KEYS}

AVAILABLE CODE SNIPPET KEYS:
${CODE_SNIPPET_KEYS}

${CODE_PLACEMENT_HINT}

PRIOR ANALYSIS (use as source material — DO NOT repeat verbatim):
EOF
)

PASS5_PROMPT_FILE="${WORK_DIR}/pass_5_prompt.txt"
build_prompt_file "${PASS5_PROMPT_FILE}" \
  "str" "${PASS5_PROMPT_STATIC}" \
  "file" "${PRIOR_SECTIONS_FILE}" \
  "str" "EXPERIMENT DATA:" \
  "file" "${SUMMARY_FILE}" \
  "str" "${STUDY_CONTEXT}" \
  "str" "Output ONLY the JSON object with the \"body\" key, no markdown fences or extra text."
run_pass "pass_5_body_synthesis" "${PASS5_PROMPT_FILE}" "${PASS5_FILE}" || true

else
  echo "==> Skipping pass 5 (body synthesis) — body not requested"
  echo '{}' > "${PASS5_FILE}"
fi

# ============================================================================
# Assembly: Merge all passes into final AnalysisResult
# ============================================================================
echo "==> Assembling final analysis from 5 passes"

FINAL_FILE="${WORK_DIR}/analysis.json"

# Merge all pass outputs, then add backward-compat fields and metadata
jq -s --arg ts "$(date -u +%Y-%m-%dT%H:%M:%SZ)" --arg model "claude-opus-4-6" '
  # Start with pass 2 (core: abstract, targetAnalysis, performanceAnalysis, metricInsights, architectureDiagram)
  (.[1] // {}) *
  # Merge pass 3 (finopsAnalysis, secopsAnalysis)
  (.[2] // {}) *
  # Merge pass 4 (capabilitiesMatrix, feedback)
  (.[3] // {}) *
  # Merge pass 5 (body synthesis)
  (.[4] // {}) +
  # Add backward-compat fields + metadata
  {
    summary: ((.[1] // {}).abstract // "Analysis incomplete"),
    generatedAt: $ts,
    model: $model
  }
' "${PASS1_FILE}" "${PASS2_FILE}" "${PASS3_FILE}" "${PASS4_FILE}" "${PASS5_FILE}" > "${FINAL_FILE}"

# Remove the plan fields from the final output (they were just for inter-pass context)
jq 'del(.technologies, .isComparison, .focusAreas, .domainContext, .domain)' \
  "${FINAL_FILE}" > "${FINAL_FILE}.tmp" && mv "${FINAL_FILE}.tmp" "${FINAL_FILE}"

# Remove null capabilitiesMatrix for non-comparison experiments
jq 'if .capabilitiesMatrix == null then del(.capabilitiesMatrix) else . end' \
  "${FINAL_FILE}" > "${FINAL_FILE}.tmp" && mv "${FINAL_FILE}.tmp" "${FINAL_FILE}"

# Strip any sections that weren't explicitly requested
ALL_SECTIONS="abstract targetAnalysis performanceAnalysis metricInsights finopsAnalysis secopsAnalysis body capabilitiesMatrix feedback architectureDiagram vocabulary"
for section in ${ALL_SECTIONS}; do
  if ! section_requested "${section}"; then
    jq "del(.${section})" "${FINAL_FILE}" > "${FINAL_FILE}.tmp" && mv "${FINAL_FILE}.tmp" "${FINAL_FILE}"
  fi
done

echo "==> Final analysis assembled ($(wc -c < "${FINAL_FILE}") bytes)"
echo "==> Sections present: $(jq -r 'keys | join(", ")' "${FINAL_FILE}")"

# Merge analysis into summary
echo "==> Merging analysis into summary.json"
jq --slurpfile analysis "${FINAL_FILE}" \
  '. + {analysis: $analysis[0]}' \
  "${SUMMARY_FILE}" > "${ENRICHED_FILE}"

# Ensure architectureDiagramFormat is set when diagram contains Mermaid syntax
if jq -e '.analysis.architectureDiagram' "${ENRICHED_FILE}" > /dev/null 2>&1; then
  if ! jq -e '.analysis.architectureDiagramFormat' "${ENRICHED_FILE}" > /dev/null 2>&1; then
    DIAGRAM_CONTENT=$(jq -r '.analysis.architectureDiagram // ""' "${ENRICHED_FILE}")
    if echo "${DIAGRAM_CONTENT}" | grep -qE '^(flowchart|graph|sequenceDiagram|classDiagram)'; then
      echo "==> Auto-setting architectureDiagramFormat to 'mermaid' (detected Mermaid syntax)"
      jq '.analysis.architectureDiagramFormat = "mermaid"' "${ENRICHED_FILE}" > "${ENRICHED_FILE}.tmp" \
        && mv "${ENRICHED_FILE}.tmp" "${ENRICHED_FILE}"
    fi
  fi
fi

# ============================================================================
# Verbose logging: Upload all intermediate pass outputs to S3
# ============================================================================
echo "==> Uploading verbose analysis artifacts to S3"
export AWS_ACCESS_KEY_ID="${S3_ACCESS_KEY:-any}"
export AWS_SECRET_ACCESS_KEY="${S3_SECRET_KEY:-any}"
S3_BASE="s3://experiment-results/${EXPERIMENT_NAME}"

# Build a trace manifest with timing and sizes
TRACE_FILE="${WORK_DIR}/analysis_trace.json"
jq -n \
  --arg experiment "${EXPERIMENT_NAME}" \
  --arg ts "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --argjson pass1_size "$(wc -c < "${PASS1_FILE}")" \
  --argjson pass2_size "$(wc -c < "${PASS2_FILE}")" \
  --argjson pass3_size "$(wc -c < "${PASS3_FILE}")" \
  --argjson pass4_size "$(wc -c < "${PASS4_FILE}")" \
  --argjson pass5_size "$(wc -c < "${PASS5_FILE}")" \
  --argjson final_size "$(wc -c < "${FINAL_FILE}")" \
  '{
    experiment: $experiment,
    analyzedAt: $ts,
    passes: {
      "1_plan": {file: "pass_1_plan.json", bytes: $pass1_size},
      "2_core": {file: "pass_2_core.json", bytes: $pass2_size},
      "3_finops_secops": {file: "pass_3_finops_secops.json", bytes: $pass3_size},
      "4_capabilities": {file: "pass_4_capabilities.json", bytes: $pass4_size},
      "5_body_synthesis": {file: "pass_5_body_synthesis.json", bytes: $pass5_size}
    },
    final: {file: "analysis_final.json", bytes: $final_size}
  }' > "${TRACE_FILE}"

# Upload each pass output + raw JSONL + stderr logs
for artifact in \
  "${PASS1_FILE}:analysis/pass_1_plan.json" \
  "${PASS2_FILE}:analysis/pass_2_core.json" \
  "${PASS3_FILE}:analysis/pass_3_finops_secops.json" \
  "${PASS4_FILE}:analysis/pass_4_capabilities.json" \
  "${PASS5_FILE}:analysis/pass_5_body_synthesis.json" \
  "${FINAL_FILE}:analysis/analysis_final.json" \
  "${TRACE_FILE}:analysis/trace.json"; do
  LOCAL="${artifact%%:*}"
  REMOTE="${artifact##*:}"
  if [ -f "${LOCAL}" ]; then
    aws --endpoint-url "http://${S3_ENDPOINT}" s3 cp "${LOCAL}" "${S3_BASE}/${REMOTE}" --no-sign-request 2>/dev/null || \
    aws --endpoint-url "http://${S3_ENDPOINT}" s3 cp "${LOCAL}" "${S3_BASE}/${REMOTE}" 2>/dev/null || \
    echo "WARNING: Failed to upload ${REMOTE}"
  fi
done

# Upload raw JSONL outputs (the full claude response before extraction) and stderr
for pass_name in pass_1_plan pass_2_core pass_3_finops_secops pass_4_capabilities pass_5_body_synthesis diagram_retry; do
  for suffix in _raw.json _stderr.log; do
    LOCAL="${WORK_DIR}/${pass_name}${suffix}"
    if [ -f "${LOCAL}" ] && [ -s "${LOCAL}" ]; then
      aws --endpoint-url "http://${S3_ENDPOINT}" s3 cp "${LOCAL}" "${S3_BASE}/analysis/${pass_name}${suffix}" --no-sign-request 2>/dev/null || \
      aws --endpoint-url "http://${S3_ENDPOINT}" s3 cp "${LOCAL}" "${S3_BASE}/analysis/${pass_name}${suffix}" 2>/dev/null || true
    fi
  done
done

echo "==> Verbose artifacts uploaded to ${S3_BASE}/analysis/"

# Upload enriched summary back to S3
echo "==> Uploading enriched summary to S3"
S3_DEST="s3://experiment-results/${EXPERIMENT_NAME}/summary.json"
if ! aws --endpoint-url "http://${S3_ENDPOINT}" s3 cp "${ENRICHED_FILE}" "${S3_DEST}" --no-sign-request 2>/dev/null; then
  # Try with signing
  if ! aws --endpoint-url "http://${S3_ENDPOINT}" s3 cp "${ENRICHED_FILE}" "${S3_DEST}" 2>&1; then
    echo "WARNING: Failed to upload enriched summary to S3 — continuing to GitHub commit"
  fi
fi

# Commit to GitHub if token is available and branch was explicitly set
if [ -n "${SKIP_GITHUB_COMMIT}" ]; then
  echo "==> Skipping GitHub commit — GITHUB_BRANCH was not set"
elif [ -n "${GITHUB_TOKEN:-}" ] && [ -n "${GITHUB_REPO:-}" ]; then
  echo "==> Committing enriched results to GitHub"

  OWNER=$(echo "${GITHUB_REPO}" | cut -d/ -f1)
  REPO=$(echo "${GITHUB_REPO}" | cut -d/ -f2)
  FILE_PATH="${GITHUB_RESULTS_PATH}/${EXPERIMENT_NAME}.json"
  API_URL="https://api.github.com/repos/${OWNER}/${REPO}/contents/${FILE_PATH}"

  # Base64-encode the enriched JSON to a file (avoids ARG_MAX for large payloads)
  base64 -w0 "${ENRICHED_FILE}" > "${WORK_DIR}/content_b64.txt"

  # Check if file exists (get SHA for update)
  EXISTING_SHA=""
  EXISTING=$(curl -sf -H "Authorization: token ${GITHUB_TOKEN}" \
    -H "Accept: application/vnd.github.v3+json" \
    "${API_URL}?ref=${GITHUB_BRANCH}" 2>/dev/null || true)

  if echo "${EXISTING}" | jq -e '.sha' > /dev/null 2>&1; then
    EXISTING_SHA=$(echo "${EXISTING}" | jq -r '.sha')
  fi

  # Build the commit payload using file-based jq (--rawfile avoids ARG_MAX)
  PAYLOAD_FILE="${WORK_DIR}/payload.json"
  if [ -n "${EXISTING_SHA}" ]; then
    jq -n \
      --arg msg "data: Update ${EXPERIMENT_NAME} with AI analysis" \
      --rawfile content "${WORK_DIR}/content_b64.txt" \
      --arg branch "${GITHUB_BRANCH}" \
      --arg sha "${EXISTING_SHA}" \
      '{message: $msg, content: $content, branch: $branch, sha: $sha}' > "${PAYLOAD_FILE}"
  else
    jq -n \
      --arg msg "data: Add ${EXPERIMENT_NAME} with AI analysis" \
      --rawfile content "${WORK_DIR}/content_b64.txt" \
      --arg branch "${GITHUB_BRANCH}" \
      '{message: $msg, content: $content, branch: $branch}' > "${PAYLOAD_FILE}"
  fi

  if curl -sf -X PUT "${API_URL}" \
    -H "Authorization: token ${GITHUB_TOKEN}" \
    -H "Accept: application/vnd.github.v3+json" \
    -d @"${PAYLOAD_FILE}" > /dev/null; then
    echo "==> Results committed to GitHub: ${FILE_PATH}"
  else
    echo "WARNING: Failed to commit results to GitHub"
  fi
else
  echo "==> GITHUB_TOKEN or GITHUB_REPO not set — skipping GitHub commit"
fi

echo "==> Analysis complete for ${EXPERIMENT_NAME} (5 passes)"
