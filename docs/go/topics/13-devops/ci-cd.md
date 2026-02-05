# CI/CD

GitHub Actions patterns for Go projects.

## Quick Reference

| Task | Action | Notes |
|------|--------|-------|
| Go setup | `actions/setup-go@v5` | Auto-caches modules |
| Lint | `golangci/golangci-lint-action@v4` | Configurable via `.golangci.yml` |
| Test | `go test -race -cover` | Use `-race` in CI always |
| Module cache | `actions/cache@v4` | Key: `go-mod-${{ hashFiles('**/go.sum') }}` |
| Matrix testing | `strategy: matrix: go: [1.22, 1.23]` | Test across versions/OS |
| Release | `goreleaser/goreleaser-action@v5` | Multi-platform builds |
| Security scan | `securego/gosec@master` | SAST for Go |
| Dependency scan | `aquasecurity/trivy-action@master` | CVE detection |
| Vulnerability check | `golang.org/x/vuln/cmd/govulncheck` | Official Go tool |

```.yaml
# .github/workflows/test.yml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true
      - run: go test -v -race ./...
```

## Full Workflow

```.yaml
name: CI
on: [push, pull_request]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - uses: golangci/golangci-lint-action@v4

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - run: go test -v -race -coverprofile=coverage.txt ./...
      - uses: codecov/codecov-action@v3

  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - run: go build -v ./cmd/...
```

## Module Caching

```.yaml
# Manual cache (actions/setup-go@v5 auto-caches by default)
- uses: actions/cache@v4
  with:
    path: |
      ~/.cache/go-build
      ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-

- run: go mod download
- run: go mod verify  # Checksum verification
```

**Speedup metrics:** 2-5x faster builds (50s → 15s typical).

**Best practices:**
- Use `cache: true` in `actions/setup-go@v5` (automatic)
- Cache key MUST include `go.sum` hash
- Run `go mod verify` after restore
- Separate build cache from module cache for fine control

## Matrix Testing

Test across Go versions, OS, and architectures:

```.yaml
jobs:
  test:
    strategy:
      matrix:
        go: ['1.22', '1.23']
        os: [ubuntu-latest, macos-latest, windows-latest]
        arch: [amd64, arm64]
        exclude:
          - os: windows-latest
            arch: arm64
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go }}
          cache: true
      - run: go test -race ./...
        env:
          GOARCH: ${{ matrix.arch }}
```

**Patterns:**
- Test minimum supported + latest Go version
- Include Linux/macOS/Windows if multi-platform
- Use `exclude` to skip invalid combinations
- Set `fail-fast: false` to see all failures

## Goreleaser

Multi-platform builds, Docker images, GitHub releases:

```.yaml
# .github/workflows/release.yml
name: Release
on:
  push:
    tags: ['v*']

jobs:
  release:
    runs-on: ubuntu-latest
    permissions:
      contents: write
      packages: write
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**`.goreleaser.yml` example:**
```.yaml
builds:
  - env: [CGO_ENABLED=0]
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ldflags: -s -w -X main.version={{.Version}}

dockers:
  - image_templates:
      - ghcr.io/owner/repo:{{ .Tag }}
      - ghcr.io/owner/repo:latest
    dockerfile: Dockerfile

archives:
  - format: tar.gz
    name_template: '{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}'

brews:
  - repository:
      owner: your-org
      name: homebrew-tap
    folder: Formula
    homepage: https://github.com/your-org/repo
    description: Your app description
```

**Features:**
- Cross-compilation for 10+ platforms
- Docker images (multi-arch)
- GitHub releases with checksums
- Homebrew tap auto-update
- Changelog generation

## Security Scanning

```.yaml
jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true

      # SAST - Static Application Security Testing
      - name: Run gosec
        uses: securego/gosec@master
        with:
          args: '-no-fail -fmt sarif -out gosec.sarif ./...'
      - uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: gosec.sarif

      # Container/dependency vulnerabilities
      - name: Run Trivy
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          format: 'sarif'
          output: 'trivy.sarif'
      - uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: trivy.sarif

      # Go vulnerability database
      - name: Run govulncheck
        run: |
          go install golang.org/x/vuln/cmd/govulncheck@latest
          govulncheck ./...

      # Dependency scanning
      - name: Dependency review
        uses: actions/dependency-review-action@v4
        if: github.event_name == 'pull_request'
```

**Tools comparison:**
- **gosec:** Go-specific SAST (hardcoded secrets, SQL injection)
- **trivy:** Universal scanner (CVEs in deps, containers, IaC)
- **govulncheck:** Official Go vuln DB (Go-specific CVEs)
- **dependency-review:** GitHub native (PRs only)

## Performance

```.yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true  # Auto-caches go.mod/go.sum

      # Parallel test execution
      - run: go test -race -parallel 4 ./...

      # Artifact caching for build outputs
      - uses: actions/cache@v4
        with:
          path: ./dist
          key: build-${{ github.sha }}

      # Optimized Docker build
      - uses: docker/build-push-action@v5
        with:
          context: .
          cache-from: type=gha
          cache-to: type=gha,mode=max
          platforms: linux/amd64,linux/arm64
```

**Optimization strategies:**
- Use `cache: true` in setup-go (2-5x speedup)
- Enable Docker layer caching with `type=gha`
- Run tests with `-parallel` matching CPU count
- Use GitHub's built-in concurrency limits wisely
- Cache intermediate artifacts between jobs
- Multi-stage Docker builds with minimal final layer

## Common Mistakes

| Mistake | Problem | Fix |
|---------|---------|-----|
| No module caching | Slow builds (2-5min) | Use `cache: true` in setup-go |
| Sequential jobs | Wasted time | Use `needs:` sparingly, parallelize tests |
| Missing security scan | Vulnerabilities in prod | Add gosec/trivy/govulncheck |
| Hardcoded Go version | Outdated quickly | Use matrix with `['1.22', '1.23']` |
| No integration tests in CI | Bugs reach production | Use testcontainers in CI workflow |
| Testing on single OS | Cross-platform bugs | Matrix with ubuntu/macos/windows |
| No race detector | Concurrency bugs | Always use `-race` in CI |
| Skipping `go mod verify` | Compromised dependencies | Run after cache restore |
| Large Docker images | Slow deployments | Multi-stage builds, distroless base |
| No artifact retention | Can't debug failures | Upload logs/coverage with `actions/upload-artifact@v4` |

## See Also

- [Docker](./docker.md)
- [Kubernetes](./kubernetes.md)
- [Linting](./linting.md)
- [Integration Testing](../08-testing/integration.md)
