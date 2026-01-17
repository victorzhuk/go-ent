package metrics

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// Prometheus metric name regex: must start with [a-zA-Z_:], followed by [a-zA-Z0-9_:]
	metricNameRegex = regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)
)

// validatePrometheusLabelValue escapes label values per Prometheus spec
// Backslashes and quotes need to be escaped with backslash
func validatePrometheusLabelValue(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	value = strings.ReplaceAll(value, "\n", "\\n")
	return value
}

// validatePrometheusMetricName checks if a metric name is valid per Prometheus spec
func validatePrometheusMetricName(name string) error {
	if !metricNameRegex.MatchString(name) {
		return fmt.Errorf("invalid prometheus metric name: %s", name)
	}
	return nil
}

type Exporter struct {
	store *Store
}

func NewExporter(store *Store) *Exporter {
	return &Exporter{store: store}
}

func (e *Exporter) ExportJSON(filter Filter, filename string) ([]byte, error) {
	metrics := e.applyFilter(filter)

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}

	if filename != "" {
		if err := e.exportToFile(data, filename); err != nil {
			return nil, fmt.Errorf("export file: %w", err)
		}
		slog.Info("metrics exported",
			"format", "json",
			"filename", filename,
			"records", len(metrics),
			"file_size_bytes", len(data),
		)
	}

	return data, nil
}

func (e *Exporter) ExportCSV(filter Filter, filename string) ([]byte, error) {
	metrics := e.applyFilter(filter)

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	headers := []string{"SessionID", "ToolName", "TokensIn", "TokensOut", "Duration", "Success", "ErrorMsg", "Timestamp"}
	if err := w.Write(headers); err != nil {
		return nil, fmt.Errorf("write headers: %w", err)
	}

	for _, m := range metrics {
		record := []string{
			m.SessionID,
			m.ToolName,
			strconv.Itoa(m.TokensIn),
			strconv.Itoa(m.TokensOut),
			m.Duration.String(),
			strconv.FormatBool(m.Success),
			m.ErrorMsg,
			m.Timestamp.Format(time.RFC3339),
		}
		if err := w.Write(record); err != nil {
			return nil, fmt.Errorf("write record: %w", err)
		}
	}

	w.Flush()
	// Flush ensures all buffered data is written; Error() must be checked after Flush
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("flush csv: %w", err)
	}

	if filename != "" {
		if err := e.exportToFile(buf.Bytes(), filename); err != nil {
			return nil, fmt.Errorf("export file: %w", err)
		}
		slog.Info("metrics exported",
			"format", "csv",
			"filename", filename,
			"records", len(metrics),
			"file_size_bytes", buf.Len(),
		)
	}

	return buf.Bytes(), nil
}

func (e *Exporter) ExportPrometheus(filter Filter, filename string) ([]byte, error) {
	metrics := e.applyFilter(filter)

	var buf bytes.Buffer

	buf.WriteString("# HELP tool_tokens_total Total tokens consumed per tool\n")
	buf.WriteString("# TYPE tool_tokens_total gauge\n")
	buf.WriteString("# HELP tool_duration_seconds Tool execution duration in seconds\n")
	buf.WriteString("# TYPE tool_duration_seconds gauge\n")
	buf.WriteString("# HELP tool_success_total Total tool execution success count\n")
	buf.WriteString("# TYPE tool_success_total counter\n")

	for _, m := range metrics {
		success := "0"
		if m.Success {
			success = "1"
		}

		tokensTotal := m.TokensIn + m.TokensOut
		label := fmt.Sprintf(`{session="%s",tool="%s",success="%s"}`,
			validatePrometheusLabelValue(m.SessionID),
			validatePrometheusLabelValue(m.ToolName),
			success)

		buf.WriteString(fmt.Sprintf("tool_tokens_total%s %d\n", label, tokensTotal))
		buf.WriteString(fmt.Sprintf("tool_duration_seconds%s %.3f\n", label, m.Duration.Seconds()))
		buf.WriteString(fmt.Sprintf("tool_success_total%s %s\n", label, success))
	}

	if filename != "" {
		if err := e.exportToFile(buf.Bytes(), filename); err != nil {
			return nil, fmt.Errorf("export file: %w", err)
		}
		slog.Info("metrics exported",
			"format", "prometheus",
			"filename", filename,
			"records", len(metrics),
			"file_size_bytes", buf.Len(),
		)
	}

	return buf.Bytes(), nil
}

func (e *Exporter) ExportJSONWithTimestamp(filter Filter) ([]byte, string, error) {
	filename := fmt.Sprintf("metrics_%s.json", time.Now().Format("20060102_150405"))
	data, err := e.ExportJSON(filter, filename)
	return data, filename, err
}

func (e *Exporter) ExportCSVWithTimestamp(filter Filter) ([]byte, string, error) {
	filename := fmt.Sprintf("metrics_%s.csv", time.Now().Format("20060102_150405"))
	data, err := e.ExportCSV(filter, filename)
	return data, filename, err
}

func (e *Exporter) ExportPrometheusWithTimestamp(filter Filter) ([]byte, string, error) {
	filename := fmt.Sprintf("metrics_%s.prom", time.Now().Format("20060102_150405"))
	data, err := e.ExportPrometheus(filter, filename)
	return data, filename, err
}

func (e *Exporter) applyFilter(filter Filter) []Metric {
	if filter == nil {
		return e.store.GetAll()
	}
	return e.store.Filter(filter)
}

func (e *Exporter) exportToFile(data []byte, filename string) error {
	dir := filename[:max(0, lastSlash(filename))]
	if dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("create dir: %w", err)
		}
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func lastSlash(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '/' || s[i] == '\\' {
			return i
		}
	}
	return -1
}
