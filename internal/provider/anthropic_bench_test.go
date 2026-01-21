package provider

import (
	"context"
	"encoding/json"
	"runtime"
	"sync"
	"testing"
	"time"
)

var (
	testAPIClient *Client
	onceAPI       sync.Once
)

func setupAPIBenchmark(b *testing.B) *Client {
	var err error
	onceAPI.Do(func() {
		testAPIClient, err = NewAnthropicClient(nil)
		if err != nil {
			return
		}
	})

	if testAPIClient == nil {
		b.Skip("API client not available (ANTHROPIC_API_KEY not set)")
	}

	return testAPIClient
}

func BenchmarkAPI_ClientCreation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		start := time.Now()

		_, err := NewAnthropicClient(nil)
		if err != nil {
			b.Skipf("client creation failed: %v", err)
			continue
		}

		createTime := time.Since(start)
		b.ReportMetric(float64(createTime.Microseconds()), "creation_us")
	}
}

func BenchmarkAPI_SimpleComplete(b *testing.B) {
	client := setupAPIBenchmark(b)
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		start := time.Now()

		_, err := client.Complete(ctx, ModelHaiku, "Say 'Hello, World!'")
		if err != nil {
			b.Error(err)
			continue
		}

		latency := time.Since(start)
		b.StopTimer()
		b.ReportMetric(float64(latency.Microseconds()), "latency_us")
		b.StartTimer()
	}
}

func BenchmarkAPI_ComplexComplete(b *testing.B) {
	client := setupAPIBenchmark(b)
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	complexPrompt := `Analyze the following code and suggest improvements:

package main

import "fmt"

func processItems(items []string) []string {
	var result []string
	for _, item := range items {
		if len(item) > 5 {
			result = append(result, item)
		}
	}
	return result
}

func main() {
	items := []string{"short", "longer", "very long item", "tiny"}
	processed := processItems(items)
	fmt.Println(processed)
}

Provide recommendations for:
1. Performance optimization
2. Memory efficiency
3. Error handling
4. Code organization`

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		start := time.Now()

		_, err := client.Complete(ctx, ModelHaiku, complexPrompt)
		if err != nil {
			b.Error(err)
			continue
		}

		latency := time.Since(start)
		b.StopTimer()
		b.ReportMetric(float64(latency.Microseconds()), "latency_us")
		b.StartTimer()
	}
}

func BenchmarkAPI_StreamingComplete(b *testing.B) {
	client := setupAPIBenchmark(b)
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var chunks []string

		start := time.Now()

		err := client.Stream(ctx, ModelHaiku, "Count from 1 to 10", func(text string) {
			chunks = append(chunks, text)
		})

		if err != nil {
			b.Error(err)
			continue
		}

		latency := time.Since(start)
		b.StopTimer()

		b.ReportMetric(float64(latency.Microseconds()), "latency_us")
		b.ReportMetric(float64(len(chunks)), "chunks")

		b.StartTimer()
	}
}

func BenchmarkAPI_CompleteWithHistory(b *testing.B) {
	client := setupAPIBenchmark(b)
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages := []Message{
		{
			Role:    "user",
			Content: "What is 2+2?",
		},
		{
			Role:    "assistant",
			Content: "2+2 equals 4.",
		},
		{
			Role:    "user",
			Content: "What about 3+3?",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		start := time.Now()

		_, err := client.CompleteWithHistory(ctx, ModelHaiku, messages)
		if err != nil {
			b.Error(err)
			continue
		}

		latency := time.Since(start)
		b.StopTimer()
		b.ReportMetric(float64(latency.Microseconds()), "latency_us")
		b.StartTimer()
	}
}

func BenchmarkAPI_StreamWithHistory(b *testing.B) {
	client := setupAPIBenchmark(b)
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages := []Message{
		{
			Role:    "user",
			Content: "Tell me a short joke",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var chunks []string

		start := time.Now()

		err := client.StreamWithHistory(ctx, ModelHaiku, messages, func(text string) {
			chunks = append(chunks, text)
		})

		if err != nil {
			b.Error(err)
			continue
		}

		latency := time.Since(start)
		b.StopTimer()

		b.ReportMetric(float64(latency.Microseconds()), "latency_us")
		b.ReportMetric(float64(len(chunks)), "chunks")

		b.StartTimer()
	}
}

func BenchmarkAPI_ConcurrentRequests(b *testing.B) {
	client := setupAPIBenchmark(b)
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	concurrency := 5

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
			_, err := client.Complete(ctx, ModelHaiku, "Say hello")
			if err == nil {
				latency := time.Since(start)
				b.StopTimer()
				totalLatency += latency
				b.StartTimer()
			}
		}()
	}

	wg.Wait()
	if b.N > 0 {
		avgLatency := totalLatency / time.Duration(b.N)
		b.ReportMetric(float64(avgLatency.Microseconds()), "avg_latency_us")
	}

	cancel()
}

func BenchmarkAPI_MemoryUsage(b *testing.B) {
	client := setupAPIBenchmark(b)
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b.ResetTimer()

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	for i := 0; i < b.N; i++ {
		_, _ = client.Complete(ctx, ModelHaiku, "Test prompt")
	}

	runtime.ReadMemStats(&m2)

	allocDiff := m2.Alloc - m1.Alloc
	b.ReportMetric(float64(allocDiff), "alloc_bytes")
}

func BenchmarkAPI_DifferentModels(b *testing.B) {
	client := setupAPIBenchmark(b)
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	models := []Model{ModelHaiku, ModelSonnet}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		model := models[i%len(models)]

		start := time.Now()

		_, err := client.Complete(ctx, model, "Say hello")
		if err != nil {
			continue
		}

		latency := time.Since(start)
		b.StopTimer()
		b.ReportMetric(float64(latency.Microseconds()), "latency_us")
		b.StartTimer()
	}
}

func BenchmarkAPI_Throughput(b *testing.B) {
	client := setupAPIBenchmark(b)
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b.ResetTimer()
	b.ReportAllocs()

	startTime := time.Now()
	var wg sync.WaitGroup

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = client.Complete(ctx, ModelHaiku, "Hello")
		}()
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	throughput := float64(b.N) / totalTime.Seconds()
	b.ReportMetric(throughput, "req_per_sec")

	cancel()
}

func BenchmarkAPI_LongPrompt(b *testing.B) {
	client := setupAPIBenchmark(b)
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	longPrompt := makeString(2000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		start := time.Now()

		_, err := client.Complete(ctx, ModelHaiku, longPrompt)
		if err != nil {
			continue
		}

		latency := time.Since(start)
		b.StopTimer()
		b.ReportMetric(float64(latency.Microseconds()), "latency_us")
		b.StartTimer()
	}
}

func BenchmarkAPI_StreamingLongPrompt(b *testing.B) {
	client := setupAPIBenchmark(b)
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	longPrompt := makeString(2000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var chunks []string

		start := time.Now()

		err := client.Stream(ctx, ModelHaiku, longPrompt, func(text string) {
			chunks = append(chunks, text)
		})

		if err != nil {
			continue
		}

		latency := time.Since(start)
		b.StopTimer()

		b.ReportMetric(float64(latency.Microseconds()), "latency_us")
		b.ReportMetric(float64(len(chunks)), "chunks")

		b.StartTimer()
	}
}

func BenchmarkAPI_RequestSerialization(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := Request{
			Model:     string(ModelHaiku),
			MaxTokens: 4096,
			Messages: []Message{
				{
					Role:    "user",
					Content: "Test prompt",
				},
			},
			Stream: false,
		}

		start := time.Now()

		_, err := json.Marshal(req)
		if err != nil {
			b.Error(err)
			continue
		}

		serializeTime := time.Since(start)
		b.StopTimer()
		b.ReportMetric(float64(serializeTime.Nanoseconds()), "serialize_ns")
		b.StartTimer()
	}
}

func BenchmarkAPI_ResponseParsing(b *testing.B) {
	b.ReportAllocs()

	jsonResponse := `{
		"id": "msg_abc123",
		"type": "message",
		"role": "assistant",
		"content": [
			{
				"type": "text",
				"text": "Hello, World!"
			}
		],
		"model": "claude-3-haiku-20240307",
		"stop_reason": "end_turn"
	}`

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		start := time.Now()

		var event StreamEvent
		err := json.Unmarshal([]byte(jsonResponse), &event)
		if err != nil {
			b.Error(err)
			continue
		}

		parseTime := time.Since(start)
		b.StopTimer()
		b.ReportMetric(float64(parseTime.Nanoseconds()), "parse_ns")
		b.StartTimer()
	}
}

func makeString(n int) string {
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		result[i] = byte('a' + (i % 26))
	}
	return string(result)
}
