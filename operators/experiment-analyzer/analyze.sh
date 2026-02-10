#!/usr/bin/env bash
set -euo pipefail

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
ANALYSIS_FILE="${WORK_DIR}/analysis.json"
ENRICHED_FILE="${WORK_DIR}/enriched.json"

cleanup() { rm -rf "${WORK_DIR}"; }
trap cleanup EXIT

echo "==> Fetching summary.json from S3: ${S3_URL}"
if ! curl -sf "${S3_URL}" -o "${SUMMARY_FILE}"; then
  echo "ERROR: Failed to fetch summary.json from S3"
  exit 1
fi

echo "==> Summary fetched ($(wc -c < "${SUMMARY_FILE}") bytes)"

# Build the analysis prompt
PROMPT=$(cat <<'PROMPT_EOF'
You are analyzing Kubernetes experiment benchmark results. The experiment data is a JSON object with metrics, targets, workflow info, and cost estimates.

Your analysis will be published on the Testbed Benchmarks site (Astro + Tailwind, GitHub Pages).
The site organizes experiments by domain (observability, networking, storage, cicd) and type
(comparison, tutorial, demo, baseline) — derived from experiment tags. Your analysis appears on:
- The experiment detail page (/experiments/{slug}/) as an "AI Analysis" box (summary) with
  per-metric insights below each chart and a recommendations list at the bottom.
- Comparison experiments are featured on the landing page and /comparisons/ index.

Produce a JSON object with this exact structure:
{
  "summary": "<2-4 sentence overview of the experiment findings>",
  "metricInsights": {
    "<metric_name>": "<1-2 sentence insight about what this metric shows>"
  },
  "recommendations": ["<optional improvement suggestion>"]
}

Rules:
- "summary" should describe the key findings, which target/component performed best, and any notable patterns. For comparisons, clearly state the winner and by how much.
- "metricInsights" must have one entry per metric key in the input's metrics.queries object, using the exact same key names. Each insight appears directly below its Vega-Lite chart on the site.
- "recommendations" should contain 1-3 actionable suggestions if there's clear room for improvement; omit or leave empty if results are already good
- Be specific with numbers — reference actual values from the data
- Keep language concise and technical
- Output ONLY the JSON object, no markdown fences or extra text

Here is the experiment data:
PROMPT_EOF
)

FULL_PROMPT="${PROMPT}
$(cat "${SUMMARY_FILE}")"

echo "==> Running Claude Code analysis..."
STDERR_FILE="${WORK_DIR}/claude_stderr.log"
if ! claude -p "${FULL_PROMPT}" --output-format json > "${ANALYSIS_FILE}" 2>"${STDERR_FILE}"; then
  echo "ERROR: Claude Code analysis failed"
  echo "==> stderr:" && cat "${STDERR_FILE}" || true
  echo "==> stdout:" && cat "${ANALYSIS_FILE}" || true
  exit 1
fi

echo "==> Analysis produced ($(wc -c < "${ANALYSIS_FILE}") bytes)"

# Claude --output-format json outputs newline-delimited JSON (JSONL).
# The final object has type:"result" with a .result string containing the actual text.
# Extract the result text from the JSONL stream.
if RESULT_TEXT=$(jq -rs '[.[] | select(.type == "result")] | last | .result' "${ANALYSIS_FILE}" 2>/dev/null) && [ -n "${RESULT_TEXT}" ] && [ "${RESULT_TEXT}" != "null" ]; then
  echo "${RESULT_TEXT}" > "${ANALYSIS_FILE}"
elif jq -e '.[0].text' "${ANALYSIS_FILE}" > /dev/null 2>&1; then
  # Fallback: array format [{type:"text", text:"..."}]
  INNER_TEXT=$(jq -r '.[0].text' "${ANALYSIS_FILE}")
  echo "${INNER_TEXT}" > "${ANALYSIS_FILE}"
fi

# Strip markdown fences if present
if head -1 "${ANALYSIS_FILE}" | grep -q '```'; then
  sed -i '1d' "${ANALYSIS_FILE}"
  # Remove trailing fence
  if tail -1 "${ANALYSIS_FILE}" | grep -q '```'; then
    sed -i '$d' "${ANALYSIS_FILE}"
  fi
fi

# Validate it's valid JSON
if ! jq empty "${ANALYSIS_FILE}" 2>/dev/null; then
  echo "ERROR: Analysis output is not valid JSON"
  cat "${ANALYSIS_FILE}"
  exit 1
fi

# Add generatedAt and model fields
ANALYSIS_WITH_META=$(jq \
  --arg ts "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --arg model "claude-opus-4-6" \
  '. + {generatedAt: $ts, model: $model}' \
  "${ANALYSIS_FILE}")

echo "${ANALYSIS_WITH_META}" > "${ANALYSIS_FILE}"

# Merge analysis into summary
echo "==> Merging analysis into summary.json"
jq --slurpfile analysis "${ANALYSIS_FILE}" \
  '. + {analysis: $analysis[0]}' \
  "${SUMMARY_FILE}" > "${ENRICHED_FILE}"

# Upload enriched summary back to S3 using AWS CLI (supports S3v4 signing)
echo "==> Uploading enriched summary to S3"
export AWS_ACCESS_KEY_ID="${S3_ACCESS_KEY:-any}"
export AWS_SECRET_ACCESS_KEY="${S3_SECRET_KEY:-any}"
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

echo "==> Analysis complete for ${EXPERIMENT_NAME}"
