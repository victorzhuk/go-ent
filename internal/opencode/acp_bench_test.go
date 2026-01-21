package opencode

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"
)

var (
	testACPClient *ACPClient
	onceACP       sync.Once
)

func setupACPBenchmark(b *testing.B) *ACPClient {
	var err error
	onceACP.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		cfg := Config{
			ClientName: "benchmark-client",
		}

		testACPClient, err = NewACPClient(ctx, cfg)
		if err != nil {
			b.Skipf("ACP not available: %v", err)
			return
		}

		if err := testACPClient.Initialize(ctx); err != nil {
			b.Skipf("ACP initialization failed: %v", err)
			return
		}
	})

	if testACPClient == nil {
		b.Skip("ACP client not available")
	}

	return testACPClient
}

func BenchmarkACP_Startup(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		start := time.Now()

		cfg := Config{
			ClientName: "benchmark-client",
		}

		client, err := NewACPClient(ctx, cfg)
		if err != nil {
			b.Skipf("cannot create client: %v", err)
			continue
		}

		startupTime := time.Since(start)
		b.ReportMetric(float64(startupTime.Microseconds()), "startup_us")

		_ = client.Close()
		cancel()
	}
}

func BenchmarkACP_SimplePrompt(b *testing.B) {
	client := setupACPBenchmark(b)
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	provider := "test"
	model := "test-model"
	if _, err := client.SessionNew(ctx, provider, model, nil); err != nil {
		b.Skipf("session creation failed: %v", err)
		return
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		start := time.Now()

		_, err := client.SessionPrompt(ctx, "What is 2+2?", nil, nil)
		if err != nil {
			b.Error(err)
		}

		b.StopTimer()
		latency := time.Since(start)
		b.ReportMetric(float64(latency.Microseconds()), "latency_us")
		b.StartTimer()
	}
}

func BenchmarkACP_ComplexPrompt(b *testing.B) {
	client := setupACPBenchmark(b)
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	provider := "test"
	model := "test-model"
	if _, err := client.SessionNew(ctx, provider, model, nil); err != nil {
		b.Skipf("session creation failed: %v", err)
		return
	}

	complexPrompt := `Analyze the following code snippet and suggest improvements:
func processData(data []string) []string {
	var result []string
	for _, item := range data {
		if len(item) > 0 {
			result = append(result, item)
		}
	}
	return result
}

Consider:
1. Performance optimization
2. Memory usage
3. Edge cases
4. Code readability`

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		start := time.Now()

		_, err := client.SessionPrompt(ctx, complexPrompt, nil, nil)
		if err != nil {
			b.Error(err)
		}

		b.StopTimer()
		latency := time.Since(start)
		b.ReportMetric(float64(latency.Microseconds()), "latency_us")
		b.StartTimer()
	}
}

func BenchmarkACP_StreamingUpdates(b *testing.B) {
	client := setupACPBenchmark(b)
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	provider := "test"
	model := "test-model"
	if _, err := client.SessionNew(ctx, provider, model, nil); err != nil {
		b.Skipf("session creation failed: %v", err)
		return
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		updateCount := 0
		updateDone := make(chan struct{})

		go func() {
			for range client.Updates() {
				updateCount++
			}
			close(updateDone)
		}()

		start := time.Now()

		_, err := client.SessionPrompt(ctx, "Count from 1 to 10", nil, nil)
		if err != nil {
			b.Error(err)
		}

		latency := time.Since(start)
		b.StopTimer()

		time.Sleep(100 * time.Millisecond)

		b.ReportMetric(float64(latency.Microseconds()), "latency_us")
		b.ReportMetric(float64(updateCount), "updates")

		<-updateDone
		b.StartTimer()
	}
}

func BenchmarkACP_ConcurrentRequests(b *testing.B) {
	client := setupACPBenchmark(b)
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	provider := "test"
	model := "test-model"
	if _, err := client.SessionNew(ctx, provider, model, nil); err != nil {
		b.Skipf("session creation failed: %v", err)
		return
	}

	concurrency := 10
	prompt := "Say hello"

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

			start := time.Now()
			_, err := client.SessionPrompt(ctx, prompt, nil, nil)
			if err != nil {
				b.Error(err)
			}
			latency := time.Since(start)

			b.StopTimer()
			totalLatency += latency
			b.StartTimer()
		}()
	}

	wg.Wait()
	avgLatency := totalLatency / time.Duration(b.N)
	b.ReportMetric(float64(avgLatency.Microseconds()), "avg_latency_us")
}

func BenchmarkACP_MemoryUsage(b *testing.B) {
	client := setupACPBenchmark(b)
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	provider := "test"
	model := "test-model"
	if _, err := client.SessionNew(ctx, provider, model, nil); err != nil {
		b.Skipf("session creation failed: %v", err)
		return
	}

	b.ResetTimer()

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	for i := 0; i < b.N; i++ {
		_, _ = client.SessionPrompt(ctx, "Test prompt", nil, nil)
	}

	runtime.ReadMemStats(&m2)

	allocDiff := m2.Alloc - m1.Alloc
	b.ReportMetric(float64(allocDiff), "alloc_bytes")
}

func BenchmarkACP_Serialization(b *testing.B) {
	client := setupACPBenchmark(b)
	if client == nil {
		return
	}

	b.ReportAllocs()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	params := SessionPromptParams{
		SessionID: "test-session",
		Prompt:    "Test prompt",
		Context: []MessageContext{
			{Role: "user", Content: "Previous message"},
		},
		Options: map[string]any{
			"max_tokens": 1000,
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		start := time.Now()

		err := client.sendRequest(ctx, "session/prompt", params, &SessionPromptResult{})
		if err != nil {
			continue
		}

		serializeTime := time.Since(start)
		b.StopTimer()
		b.ReportMetric(float64(serializeTime.Microseconds()), "serialize_us")
		b.StartTimer()
	}

	cancel()
}

func BenchmarkACP_SessionManagement(b *testing.B) {
	client := setupACPBenchmark(b)
	if client == nil {
		return
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		start := time.Now()

		_, err := client.SessionNew(ctx, "test-provider", "test-model", nil)
		if err != nil {
			cancel()
			continue
		}

		sessionCreateTime := time.Since(start)
		b.StopTimer()
		b.ReportMetric(float64(sessionCreateTime.Microseconds()), "session_creation_us")

		cancel()
		b.StartTimer()
	}
}

func BenchmarkACP_LongRunningTask(b *testing.B) {
	client := setupACPBenchmark(b)
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	provider := "test"
	model := "test-model"
	if _, err := client.SessionNew(ctx, provider, model, nil); err != nil {
		b.Skipf("session creation failed: %v", err)
		return
	}

	longPrompt := `Analyze and explain the following complex algorithm in detail:

1. Read the code
2. Identify the algorithm
3. Explain its purpose
4. Analyze time complexity
5. Analyze space complexity
6. Identify potential optimizations
7. Discuss edge cases
8. Provide examples

func quicksort(arr []int) []int {
	if len(arr) <= 1 {
		return arr
	}
	pivot := arr[len(arr)/2]
	left := []int{}
	right := []int{}
	for _, x := range arr {
		if x < pivot {
			left = append(left, x)
		} else if x > pivot {
			right = append(right, x)
		}
	}
	return append(quicksort(left), pivot, quicksort(right)...)
}`

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		start := time.Now()

		_, err := client.SessionPrompt(ctx, longPrompt, nil, nil)
		if err != nil {
			b.Error(err)
		}

		b.StopTimer()
		taskTime := time.Since(start)
		b.ReportMetric(float64(taskTime.Milliseconds()), "task_time_ms")
		b.StartTimer()
	}
}

func BenchmarkACP_Throughput(b *testing.B) {
	client := setupACPBenchmark(b)
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	provider := "test"
	model := "test-model"
	if _, err := client.SessionNew(ctx, provider, model, nil); err != nil {
		b.Skipf("session creation failed: %v", err)
		return
	}

	b.ResetTimer()
	b.ReportAllocs()

	startTime := time.Now()
	var wg sync.WaitGroup

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = client.SessionPrompt(ctx, "Hello", nil, nil)
		}()
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	throughput := float64(b.N) / totalTime.Seconds()
	b.ReportMetric(throughput, "req_per_sec")
}
