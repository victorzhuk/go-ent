---
description: Generate code from OpenAPI/Proto specs
allowed-tools: Read, Bash, Edit
---

# Generate Code from Specs

Input: `$ARGUMENTS` (optional: `api`, `proto`, or `all`)

## Generate OpenAPI (ogen)

```bash
go run github.com/ogen-go/ogen/cmd/ogen@v1.8.1 \
    --target gen/api/v1 \
    --package apiv1 \
    --clean \
    api/openapi/v1/openapi.yaml
```

## Generate Proto (buf)

```bash
buf generate
```

## Makefile

```makefile
.PHONY: gen
gen: gen-api gen-proto

.PHONY: gen-api
gen-api:
	go run github.com/ogen-go/ogen/cmd/ogen@v1.8.1 \
		--target gen/api/v1 --package apiv1 --clean \
		api/openapi/v1/openapi.yaml

.PHONY: gen-proto
gen-proto:
	buf generate
```

## After Generation

1. Implement handler interface
2. Run `go build ./...` to verify
