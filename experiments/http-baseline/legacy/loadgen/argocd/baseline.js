// HTTP Baseline Load Test Scenario
// Simple GET/POST mix to establish baseline metrics

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const rootLatency = new Trend('root_latency');
const healthLatency = new Trend('health_latency');
const echoLatency = new Trend('echo_latency');

// Test configuration
export const options = {
  // Default: 10 users for 60s
  vus: __ENV.USERS || 10,
  duration: __ENV.DURATION || '60s',

  // Thresholds for pass/fail
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests under 500ms
    errors: ['rate<0.1'],              // Error rate under 10%
  },
};

const BASE_URL = __ENV.TARGET_URL || 'http://demo-app.demo.svc:80';

export default function () {
  // Weighted distribution: 70% root, 20% health, 10% echo
  const rand = Math.random();

  if (rand < 0.7) {
    // GET root endpoint - most common operation
    const res = http.get(`${BASE_URL}/`);
    rootLatency.add(res.timings.duration);
    const success = check(res, {
      'root status 200': (r) => r.status === 200,
    });
    errorRate.add(!success);

  } else if (rand < 0.9) {
    // GET health endpoint
    const res = http.get(`${BASE_URL}/health`);
    healthLatency.add(res.timings.duration);
    const success = check(res, {
      'health status 200': (r) => r.status === 200,
    });
    errorRate.add(!success);

  } else {
    // POST echo endpoint
    const payload = JSON.stringify({
      message: 'baseline test',
      timestamp: new Date().toISOString(),
    });
    const params = {
      headers: { 'Content-Type': 'application/json' },
    };
    const res = http.post(`${BASE_URL}/echo`, payload, params);
    echoLatency.add(res.timings.duration);
    const success = check(res, {
      'echo status 200': (r) => r.status === 200,
    });
    errorRate.add(!success);
  }

  // Wait between requests (0.5-2 seconds like original)
  sleep(0.5 + Math.random() * 1.5);
}

// Summary output
export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    '/tmp/summary.json': JSON.stringify(data),
  };
}

function textSummary(data, opts) {
  const metrics = data.metrics;
  return `
=== HTTP Baseline Test Results ===

Duration: ${data.state.testRunDurationMs}ms
VUs: ${options.vus}

Requests:
  Total: ${metrics.http_reqs?.values?.count || 0}
  Rate: ${(metrics.http_reqs?.values?.rate || 0).toFixed(2)}/s

Latency (ms):
  Avg: ${(metrics.http_req_duration?.values?.avg || 0).toFixed(2)}
  P50: ${(metrics.http_req_duration?.values?.['p(50)'] || 0).toFixed(2)}
  P95: ${(metrics.http_req_duration?.values?.['p(95)'] || 0).toFixed(2)}
  P99: ${(metrics.http_req_duration?.values?.['p(99)'] || 0).toFixed(2)}

Errors: ${(metrics.errors?.values?.rate || 0) * 100}%

Thresholds: ${Object.entries(data.thresholds || {}).every(([k,v]) => v.ok) ? 'PASSED' : 'FAILED'}
`;
}
