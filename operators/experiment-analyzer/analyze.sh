#!/usr/bin/env bash
set -euo pipefail

# Required environment variables:
#   EXPERIMENT_NAME  - experiment name (used for S3 key and GitHub path)
#   S3_ENDPOINT      - SeaweedFS S3 endpoint (e.g. seaweedfs-s3.seaweedfs.svc.cluster.local:8333)
#   GITHUB_TOKEN     - GitHub PAT for committing results
#   GITHUB_REPO      - GitHub repo (e.g. illMadeCoder/k8s-ai-cloud-testbed)
#   CLAUDE_CODE_OAUTH_TOKEN - Claude Code auth token (from claude setup-token)
#
# Optional:
#   GITHUB_BRANCH    - branch to commit to (default: main)
#   GITHUB_RESULTS_PATH - path in repo for result JSONs (default: site/data)

: "${EXPERIMENT_NAME:?EXPERIMENT_NAME is required}"
: "${S3_ENDPOINT:?S3_ENDPOINT is required}"
: "${CLAUDE_CODE_OAUTH_TOKEN:?CLAUDE_CODE_OAUTH_TOKEN is required}"

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

Produce a JSON object with this exact structure:
{
  "summary": "<2-4 sentence overview of the experiment findings>",
  "metricInsights": {
    "<metric_name>": "<1-2 sentence insight about what this metric shows>"
  },
  "recommendations": ["<optional improvement suggestion>"]
}

Rules:
- "summary" should describe the key findings, which target/component performed best, and any notable patterns
- "metricInsights" must have one entry per metric key in the input's metrics.queries object, using the exact same key names
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
if ! claude -p "${FULL_PROMPT}" --output-format json > "${ANALYSIS_FILE}" 2>/dev/null; then
  echo "ERROR: Claude Code analysis failed"
  exit 1
fi

echo "==> Analysis produced ($(wc -c < "${ANALYSIS_FILE}") bytes)"

# The claude --output-format json wraps the result; extract the actual content.
# The output is a JSON array with a single object containing a "text" field with the analysis.
# Try to extract the inner JSON from the claude output format.
if jq -e '.[0].text' "${ANALYSIS_FILE}" > /dev/null 2>&1; then
  # Claude --output-format json wraps in [{type:"text", text:"..."}]
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

# Upload enriched summary back to S3
echo "==> Uploading enriched summary to S3"
if ! curl -sf -X PUT -T "${ENRICHED_FILE}" "${S3_URL}"; then
  echo "WARNING: Failed to upload enriched summary to S3 — continuing to GitHub commit"
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
