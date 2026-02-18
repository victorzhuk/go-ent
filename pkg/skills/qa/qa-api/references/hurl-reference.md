# Hurl Reference

## Install

```bash
cargo install hurl xh jaq
brew install k6
brew install grpcurl
```

## hurl Syntax

### Basic request + assertions

```hurl
GET https://api.example.com/health
HTTP 200
[Asserts]
jsonpath "$.status" == "ok"
jsonpath "$.version" matches /\d+\.\d+\.\d+/
duration < 500
```

### Chained requests with captures

```hurl
POST https://api.example.com/auth/login
Content-Type: application/json
{
  "email": "{{email}}",
  "password": "{{password}}"
}
HTTP 200
[Captures]
access_token: jsonpath "$.access_token"
user_id: jsonpath "$.user.id"

GET https://api.example.com/users/{{user_id}}
Authorization: Bearer {{access_token}}
HTTP 200
[Asserts]
jsonpath "$.email" == "{{email}}"
jsonpath "$.role" == "user"

PATCH https://api.example.com/users/{{user_id}}
Authorization: Bearer {{access_token}}
Content-Type: application/json
{ "name": "Updated Name" }
HTTP 200
[Asserts]
jsonpath "$.name" == "Updated Name"
```

### GraphQL

```hurl
POST https://api.example.com/graphql
Content-Type: application/json
Authorization: Bearer {{token}}
```graphql
query GetUser($id: ID!) {
  user(id: $id) {
    id
    email
    posts { title createdAt }
  }
}
```
variables {
  "id": "{{user_id}}"
}
HTTP 200
[Asserts]
jsonpath "$.data.user.email" != null
jsonpath "$.data.user.posts" count > 0
jsonpath "$.errors" not exists
```

### CRUD suite

```hurl
POST https://api.example.com/products
Authorization: Bearer {{token}}
Content-Type: application/json
{
  "name": "Test Product",
  "price": 9.99,
  "stock": 100
}
HTTP 201
[Captures]
product_id: jsonpath "$.id"
[Asserts]
jsonpath "$.name" == "Test Product"
header "Location" exists

GET https://api.example.com/products/{{product_id}}
Authorization: Bearer {{token}}
HTTP 200
[Asserts]
jsonpath "$.price" == 9.99

PUT https://api.example.com/products/{{product_id}}
Authorization: Bearer {{token}}
Content-Type: application/json
{ "price": 14.99 }
HTTP 200
[Asserts]
jsonpath "$.price" == 14.99

DELETE https://api.example.com/products/{{product_id}}
Authorization: Bearer {{token}}
HTTP 204

GET https://api.example.com/products/{{product_id}}
Authorization: Bearer {{token}}
HTTP 404
```

### Error case testing

```hurl
GET https://api.example.com/protected
HTTP 401
[Asserts]
jsonpath "$.error" == "unauthorized"

GET https://api.example.com/users/nonexistent-id-99999
Authorization: Bearer {{token}}
HTTP 404

POST https://api.example.com/users
Content-Type: application/json
{ "email": "not-an-email" }
HTTP 422
[Asserts]
jsonpath "$.errors[0].field" == "email"

GET https://api.example.com/api/data
HTTP *
[Asserts]
status >= 200
header "X-RateLimit-Remaining" exists
```

## hurl CLI Commands

```bash
hurl auth-flow.hurl
hurl --test auth-flow.hurl
hurl --variable email=qa@test.com --variable password=secret auth-flow.hurl
hurl --variables-file .env.test auth-flow.hurl
hurl --test --glob "tests/api/**/*.hurl"
hurl --parallel --test tests/api/
hurl --test --report-json results/api-report.json tests/
hurl --test --report-html results/api-report/ tests/
hurl --test --report-tap results/api.tap tests/
hurl --cookie-jar cookies.txt auth-flow.hurl
hurl --very-verbose --test auth-flow.hurl 2>&1 | head -100
hurl --continue-on-error --test tests/
hurl --insecure tests/
```

## Variables File (`.env.test`)

```ini
base_url=https://staging.api.example.com
email=qa@test.com
password=TestPass123!
admin_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

```bash
hurl --variables-file .env.staging --test tests/
hurl --variables-file .env.prod --test tests/smoke/
```

## xh — Quick API Calls

```bash
xh GET https://api.example.com/users
xh POST https://api.example.com/users email=test@test.com name="John"
xh GET https://api.example.com/me "Authorization: Bearer $TOKEN"
xh GET https://api.example.com/users page==1 limit==20
xh GET https://api.example.com/users | jaq '.users[].email'
xh --follow --verbose GET https://api.example.com/redirect
```

## k6 — Load & Performance

```javascript
// load-test.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 20 },
    { duration: '1m',  target: 20 },
    { duration: '10s', target: 0  },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed:   ['rate<0.01'],
  },
};

export default function () {
  const res = http.get('https://api.example.com/products', {
    headers: { Authorization: `Bearer ${__ENV.API_TOKEN}` },
  });
  check(res, {
    'status 200':       (r) => r.status === 200,
    'response time OK': (r) => r.timings.duration < 500,
  });
  sleep(1);
}
```

```bash
k6 run load-test.js
k6 run -e API_TOKEN=$TOKEN load-test.js
k6 run --vus 1 --iterations 1 load-test.js
k6 run --out json=results/k6-output.json load-test.js
```

## grpcurl — gRPC Testing

```bash
grpcurl -plaintext localhost:9090 list
grpcurl -plaintext localhost:9090 describe UserService
grpcurl -plaintext -d '{"id": "123"}' localhost:9090 UserService/GetUser
grpcurl -plaintext \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"email": "test@test.com"}' \
  localhost:9090 UserService/GetUserByEmail
grpcurl -plaintext localhost:9090 list grpc.health.v1.Health
grpcurl -plaintext -d '{}' localhost:9090 grpc.health.v1.Health/Check
```

## Agent Patterns

### API discovery

```bash
xh GET https://api.example.com/openapi.json | jaq '.paths | keys[]'
xh GET https://api.example.com/docs | jaq '.components.schemas | keys'
hurl --test tests/smoke/health-checks.hurl
```

### Contract testing flow

```bash
hurl --test \
  --variables-file .env.test \
  --report-json results/contract-report.json \
  --parallel \
  tests/api/contracts/

hurl --test --variable user_id=123 tests/api/users/get-user.hurl
```

### Regression suite

```bash
hurl --test \
  --continue-on-error \
  --variables-file .env.staging \
  --report-html results/regression/ \
  --glob "tests/**/*.hurl"
```
