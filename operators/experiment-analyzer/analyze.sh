#!/usr/bin/env bash
set -euo pipefail

# Multi-pass AI experiment analyzer
#
# Produces research-paper quality analysis via 5 focused claude -p calls:
#   Pass 1: Analysis plan (technologies, focus areas, domain context)
#   Pass 2: Core analysis (abstract, targetAnalysis, performanceAnalysis, metricInsights)
#   Pass 3: FinOps + SecOps analysis
#   Pass 4: Deep dive + capabilities matrix + feedback
#   Pass 5: ASCII architecture diagram
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
#   GITHUB_BRANCH    - branch to commit to (default: main)
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

GITHUB_BRANCH="${GITHUB_BRANCH:-main}"
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
if jq -e '.analysisConfig.sections' "${SUMMARY_FILE}" > /dev/null 2>&1; then
  REQUESTED_SECTIONS=$(jq -r '.analysisConfig.sections | join(",")' "${SUMMARY_FILE}")
  echo "==> Requested analysis sections: ${REQUESTED_SECTIONS}"
else
  echo "==> No analysisConfig.sections in summary.json — skipping analysis"
  echo "==> To enable analysis, add spec.analysis.sections to the Experiment CR"
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

  # Validate JSON
  if ! jq empty "${out_file}" 2>/dev/null; then
    echo "ERROR: Output is not valid JSON"
    cat "${out_file}" >&2
    return 1
  fi
}

# --- Helper: run a single analysis pass with retry ---
run_pass() {
  local pass_name="$1"
  local prompt="$2"
  local out_file="$3"
  local raw_file="${WORK_DIR}/${pass_name}_raw.json"
  local stderr_file="${WORK_DIR}/${pass_name}_stderr.log"

  echo "==> Running ${pass_name}..."

  local attempt
  for attempt in 1 2; do
    if claude -p "${prompt}" --output-format json > "${raw_file}" 2>"${stderr_file}"; then
      if extract_json "${raw_file}" "${out_file}"; then
        echo "==> ${pass_name} complete ($(wc -c < "${out_file}") bytes)"
        return 0
      fi
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

SUMMARY_DATA=$(cat "${SUMMARY_FILE}")

# Extract study context if present (hypothesis, questions, focus from experiment spec)
STUDY_CONTEXT=""
if jq -e '.study' "${SUMMARY_FILE}" > /dev/null 2>&1; then
  STUDY_HYPOTHESIS=$(jq -r '.study.hypothesis // empty' "${SUMMARY_FILE}")
  STUDY_QUESTIONS=$(jq -r '.study.questions // [] | join("; ")' "${SUMMARY_FILE}")
  STUDY_FOCUS=$(jq -r '.study.focus // [] | join(", ")' "${SUMMARY_FILE}")

  STUDY_CONTEXT="
STUDY CONTEXT (from the experimenter — use this to guide your analysis):
"
  [ -n "${STUDY_HYPOTHESIS}" ] && STUDY_CONTEXT="${STUDY_CONTEXT}Hypothesis: ${STUDY_HYPOTHESIS}
"
  [ -n "${STUDY_QUESTIONS}" ] && STUDY_CONTEXT="${STUDY_CONTEXT}Questions to answer: ${STUDY_QUESTIONS}
"
  [ -n "${STUDY_FOCUS}" ] && STUDY_CONTEXT="${STUDY_CONTEXT}Focus areas: ${STUDY_FOCUS}
"
  echo "==> Study context found: hypothesis=$(echo "${STUDY_HYPOTHESIS}" | head -c 80)..."
else
  echo "==> No study context in experiment spec — analyzer will infer intent"
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
run_pass "pass_1_plan" "${PASS1_PROMPT}
${SUMMARY_DATA}
${STUDY_CONTEXT}" "${PASS1_FILE}" || true

PLAN_DATA=$(cat "${PASS1_FILE}")
echo "==> Analysis plan: $(jq -c '{technologies, isComparison, focusAreas}' "${PASS1_FILE}" 2>/dev/null || echo '{}')"

# ============================================================================
# Pass 2: Core Analysis (abstract, targetAnalysis, performanceAnalysis, metricInsights)
# ============================================================================
PASS2_FILE="${WORK_DIR}/pass_2.json"
if any_section_requested "abstract" "targetAnalysis" "performanceAnalysis" "metricInsights"; then

PASS2_PROMPT=$(cat <<'EOF'
You are writing research-paper quality analysis of Kubernetes experiment benchmark results.
You have an analysis plan and the full experiment data. Generate the core analysis sections.

Your analysis will be published on the Testbed Benchmarks site. Each section appears as a
styled card on the experiment detail page. Be specific with numbers from the data.

Output ONLY a JSON object with these sections:

{
  "abstract": "<4-6 sentence abstract. Start by stating whether the experiment conclusively supports, partially supports, or fails to support the hypothesis, and WHY. Summarize the key evidence. If the experiment was insufficient to evaluate the hypothesis (e.g. missing metrics, wrong granularity, too short), say so explicitly and what would be needed. End with the most actionable finding.>",

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
  }
}

Rules:
- "abstract" is the most important section — it appears directly below the hypothesis on the experiment page
- The abstract MUST open with a verdict on the hypothesis: "supported", "partially supported", "not supported", or "insufficient data to evaluate"
- Explain the causal reasoning: does the data confirm WHY the hypothesis predicted this outcome?
- If the experiment design was insufficient (wrong metrics, missing isolation, too short), state what specifically was missing
- "targetAnalysis.perTarget" must have one entry per target in the experiment
- "performanceAnalysis.findings" should have 3-6 numbered findings with actual data
- If study questions exist, findings should directly answer as many as possible
- "metricInsights" must have one entry per metric key in metrics.queries, using exact key names
- Reference specific numbers from the data (CPU cores, memory bytes, durations)
- Be technical and concise — this is for infrastructure engineers
- Output ONLY the JSON object, no markdown fences or extra text

ANALYSIS PLAN:
EOF
)

run_pass "pass_2_core" "${PASS2_PROMPT}
${PLAN_DATA}
${STUDY_CONTEXT}
EXPERIMENT DATA:
${SUMMARY_DATA}" "${PASS2_FILE}" || true

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

run_pass "pass_3_finops_secops" "${PASS3_PROMPT}
${PLAN_DATA}

EXPERIMENT DATA:
${SUMMARY_DATA}" "${PASS3_FILE}" || true

else
  echo "==> Skipping pass 3 (finops/secops) — no relevant sections requested"
  echo '{}' > "${PASS3_FILE}"
fi

# ============================================================================
# Pass 4: Deep Dive + Capabilities Matrix + Feedback
# ============================================================================
PASS4_FILE="${WORK_DIR}/pass_4.json"
if any_section_requested "body" "capabilitiesMatrix" "feedback"; then

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
You are writing the research-paper body and capabilities assessment for a Kubernetes experiment.
You have an analysis plan and the full experiment data.

Output ONLY a JSON object with these sections:

{
  ${CAP_MATRIX_INSTRUCTION}

  "body": {
    "methodology": "<How the experiment was structured: what was deployed, how it was measured, what the workflow did, and any limitations of the methodology>",
    "results": "<Key findings with specific data points. Reference actual metric values, timings, and resource usage. Structure as a narrative connecting the data points.>",
    "discussion": "<Interpretation of results: what they mean for real-world usage, how they compare to industry expectations, limitations of this benchmark, and caveats.>"
  },

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
- body.methodology should describe the actual experiment structure from the data
- body.results must reference specific metric values and timings
- body.discussion should connect findings to real-world implications
- capabilitiesMatrix (if comparison): 3-5 categories with 2-4 capabilities each. Values should be concise assessments (e.g., "Limited (LogQL)", "Full Lucene syntax", "~0.1 cores avg"). summary: a direct critical verdict — which technology wins and why, with the key trade-off
- feedback.recommendations: 2-4 actionable items for the next experiment iteration
- feedback.experimentDesign: 1-3 suggestions for improving the benchmark methodology
- Output ONLY the JSON object, no markdown fences or extra text

ANALYSIS PLAN:
EOF
)

run_pass "pass_4_deep_dive" "${PASS4_PROMPT}
${PLAN_DATA}

EXPERIMENT DATA:
${SUMMARY_DATA}" "${PASS4_FILE}" || true

else
  echo "==> Skipping pass 4 (deep dive) — no relevant sections requested"
  echo '{}' > "${PASS4_FILE}"
fi

# ============================================================================
# Pass 5: ASCII Architecture Diagram
# ============================================================================
PASS5_FILE="${WORK_DIR}/pass_5.json"
if section_requested "architectureDiagram"; then

PASS5_PROMPT=$(cat <<'EOF'
You are creating an ASCII architecture diagram for a Kubernetes experiment benchmark page.
The page uses monospace font (JetBrains Mono). Generate a clear topology diagram showing
the experiment's cluster layout, components, and data flow.

Output ONLY a JSON object:
{"architectureDiagram": "<diagram string with \n for newlines>"}

Diagram rules:
- MAX 66 characters wide (must fit 600px content area at 0.75rem monospace)
- Target 15-22 lines tall
- Use box-drawing characters: ┌ ─ ┐ │ └ ┘ for inner boxes, ═ ╔ ╗ ╚ ╝ ║ for outer boundary
- Show: hub cluster → provisioning arrow → target cluster(s)
- Inside each target cluster box, group components by role
- Show data flow with arrows: ──▶ (right), ◀── (left), │▼ (down), │▲ (up)
- Show metrics backhaul path if applicable
- Group components by role (e.g. "Loki + Promtail" not separate boxes for each)
- Omit infrastructure resources (ConfigMaps, Secrets, RBAC, ServiceAccounts)
- Label arrows with what flows through them (e.g. "metrics", "logs", "queries")

Example format for a comparison experiment:
╔══════════════════════════════════════════════════════════╗
║  Hub Cluster (Talos)                                    ║
║  ┌─────────────┐  ┌──────────┐  ┌───────────────────┐  ║
║  │ ArgoCD      │  │Crossplane│  │ Argo Workflows    │  ║
║  └──────┬──────┘  └────┬─────┘  └─────────┬─────────┘  ║
╚═════════│══════════════│═══════════════════│═════════════╝
          │ sync         │ provision         │ orchestrate
    ┌─────▼──────┐ ┌─────▼──────┐           │
    │ Target: A  │ │ Target: B  │◀──────────┘
    │            │ │            │
    │ Component1 │ │ Component2 │
    │ Component3 │ │ Component4 │
    │     │      │ │     │      │
    └─────│──────┘ └─────│──────┘
          │  metrics     │  metrics
          └──────┬───────┘
                 ▼
          ┌────────────┐
          │VictoriaM.  │
          └────────────┘

Rules:
- Adapt the layout to the actual experiment topology (number of targets, components)
- For single-target experiments, show one target box with all components
- Keep it clean and readable — whitespace is better than clutter
- Output ONLY the JSON object, no markdown fences or extra text
- The diagram value must be a single JSON string with \n for line breaks

ANALYSIS PLAN:
EOF
)

run_pass "pass_5_diagram" "${PASS5_PROMPT}
${PLAN_DATA}

EXPERIMENT DATA:
${SUMMARY_DATA}" "${PASS5_FILE}" || true

else
  echo "==> Skipping pass 5 (diagram) — architectureDiagram not requested"
  echo '{}' > "${PASS5_FILE}"
fi

# ============================================================================
# Assembly: Merge all passes into final AnalysisResult
# ============================================================================
echo "==> Assembling final analysis from 5 passes"

FINAL_FILE="${WORK_DIR}/analysis.json"

# Merge all pass outputs, then add backward-compat fields and metadata
jq -s --arg ts "$(date -u +%Y-%m-%dT%H:%M:%SZ)" --arg model "claude-opus-4-6" '
  # Start with pass 2 (core: abstract, targetAnalysis, performanceAnalysis, metricInsights)
  (.[1] // {}) *
  # Merge pass 3 (finopsAnalysis, secopsAnalysis)
  (.[2] // {}) *
  # Merge pass 4 (capabilitiesMatrix, body, feedback)
  (.[3] // {}) *
  # Merge pass 5 (architectureDiagram)
  (.[4] // {}) +
  # Add backward-compat fields + metadata
  {
    summary: ((.[1] // {}).abstract // "Analysis incomplete"),
    recommendations: (((.[3] // {}).feedback // {}).recommendations // []),
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
ALL_SECTIONS="abstract targetAnalysis performanceAnalysis metricInsights finopsAnalysis secopsAnalysis body capabilitiesMatrix feedback architectureDiagram"
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
      "4_deep_dive": {file: "pass_4_deep_dive.json", bytes: $pass4_size},
      "5_diagram": {file: "pass_5_diagram.json", bytes: $pass5_size}
    },
    final: {file: "analysis_final.json", bytes: $final_size}
  }' > "${TRACE_FILE}"

# Upload each pass output + raw JSONL + stderr logs
for artifact in \
  "${PASS1_FILE}:analysis/pass_1_plan.json" \
  "${PASS2_FILE}:analysis/pass_2_core.json" \
  "${PASS3_FILE}:analysis/pass_3_finops_secops.json" \
  "${PASS4_FILE}:analysis/pass_4_deep_dive.json" \
  "${PASS5_FILE}:analysis/pass_5_diagram.json" \
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
for pass_name in pass_1_plan pass_2_core pass_3_finops_secops pass_4_deep_dive pass_5_diagram; do
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

# Commit to GitHub if token is available
if [ -n "${GITHUB_TOKEN:-}" ] && [ -n "${GITHUB_REPO:-}" ]; then
  echo "==> Committing enriched results to GitHub"

  OWNER=$(echo "${GITHUB_REPO}" | cut -d/ -f1)
  REPO=$(echo "${GITHUB_REPO}" | cut -d/ -f2)
  FILE_PATH="${GITHUB_RESULTS_PATH}/${EXPERIMENT_NAME}.json"
  API_URL="https://api.github.com/repos/${OWNER}/${REPO}/contents/${FILE_PATH}"

  # Base64-encode the enriched JSON
  CONTENT_B64=$(base64 -w0 "${ENRICHED_FILE}")

  # Check if file exists (get SHA for update)
  EXISTING_SHA=""
  EXISTING=$(curl -sf -H "Authorization: token ${GITHUB_TOKEN}" \
    -H "Accept: application/vnd.github.v3+json" \
    "${API_URL}?ref=${GITHUB_BRANCH}" 2>/dev/null || true)

  if echo "${EXISTING}" | jq -e '.sha' > /dev/null 2>&1; then
    EXISTING_SHA=$(echo "${EXISTING}" | jq -r '.sha')
  fi

  # Build the commit payload
  if [ -n "${EXISTING_SHA}" ]; then
    PAYLOAD=$(jq -n \
      --arg msg "data: Update ${EXPERIMENT_NAME} with AI analysis" \
      --arg content "${CONTENT_B64}" \
      --arg branch "${GITHUB_BRANCH}" \
      --arg sha "${EXISTING_SHA}" \
      '{message: $msg, content: $content, branch: $branch, sha: $sha}')
  else
    PAYLOAD=$(jq -n \
      --arg msg "data: Add ${EXPERIMENT_NAME} with AI analysis" \
      --arg content "${CONTENT_B64}" \
      --arg branch "${GITHUB_BRANCH}" \
      '{message: $msg, content: $content, branch: $branch}')
  fi

  if curl -sf -X PUT "${API_URL}" \
    -H "Authorization: token ${GITHUB_TOKEN}" \
    -H "Accept: application/vnd.github.v3+json" \
    -d "${PAYLOAD}" > /dev/null; then
    echo "==> Results committed to GitHub: ${FILE_PATH}"
  else
    echo "WARNING: Failed to commit results to GitHub"
  fi
else
  echo "==> GITHUB_TOKEN or GITHUB_REPO not set — skipping GitHub commit"
fi

echo "==> Analysis complete for ${EXPERIMENT_NAME} (5 passes)"
