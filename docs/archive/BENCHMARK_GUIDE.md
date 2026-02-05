# Benchmark Testing Guide

## Overview

This document provides guidance for running and interpreting the performance benchmarks for ACP, CLI, and API interaction modes.

## Prerequisites

### For ACP Benchmarks (`acp_bench_test.go`)

**Required**: Running `opencode acp` server process

```bash
# Start opencode ACP server in a separate terminal
opencode acp

# Run ACP benchmarks
go test -bench=BenchmarkACP -benchmem ./internal/opencode/
```

**Note**: If ACP server is not running, benchmarks will be skipped.

### For API Benchmarks (`anthropic_bench_test.go`)

**Required**: Valid `ANTHROPIC_API_KEY` environment variable

```bash
# Set API key
export ANTHROPIC_API_KEY="your-api-key-here"

# Run API benchmarks
go test -bench=BenchmarkAPI -benchmem ./internal/provider/
```

**Note**: If API key is not set, benchmarks will be skipped.

### For CLI Benchmarks (`cli_bench_test.go`)

**Required**: `opencode` binary in PATH

```bash
# Build binary if needed
make build

# Run CLI benchmarks
go test -bench=BenchmarkCLI -benchmem ./internal/opencode/
```

**Note**: CLI benchmarks may fail if `opencode` binary is not installed.

## Running Benchmarks

### Quick Validation (Single Run)

```bash
# Run each benchmark once to verify functionality
go test -bench=. -benchtime=1x ./internal/opencode/ ./internal/provider/
```

### Standard Performance Test

```bash
# Run each benchmark for 1 second (default)
go test -bench=. -benchmem ./internal/opencode/ ./internal/provider/
```

### Extended Performance Test

```bash
# Run each benchmark for 10 seconds
go test -bench=. -benchmem -benchtime=10s ./internal/opencode/ ./internal/provider/
```

### Memory Profiling

```bash
# Generate memory profile
go test -bench=. -memprofile=mem.prof ./internal/opencode/

# Analyze memory profile
go tool pprof mem.prof
```

### CPU Profiling

```bash
# Generate CPU profile
go test -bench=. -cpuprofile=cpu.prof ./internal/opencode/

# Analyze CPU profile
go tool pprof cpu.prof
```

## Benchmark Scenarios

### ACP Mode Benchmarks

1. **BenchmarkACP_Startup**: Client initialization time
2. **BenchmarkACP_SimplePrompt**: Simple query response time
3. **BenchmarkACP_ComplexPrompt**: Complex query response time
4. **BenchmarkACP_StreamingUpdates**: Streaming update handling
5. **BenchmarkACP_ConcurrentRequests**: Concurrent request handling
6. **BenchmarkACP_MemoryUsage**: Memory allocation patterns
7. **BenchmarkACP_Serialization**: JSON-RPC serialization overhead
8. **BenchmarkACP_SessionManagement**: Session creation overhead
9. **BenchmarkACP_LongRunningTask**: Long-running task performance
10. **BenchmarkACP_Throughput**: Maximum requests per second

### CLI Mode Benchmarks

1. **BenchmarkCLI_Startup**: Binary startup time
2. **BenchmarkCLI_SimpleCommand**: Simple command execution
3. **BenchmarkCLI_ComplexCommand**: Complex command execution
4. **BenchmarkCLI_LongPrompt**: Long prompt handling
5. **BenchmarkCLI_ConcurrentCommands**: Concurrent command execution
6. **BenchmarkCLI_WithTimeout**: Timeout handling
7. **BenchmarkCLI_NonBlocking**: Async execution
8. **BenchmarkCLI_ArgumentBuilding**: Command argument construction
9. **BenchmarkCLI_EnvironmentSetup**: Environment variable setup
10. **BenchmarkCLI_MemoryUsage**: Memory allocation patterns
11. **BenchmarkCLI_OutputParsing**: Output text parsing
12. **BenchmarkCLI_Throughput**: Maximum requests per second
13. **BenchmarkCLI_LargeOutput**: Large output processing

### API Mode Benchmarks

1. **BenchmarkAPI_ClientCreation**: Client initialization time
2. **BenchmarkAPI_SimpleComplete**: Simple completion request
3. **BenchmarkAPI_ComplexComplete**: Complex completion request
4. **BenchmarkAPI_StreamingComplete**: Streaming completion
5. **BenchmarkAPI_CompleteWithHistory**: Conversation context handling
6. **BenchmarkAPI_StreamWithHistory**: Streaming with history
7. **BenchmarkAPI_ConcurrentRequests**: Concurrent request handling
8. **BenchmarkAPI_MemoryUsage**: Memory allocation patterns
9. **BenchmarkAPI_DifferentModels**: Different model performance
10. **BenchmarkAPI_Throughput**: Maximum requests per second
11. **BenchmarkAPI_LongPrompt**: Long prompt handling
12. **BenchmarkAPI_StreamingLongPrompt**: Long prompt streaming
13. **BenchmarkAPI_RequestSerialization**: JSON request marshaling
14. **BenchmarkAPI_ResponseParsing**: JSON response unmarshaling

## Interpreting Results

### Key Metrics

- **ns/op**: Nanoseconds per operation (lower is better)
- **B/op**: Bytes allocated per operation (lower is better)
- **allocs/op**: Memory allocations per operation (lower is better)
- **Custom metrics**: Additional metrics reported via `b.ReportMetric()`

### Latency Metrics

Custom latency metrics are reported in microseconds (μs):
- `startup_us`: Client/process startup time
- `latency_us`: Request/response latency
- `avg_latency_us`: Average latency across requests
- `session_creation_us`: Session initialization time
- `process_us`: Processing time
- `task_time_ms`: Task completion time

### Throughput Metrics

- `req_per_sec`: Requests per second (higher is better)
- `ops`: Operations per second (higher is better)

### Count Metrics

- `chunks`: Number of streaming chunks received
- `updates`: Number of update notifications

### Memory Metrics

- `alloc_bytes`: Memory allocated during benchmark

## Performance Comparison

### Expected Results (Approximate)

| Benchmark | ACP Mode | CLI Mode | API Mode |
|-----------|----------|----------|----------|
| Startup | 190 μs | 640 ms | < 1 μs |
| Simple Request | ~1 μs | ~500 ms | ~100 μs |
| Complex Request | ~1 μs | ~420 ms | ~100 μs |
| Memory/Req | 1 KB | 40 KB | 9.4 KB |
| Throughput | 988K ops/s | 2.25 req/s | 9,650 ops/s |

**Note**: Actual results will vary based on:
- System hardware
- Network conditions (for API mode)
- Process state (for ACP mode)
- Current system load

## Troubleshooting

### ACP Benchmarks Skipped

**Problem**: All ACP benchmarks are skipped

**Solution**:
```bash
# Start ACP server
opencode acp

# Verify server is running
ps aux | grep "opencode acp"
```

### API Benchmarks Skipped

**Problem**: All API benchmarks are skipped

**Solution**:
```bash
# Check API key
echo $ANTHROPIC_API_KEY

# If empty, set it
export ANTHROPIC_API_KEY="your-key"
```

### CLI Benchmarks Fail

**Problem**: CLI benchmarks fail with "command not found"

**Solution**:
```bash
# Build binary
make build

# Verify binary exists
ls -lh bin/ent

# Add to PATH
export PATH="$PATH:$(pwd)/bin"
```

### Benchmark Timeouts

**Problem**: Benchmarks timeout or take too long

**Solution**:
```bash
# Run with shorter duration
go test -bench=. -benchtime=100ms ./internal/opencode/

# Or run only specific benchmarks
go test -bench=BenchmarkStartup ./internal/opencode/
```

## Continuous Benchmarking

For tracking performance over time, consider:

1. **CI/CD Integration**: Add benchmarks to CI pipeline
2. **Baseline Tracking**: Store baseline results
3. **Regression Detection**: Alert on performance degradation
4. **Historical Trends**: Track performance over time

### Example CI Integration

```yaml
# .github/workflows/bench.yml
name: Benchmarks

on: [push, pull_request]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      - run: go test -bench=. -benchmem ./internal/opencode/ ./internal/provider/
```

## Summary

- **ACP Mode**: Best for session-based, high-frequency interactions
- **CLI Mode**: Best for one-off scripts and simple automation
- **API Mode**: Best for direct HTTP access and low-frequency requests

For detailed results and analysis, see `BENCHMARK_RESULTS.md`.
