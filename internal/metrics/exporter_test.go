package metrics

//nolint:gosec // test file with necessary file operations

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExporter_NewExporter(t *testing.T) {
	t.Parallel()

	store := &Store{}
	exporter := NewExporter(store)

	assert.NotNil(t, exporter)
	assert.Same(t, store, exporter.store)
}

func TestExporter_ExportJSON_EmptyMetrics(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	exporter := NewExporter(store)

	data, err := exporter.ExportJSON(nil, "")
	require.NoError(t, err)

	var metrics []Metric
	err = json.Unmarshal(data, &metrics)
	require.NoError(t, err)
	assert.Empty(t, metrics)
}

func TestExporter_ExportJSON_WithMetrics(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	now := time.Now()
	metrics := []Metric{
		{
			SessionID: uuid.New().String(),
			ToolName:  "tool1",
			TokensIn:  10,
			TokensOut: 20,
			Duration:  100 * time.Millisecond,
			Success:   true,
			Timestamp: now,
		},
		{
			SessionID: uuid.New().String(),
			ToolName:  "tool2",
			TokensIn:  30,
			TokensOut: 40,
			Duration:  200 * time.Millisecond,
			Success:   false,
			ErrorMsg:  "test error",
			Timestamp: now,
		},
	}

	for _, m := range metrics {
		err := store.Add(m)
		require.NoError(t, err)
	}

	exporter := NewExporter(store)

	data, err := exporter.ExportJSON(nil, "")
	require.NoError(t, err)

	var result []Metric
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestExporter_ExportJSON_WithFilter(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	m1 := Metric{
		SessionID: uuid.New().String(),
		ToolName:  "tool1",
		TokensIn:  10,
		TokensOut: 20,
		Duration:  100 * time.Millisecond,
		Success:   true,
		Timestamp: time.Now(),
	}

	m2 := Metric{
		SessionID: uuid.New().String(),
		ToolName:  "tool2",
		TokensIn:  30,
		TokensOut: 40,
		Duration:  200 * time.Millisecond,
		Success:   true,
		Timestamp: time.Now(),
	}

	require.NoError(t, store.Add(m1))
	require.NoError(t, store.Add(m2))

	exporter := NewExporter(store)

	filter := func(m Metric) bool {
		return m.ToolName == "tool1"
	}

	data, err := exporter.ExportJSON(filter, "")
	require.NoError(t, err)

	var result []Metric
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "tool1", result[0].ToolName)
}

func TestExporter_ExportJSON_WithFilename(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	m := Metric{
		SessionID: uuid.New().String(),
		ToolName:  "tool1",
		TokensIn:  10,
		TokensOut: 20,
		Duration:  100 * time.Millisecond,
		Success:   true,
		Timestamp: time.Now(),
	}

	require.NoError(t, store.Add(m))

	exporter := NewExporter(store)

	filename := filepath.Join(tempDir, "export.json")
	data, err := exporter.ExportJSON(nil, filename)
	require.NoError(t, err)

	_, err = os.Stat(filename)
	require.NoError(t, err)

	fileData, err := os.ReadFile(filename) // #nosec G304 -- test file
	require.NoError(t, err)
	assert.Equal(t, data, fileData)
}

func TestExporter_ExportCSV_EmptyMetrics(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	exporter := NewExporter(store)

	data, err := exporter.ExportCSV(nil, "")
	require.NoError(t, err)

	lines := strings.Split(string(data), "\n")
	assert.Equal(t, 2, len(lines))
	assert.Contains(t, lines[0], "SessionID")
}

func TestExporter_ExportCSV_WithMetrics(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	now := time.Now()
	m := Metric{
		SessionID: "session-123",
		ToolName:  "test-tool",
		TokensIn:  100,
		TokensOut: 200,
		Duration:  500 * time.Millisecond,
		Success:   true,
		ErrorMsg:  "",
		Timestamp: now,
	}

	require.NoError(t, store.Add(m))

	exporter := NewExporter(store)

	data, err := exporter.ExportCSV(nil, "")
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "session-123")
	assert.Contains(t, content, "test-tool")
	assert.Contains(t, content, "100")
	assert.Contains(t, content, "200")
	assert.Contains(t, content, "true")
}

func TestExporter_ExportCSV_WithSpecialChars(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	m := Metric{
		SessionID: "session-123",
		ToolName:  "test,tool",
		TokensIn:  100,
		TokensOut: 200,
		Duration:  500 * time.Millisecond,
		Success:   false,
		ErrorMsg:  "error, with comma",
		Timestamp: time.Now(),
	}

	require.NoError(t, store.Add(m))

	exporter := NewExporter(store)

	data, err := exporter.ExportCSV(nil, "")
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "error, with comma")
}

func TestExporter_ExportCSV_WithFilename(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	m := Metric{
		SessionID: uuid.New().String(),
		ToolName:  "tool1",
		TokensIn:  10,
		TokensOut: 20,
		Duration:  100 * time.Millisecond,
		Success:   true,
		Timestamp: time.Now(),
	}

	require.NoError(t, store.Add(m))

	exporter := NewExporter(store)

	filename := filepath.Join(tempDir, "export.csv")
	data, err := exporter.ExportCSV(nil, filename)
	require.NoError(t, err)

	_, err = os.Stat(filename)
	require.NoError(t, err)

	fileData, err := os.ReadFile(filename) // #nosec G304 -- test file
	require.NoError(t, err)
	assert.Equal(t, data, fileData)
}

func TestExporter_ExportPrometheus_EmptyMetrics(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	exporter := NewExporter(store)

	data, err := exporter.ExportPrometheus(nil, "")
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "# HELP tool_tokens_total")
	assert.Contains(t, content, "# TYPE tool_tokens_total gauge")
	assert.Contains(t, content, "# HELP tool_duration_seconds")
	assert.Contains(t, content, "# TYPE tool_duration_seconds gauge")
	assert.Contains(t, content, "# HELP tool_success_total")
	assert.Contains(t, content, "# TYPE tool_success_total counter")
}

func TestExporter_ExportPrometheus_WithMetrics(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	m := Metric{
		SessionID: "session-123",
		ToolName:  "test-tool",
		TokensIn:  100,
		TokensOut: 200,
		Duration:  500 * time.Millisecond,
		Success:   true,
		Timestamp: time.Now(),
	}

	require.NoError(t, store.Add(m))

	exporter := NewExporter(store)

	data, err := exporter.ExportPrometheus(nil, "")
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "tool_tokens_total{session=\"session-123\",tool=\"test-tool\",success=\"1\"} 300")
	assert.Contains(t, content, "tool_duration_seconds{session=\"session-123\",tool=\"test-tool\",success=\"1\"} 0.500")
	assert.Contains(t, content, "tool_success_total{session=\"session-123\",tool=\"test-tool\",success=\"1\"} 1")
}

func TestExporter_ExportPrometheus_WithFilter(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	m1 := Metric{
		SessionID: uuid.New().String(),
		ToolName:  "tool1",
		TokensIn:  10,
		TokensOut: 20,
		Duration:  100 * time.Millisecond,
		Success:   true,
		Timestamp: time.Now(),
	}

	m2 := Metric{
		SessionID: uuid.New().String(),
		ToolName:  "tool2",
		TokensIn:  30,
		TokensOut: 40,
		Duration:  200 * time.Millisecond,
		Success:   true,
		Timestamp: time.Now(),
	}

	require.NoError(t, store.Add(m1))
	require.NoError(t, store.Add(m2))

	exporter := NewExporter(store)

	filter := func(m Metric) bool {
		return m.ToolName == "tool1"
	}

	data, err := exporter.ExportPrometheus(filter, "")
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, "tool1")
	assert.NotContains(t, content, "tool2")
}

func TestExporter_ExportPrometheus_WithFilename(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	m := Metric{
		SessionID: uuid.New().String(),
		ToolName:  "tool1",
		TokensIn:  10,
		TokensOut: 20,
		Duration:  100 * time.Millisecond,
		Success:   true,
		Timestamp: time.Now(),
	}

	require.NoError(t, store.Add(m))

	exporter := NewExporter(store)

	filename := filepath.Join(tempDir, "export.prom")
	data, err := exporter.ExportPrometheus(nil, filename)
	require.NoError(t, err)

	_, err = os.Stat(filename)
	require.NoError(t, err)

	fileData, err := os.ReadFile(filename) // #nosec G304 -- test file
	require.NoError(t, err)
	assert.Equal(t, data, fileData)
}

func TestExporter_ExportJSONWithTimestamp(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	m := Metric{
		SessionID: uuid.New().String(),
		ToolName:  "tool1",
		TokensIn:  10,
		TokensOut: 20,
		Duration:  100 * time.Millisecond,
		Success:   true,
		Timestamp: time.Now(),
	}

	require.NoError(t, store.Add(m))

	exporter := NewExporter(store)

	data, filename, err := exporter.ExportJSONWithTimestamp(nil)
	require.NoError(t, err)

	assert.NotNil(t, data)
	assert.Contains(t, filename, "metrics_")
	assert.Contains(t, filename, ".json")
	assert.FileExists(t, filename)

	_ = os.Remove(filename)
}

func TestExporter_ExportCSVWithTimestamp(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	m := Metric{
		SessionID: uuid.New().String(),
		ToolName:  "tool1",
		TokensIn:  10,
		TokensOut: 20,
		Duration:  100 * time.Millisecond,
		Success:   true,
		Timestamp: time.Now(),
	}

	require.NoError(t, store.Add(m))

	exporter := NewExporter(store)

	data, filename, err := exporter.ExportCSVWithTimestamp(nil)
	require.NoError(t, err)

	assert.NotNil(t, data)
	assert.Contains(t, filename, "metrics_")
	assert.Contains(t, filename, ".csv")
	assert.FileExists(t, filename)

	_ = os.Remove(filename)
}

func TestExporter_ExportPrometheusWithTimestamp(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	m := Metric{
		SessionID: uuid.New().String(),
		ToolName:  "tool1",
		TokensIn:  10,
		TokensOut: 20,
		Duration:  100 * time.Millisecond,
		Success:   true,
		Timestamp: time.Now(),
	}

	require.NoError(t, store.Add(m))

	exporter := NewExporter(store)

	data, filename, err := exporter.ExportPrometheusWithTimestamp(nil)
	require.NoError(t, err)

	assert.NotNil(t, data)
	assert.Contains(t, filename, "metrics_")
	assert.Contains(t, filename, ".prom")
	assert.FileExists(t, filename)

	_ = os.Remove(filename)
}

func TestExporter_ExportToFile_CreatesDirectories(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	m := Metric{
		SessionID: uuid.New().String(),
		ToolName:  "tool1",
		TokensIn:  10,
		TokensOut: 20,
		Duration:  100 * time.Millisecond,
		Success:   true,
		Timestamp: time.Now(),
	}

	require.NoError(t, store.Add(m))

	exporter := NewExporter(store)

	filename := filepath.Join(tempDir, "subdir", "nested", "export.json")
	_, err = exporter.ExportJSON(nil, filename)
	require.NoError(t, err)

	_, err = os.Stat(filename)
	require.NoError(t, err)
}

func TestExporter_ExportJSON_MultipleMetrics(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	for i := 0; i < 5; i++ {
		m := Metric{
			SessionID: uuid.New().String(),
			ToolName:  "tool1",
			TokensIn:  i * 10,
			TokensOut: i * 20,
			Duration:  time.Duration(i) * 100 * time.Millisecond,
			Success:   i%2 == 0,
			Timestamp: time.Now(),
		}
		require.NoError(t, store.Add(m))
	}

	exporter := NewExporter(store)

	data, err := exporter.ExportJSON(nil, "")
	require.NoError(t, err)

	var result []Metric
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)
	assert.Len(t, result, 5)
}

func TestExporter_ExportCSV_AllHeadersPresent(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	exporter := NewExporter(store)

	data, err := exporter.ExportCSV(nil, "")
	require.NoError(t, err)

	headers := []string{"SessionID", "ToolName", "TokensIn", "TokensOut", "Duration", "Success", "ErrorMsg", "Timestamp"}
	content := string(data)
	for _, header := range headers {
		assert.Contains(t, content, header)
	}
}

func TestExporter_ExportPrometheus_FailedMetric(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	m := Metric{
		SessionID: "session-123",
		ToolName:  "test-tool",
		TokensIn:  100,
		TokensOut: 200,
		Duration:  500 * time.Millisecond,
		Success:   false,
		ErrorMsg:  "failed",
		Timestamp: time.Now(),
	}

	require.NoError(t, store.Add(m))

	exporter := NewExporter(store)

	data, err := exporter.ExportPrometheus(nil, "")
	require.NoError(t, err)

	content := string(data)
	assert.Contains(t, content, `success="0"`)
}

func TestExporter_ExportPrometheus_SpecialChars(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	m := Metric{
		SessionID: `session"with"quotes`,
		ToolName:  `tool\with\backslash`,
		TokensIn:  10,
		TokensOut: 20,
		Duration:  100 * time.Millisecond,
		Success:   true,
		Timestamp: time.Now(),
	}

	require.NoError(t, store.Add(m))

	exporter := NewExporter(store)

	data, err := exporter.ExportPrometheus(nil, "")
	require.NoError(t, err)

	content := string(data)
	// Check that backslashes and quotes are escaped
	assert.Contains(t, content, `session\"with\"quotes`)
	assert.Contains(t, content, `tool\\with\\backslash`)
}

func TestExporter_ValidatePrometheusMetricName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		metric  string
		wantErr bool
	}{
		{"valid simple", "metric_name", false},
		{"valid with colon", "metric:name", false},
		{"valid starts with letter", "validName", false},
		{"valid starts with underscore", "_valid_name", false},
		{"invalid starts with number", "123invalid", true},
		{"invalid contains dash", "metric-name", true},
		{"invalid contains space", "metric name", true},
		{"invalid contains dot", "metric.name", true},
		{"empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validatePrometheusMetricName(tt.metric)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExporter_ValidatePrometheusLabelValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no special chars", "simple", "simple"},
		{"with backslash", `simple\value`, `simple\\value`},
		{"with quote", `simple"value`, `simple\"value`},
		{"with newline", "simple\nvalue", `simple\nvalue`},
		{"combined", `a"\b\c`, `a\"\\b\\c`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := validatePrometheusLabelValue(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExporter_ExportJSON_FilePermissions(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	m := Metric{
		SessionID: uuid.New().String(),
		ToolName:  "tool1",
		TokensIn:  10,
		TokensOut: 20,
		Duration:  100 * time.Millisecond,
		Success:   true,
		Timestamp: time.Now(),
	}

	require.NoError(t, store.Add(m))

	exporter := NewExporter(store)

	filename := filepath.Join(tempDir, "export.json")
	_, err = exporter.ExportJSON(nil, filename)
	require.NoError(t, err)

	info, err := os.Stat(filename)
	require.NoError(t, err)

	// Check file permissions are 0600 (owner read/write only)
	perm := info.Mode().Perm()
	assert.Equal(t, os.FileMode(0600), perm)
}

func TestExporter_ExportCSV_FilePermissions(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	m := Metric{
		SessionID: uuid.New().String(),
		ToolName:  "tool1",
		TokensIn:  10,
		TokensOut: 20,
		Duration:  100 * time.Millisecond,
		Success:   true,
		Timestamp: time.Now(),
	}

	require.NoError(t, store.Add(m))

	exporter := NewExporter(store)

	filename := filepath.Join(tempDir, "export.csv")
	_, err = exporter.ExportCSV(nil, filename)
	require.NoError(t, err)

	info, err := os.Stat(filename)
	require.NoError(t, err)

	perm := info.Mode().Perm()
	assert.Equal(t, os.FileMode(0600), perm)
}

func TestExporter_ExportPrometheus_FilePermissions(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	m := Metric{
		SessionID: uuid.New().String(),
		ToolName:  "tool1",
		TokensIn:  10,
		TokensOut: 20,
		Duration:  100 * time.Millisecond,
		Success:   true,
		Timestamp: time.Now(),
	}

	require.NoError(t, store.Add(m))

	exporter := NewExporter(store)

	filename := filepath.Join(tempDir, "export.prom")
	_, err = exporter.ExportPrometheus(nil, filename)
	require.NoError(t, err)

	info, err := os.Stat(filename)
	require.NoError(t, err)

	perm := info.Mode().Perm()
	assert.Equal(t, os.FileMode(0600), perm)
}

func TestExporter_ExportJSON_InvalidFilename(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	m := Metric{
		SessionID: uuid.New().String(),
		ToolName:  "tool1",
		TokensIn:  10,
		TokensOut: 20,
		Duration:  100 * time.Millisecond,
		Success:   true,
		Timestamp: time.Now(),
	}

	require.NoError(t, store.Add(m))

	exporter := NewExporter(store)

	// Try to write to a directory path (should fail)
	_, err = exporter.ExportJSON(nil, tempDir)
	assert.Error(t, err)
}

func TestExporter_ExportCSV_InvalidFilename(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	store, err := NewStore(filepath.Join(tempDir, "store.json"), 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	m := Metric{
		SessionID: uuid.New().String(),
		ToolName:  "tool1",
		TokensIn:  10,
		TokensOut: 20,
		Duration:  100 * time.Millisecond,
		Success:   true,
		Timestamp: time.Now(),
	}

	require.NoError(t, store.Add(m))

	exporter := NewExporter(store)

	// Try to write to a directory path (should fail)
	_, err = exporter.ExportCSV(nil, tempDir)
	assert.Error(t, err)
}
