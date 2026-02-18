# Performance Testing Reference

## Install

```bash
brew install k6
cargo install oha
go install github.com/tsenart/vegeta@latest
go install github.com/rakyll/hey@latest
npm install -g lighthouse
cargo install hyperfine
```

## k6 — Comprehensive Load Testing

### Basic load test with SLA thresholds

```javascript
// tests/perf/basic-load.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('error_rate');
const apiLatency = new Trend('api_latency', true);

export const options = {
  stages: [
    { duration: '1m',  target: 50 },
    { duration: '2m',  target: 50 },
    { duration: '30s', target: 0  },
  ],
  thresholds: {
    http_req_duration: ['p(50)<100', 'p(95)<500', 'p(99)<1000'],
    http_req_failed:   ['rate<0.01'],
    error_rate:        ['rate<0.05'],
    api_latency:       ['p(95)<300'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'https://api.example.com';

export default function () {
  const headers = { Authorization: `Bearer ${__ENV.API_TOKEN}` };
  const choice = Math.random();

  if (choice < 0.6) {
    const res = http.get(`${BASE_URL}/products`, { headers });
    check(res, { 'products 200': (r) => r.status === 200 });
    apiLatency.add(res.timings.duration);
    errorRate.add(res.status !== 200);

  } else if (choice < 0.9) {
    const res = http.get(`${BASE_URL}/products/1`, { headers });
    check(res, { 'product 200': (r) => r.status === 200 });

  } else {
    const res = http.post(
      `${BASE_URL}/cart`,
      JSON.stringify({ productId: 1, quantity: 1 }),
      { headers: { ...headers, 'Content-Type': 'application/json' } }
    );
    check(res, { 'cart 201': (r) => r.status === 201 });
  }

  sleep(1);
}
```

### Spike test

```javascript
// tests/perf/spike-test.js
export const options = {
  stages: [
    { duration: '10s', target: 10  },
    { duration: '30s', target: 500 },
    { duration: '10s', target: 10  },
    { duration: '30s', target: 10  },
    { duration: '10s', target: 0   },
  ],
  thresholds: {
    http_req_failed:   ['rate<0.05'],
    http_req_duration: ['p(95)<2000'],
  },
};
```

### Scenario-based (multiple workflows)

```javascript
// tests/perf/scenarios.js
export const options = {
  scenarios: {
    browse_products: {
      executor: 'constant-vus',
      vus: 30,
      duration: '3m',
      exec: 'browseProducts',
    },
    checkout_flow: {
      executor: 'ramping-arrival-rate',
      preAllocatedVUs: 20,
      stages: [
        { duration: '1m', target: 10 },
        { duration: '2m', target: 10 },
      ],
      exec: 'checkout',
    },
  },
};

export function browseProducts() { /* ... */ }
export function checkout() { /* ... */ }
```

### k6 CLI Commands

```bash
k6 run tests/perf/basic-load.js
k6 run -e BASE_URL=https://staging.api.example.com \
       -e API_TOKEN=$TOKEN \
       tests/perf/basic-load.js
k6 run --vus 1 --iterations 1 tests/perf/basic-load.js
k6 run --out json=results/k6-output.json tests/perf/basic-load.js
k6 run --out csv=results/k6-output.csv tests/perf/basic-load.js
k6 run --stage 30s:10,1m:50,30s:0 tests/perf/basic-load.js

jaq '.metrics.http_req_duration.values | {p50: .["p(50)"], p95: .["p(95)"], p99: .["p(99)"]}' \
  < results/raw.json
```

## oha — Quick HTTP Benchmarks

```bash
oha -n 50 -c 10 https://api.example.com/products
oha -z 30s -c 20 https://api.example.com/products

oha -z 30s -c 10 \
  -H "Authorization: Bearer $TOKEN" \
  https://api.example.com/protected

oha -n 100 -c 10 \
  -m POST \
  -H "Content-Type: application/json" \
  -d '{"productId": 1}' \
  https://api.example.com/cart

oha --json https://api.example.com/health | jaq '{
  rps: .rps,
  p50: .latencyPercentiles."50",
  p95: .latencyPercentiles."95",
  p99: .latencyPercentiles."99",
  errors: .errorCount
}'

oha --no-keepalive -n 100 https://api.example.com/health
oha --insecure -z 30s https://staging.api.example.com/health
```

## vegeta — Constant-Rate Load Testing

```bash
echo "GET https://api.example.com/health" | \
  vegeta attack -rate=100 -duration=30s | \
  vegeta report

echo "GET https://api.example.com/products" | \
  vegeta attack \
    -rate=50 \
    -duration=60s \
    -header="Authorization: Bearer $TOKEN" | \
  vegeta report

cat > targets.txt << 'EOF'
GET https://api.example.com/products
Authorization: Bearer TOKEN_HERE

GET https://api.example.com/categories
Authorization: Bearer TOKEN_HERE

POST https://api.example.com/cart
Content-Type: application/json
Authorization: Bearer TOKEN_HERE
@/tmp/cart-body.json
EOF

vegeta attack -targets=targets.txt -rate=20 -duration=30s | \
  vegeta report

echo "GET https://api.example.com/health" | \
  vegeta attack -rate=50 -duration=30s | \
  tee results/vegeta.bin | \
  vegeta plot > results/latency-plot.html

echo "GET https://api.example.com/health" | \
  vegeta attack -rate=50 -duration=30s | \
  vegeta report -type=json | \
  jaq '{
    p50: .latencies.mean,
    p95: .latencies."95th",
    p99: .latencies."99th",
    success_rate: .success,
    rps: .throughput
  }'
```

## Lighthouse CLI — Web Vitals

```bash
lighthouse https://example.com \
  --output=html \
  --output-path=results/lighthouse.html \
  --chrome-flags="--headless"

lighthouse https://example.com \
  --only-categories=performance \
  --output=json \
  --chrome-flags="--headless" \
  --quiet | jaq '{
    score: .categories.performance.score,
    lcp: .audits["largest-contentful-paint"].displayValue,
    cls: .audits["cumulative-layout-shift"].displayValue,
    inp: .audits["experimental-interaction-to-next-paint"].displayValue,
    ttfb: .audits["server-response-time"].displayValue
  }'

check_web_vitals() {
  local url=$1
  local min_score=${2:-0.9}

  local score=$(lighthouse "$url" \
    --only-categories=performance \
    --output=json \
    --chrome-flags="--headless --no-sandbox" \
    --quiet | jaq '.categories.performance.score')

  echo "Performance score: $score (min: $min_score)"
  if (( $(echo "$score < $min_score" | bc -l) )); then
    echo "FAIL: Score below threshold"
    return 1
  fi
  echo "PASS"
}

check_web_vitals https://app.example.com 0.90
```

## Flutter Performance Profiling

```bash
flutter run --profile
flutter run --profile --dart-define=FLUTTER_PERFORMANCE_OVERLAY=true

flutter drive \
  --driver=test_driver/perf_driver.dart \
  --target=test_driver/perf_app.dart \
  --profile

flutter pub run flutter_driver:analyze-timeline timeline.json
flutter test --reporter expanded test/perf/ 2>&1 | grep "frame"
```

## SLA Validation Script

```bash
#!/usr/bin/env bash
set -euo pipefail
FAILURES=0

check_sla() {
  local name=$1 metric=$2 value=$3 threshold=$4
  if (( $(echo "$value > $threshold" | bc -l) )); then
    echo "FAIL [$name]: $metric = ${value}ms > ${threshold}ms"
    FAILURES=$((FAILURES + 1))
  else
    echo "PASS [$name]: $metric = ${value}ms"
  fi
}

RESULT=$(oha --json -n 100 -c 10 \
  -H "Authorization: Bearer $TOKEN" \
  https://api.example.com/products)

P50=$(echo "$RESULT" | jaq '.latencyPercentiles."50" * 1000 | round')
P95=$(echo "$RESULT" | jaq '.latencyPercentiles."95" * 1000 | round')
P99=$(echo "$RESULT" | jaq '.latencyPercentiles."99" * 1000 | round')

check_sla "products-api" "p50" "$P50" 100
check_sla "products-api" "p95" "$P95" 500
check_sla "products-api" "p99" "$P99" 1000

echo "---"
echo "SLA Check: $FAILURES failures"
exit $FAILURES
```
