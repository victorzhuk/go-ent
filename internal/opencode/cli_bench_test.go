package opencode

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func BenchmarkCLI_Startup(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		start := time.Now()

		cmd := exec.Command("opencode", "--version")

		err := cmd.Run()
		if err != nil {
			b.Skipf("opencode not available: %v", err)
			continue
		}

		startupTime := time.Since(start)
		b.ReportMetric(float64(startupTime.Microseconds()), "startup_us")
	}
}

func BenchmarkCLI_SimpleCommand(b *testing.B) {
	client := NewCLIClient("")

	simplePrompt := "Hello"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		start := time.Now()

		result, err := client.Run(ctx, "", "", simplePrompt)
		if err != nil && result == nil {
			b.Skipf("CLI execution failed: %v", err)
			cancel()
			continue
		}

		latency := time.Since(start)
		b.StopTimer()
		b.ReportMetric(float64(latency.Microseconds()), "latency_us")
		b.StartTimer()

		cancel()
	}
}

func BenchmarkCLI_ComplexCommand(b *testing.B) {
	client := NewCLIClient("")

	complexPrompt := `Analyze the following code and suggest improvements:

func processUserData(data map[string]interface{}) (string, error) {
	var results []string
	for key, value := range data {
		strVal := fmt.Sprintf("%v", value)
		if len(strVal) > 0 {
			results = append(results, fmt.Sprintf("%s: %s", key, strVal))
		}
	}
	return fmt.Sprintf("Processed %d items", len(results)), nil
}

Consider:
1. Performance
2. Error handling
3. Memory usage
4. Code clarity`

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		start := time.Now()

		_, err := client.Run(ctx, "", "", complexPrompt)
		if err != nil {
			cancel()
			continue
		}

		latency := time.Since(start)
		b.StopTimer()
		b.ReportMetric(float64(latency.Microseconds()), "latency_us")
		b.StartTimer()

		cancel()
	}
}

func BenchmarkCLI_LongPrompt(b *testing.B) {
	client := NewCLIClient("")

	longPrompt := fmt.Sprintf("Analyze this text: %s", makeString(1000))

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

		start := time.Now()

		_, err := client.Run(ctx, "", "", longPrompt)
		if err != nil {
			cancel()
			continue
		}

		latency := time.Since(start)
		b.StopTimer()
		b.ReportMetric(float64(latency.Microseconds()), "latency_us")
		b.StartTimer()

		cancel()
	}
}

func BenchmarkCLI_ConcurrentCommands(b *testing.B) {
	client := NewCLIClient("")
	concurrency := 10

	b.ResetTimer()
	b.ReportAllocs()

	var wg sync.WaitGroup
	totalLatency := time.Duration(0)

	for i := 0; i < b.N; i++ {
		if i%concurrency == 0 {
			b.StopTimer()
			wg.Wait()
			b.StartTimer()
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

			start := time.Now()
			_, err := client.Run(ctx, "", "", "Hello")
			if err == nil {
				latency := time.Since(start)
				b.StopTimer()
				totalLatency += latency
				b.StartTimer()
			}

			cancel()
		}()
	}

	wg.Wait()
	if b.N > 0 {
		avgLatency := totalLatency / time.Duration(b.N)
		b.ReportMetric(float64(avgLatency.Microseconds()), "avg_latency_us")
	}
}

func BenchmarkCLI_WithTimeout(b *testing.B) {
	client := NewCLIClient("")
	timeout := 5 * time.Second

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		start := time.Now()

		_, err := client.RunWithTimeout("", "", "Say hello", timeout)
		if err != nil {
			continue
		}

		latency := time.Since(start)
		b.StopTimer()
		b.ReportMetric(float64(latency.Microseconds()), "latency_us")
		b.StartTimer()
	}
}

func BenchmarkCLI_NonBlocking(b *testing.B) {
	client := NewCLIClient("")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		start := time.Now()

		cmd, resultChan := client.RunNonBlocking(ctx, "", "", "Hello")

		result := <-resultChan
		if result.Error != nil {
			cancel()
			continue
		}

		_ = cmd.Process
		latency := time.Since(start)
		b.StopTimer()
		b.ReportMetric(float64(latency.Microseconds()), "latency_us")
		b.StartTimer()

		cancel()
	}
}

func BenchmarkCLI_ArgumentBuilding(b *testing.B) {
	client := NewCLIClient("")

	providers := []string{"anthropic", "openai", "moonshot"}
	models := []string{"claude-3-opus", "gpt-4", "glm-4"}
	prompts := []string{"Hello", "Test", "Example"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		start := time.Now()

		args := client.buildArgs(
			providers[i%len(providers)],
			models[i%len(models)],
			prompts[i%len(prompts)],
		)

		buildTime := time.Since(start)
		b.StopTimer()
		b.ReportMetric(float64(buildTime.Nanoseconds()), "arg_build_ns")

		if len(args) > 0 {
			_ = args[0]
		}
		b.StartTimer()
	}
}

func BenchmarkCLI_EnvironmentSetup(b *testing.B) {
	client := NewCLIClient("/test/config/path")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		cmd := exec.CommandContext(ctx, "opencode")

		start := time.Now()
		client.setEnvironment(cmd)
		envTime := time.Since(start)

		b.StopTimer()
		b.ReportMetric(float64(envTime.Nanoseconds()), "env_setup_ns")

		_ = cmd.Env
		b.StartTimer()
	}
}

func BenchmarkCLI_MemoryUsage(b *testing.B) {
	client := NewCLIClient("")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b.ResetTimer()

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	for i := 0; i < b.N; i++ {
		_, _ = client.Run(ctx, "", "", "Test prompt")
	}

	runtime.ReadMemStats(&m2)

	allocDiff := m2.Alloc - m1.Alloc
	b.ReportMetric(float64(allocDiff), "alloc_bytes")
}

func BenchmarkCLI_OutputParsing(b *testing.B) {
	outputs := []string{
		"Line 1\nLine 2\nLine 3",
		"\n\nLine 1\n\nLine 2\n\n",
		"  Line 1  \n  \n  Line 2  ",
		"",
		strings.Repeat("output line\n", 1000),
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		start := time.Now()

		result := ParseCLIOutput(outputs[i%len(outputs)])

		parsingTime := time.Since(start)
		b.StopTimer()
		b.ReportMetric(float64(parsingTime.Nanoseconds()), "parse_ns")

		_ = result
		b.StartTimer()
	}
}

func BenchmarkCLI_Throughput(b *testing.B) {
	client := NewCLIClient("")

	b.ResetTimer()
	b.ReportAllocs()

	startTime := time.Now()
	var wg sync.WaitGroup

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			_, _ = client.Run(ctx, "", "", "Hello")
			cancel()
		}()
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	throughput := float64(b.N) / totalTime.Seconds()
	b.ReportMetric(throughput, "req_per_sec")
}

func BenchmarkCLI_LargeOutput(b *testing.B) {
	client := NewCLIClient("")

	expectedOutput := strings.Repeat("output line ", 1000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		args := client.buildArgs("", "", "test")
		cmd := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf("echo '%s'", expectedOutput))
		cmd.Args = append([]string{"opencode"}, args...)
		client.setEnvironment(cmd)

		start := time.Now()
		_ = cmd.Run()
		processTime := time.Since(start)

		b.StopTimer()
		b.ReportMetric(float64(processTime.Microseconds()), "process_us")
		b.StartTimer()

		cancel()
	}
}

func makeString(n int) string {
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		result[i] = byte('a' + (i % 26))
	}
	return string(result)
}
