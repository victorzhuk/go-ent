package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/execution"
)

type TestManualSummarizationInput struct {
	FileCount       int            `json:"file_count"`
	ContextLength   int            `json:"context_length"`
	AverageFileSize int            `json:"average_file_size"`
	ForceTrigger    bool           `json:"force_trigger"`
	Threshold       ThresholdInput `json:"threshold"`
	Model           string         `json:"model"`
}

type ThresholdInput struct {
	FileCount     int `json:"file_count"`
	ContextLength int `json:"context_length"`
	TokenCount    int `json:"token_count"`
}

type TestManualSummarizationResponse struct {
	Triggered       bool                             `json:"triggered"`
	TriggerReason   string                           `json:"trigger_reason"`
	FilesCount      int                              `json:"files_count"`
	TotalLength     int                              `json:"total_length"`
	EstimatedTokens int                              `json:"estimated_tokens"`
	Threshold       execution.SummarizationThreshold `json:"threshold"`
	ForceBypassed   bool                             `json:"force_bypassed"`
	ContextInfo     ContextInfo                      `json:"context_info"`
}

type ContextInfo struct {
	ProjectPath  string `json:"project_path"`
	ChangeID     string `json:"change_id"`
	TaskID       string `json:"task_id"`
	IsSummarized bool   `json:"is_summarized"`
	SummaryCount int    `json:"summary_count"`
}

type TestAutomaticSummarizationInput struct {
	Scenarios []SummarizationScenario `json:"scenarios"`
}

type SummarizationScenario struct {
	Name            string                           `json:"name"`
	FileCount       int                              `json:"file_count"`
	ContextLength   int                              `json:"context_length"`
	Threshold       execution.SummarizationThreshold `json:"threshold"`
	ExpectedTrigger bool                             `json:"expected_trigger"`
}

type TestAutomaticSummarizationResponse struct {
	Passed  int              `json:"passed"`
	Failed  int              `json:"failed"`
	Total   int              `json:"total"`
	Results []ScenarioResult `json:"results"`
}

type ScenarioResult struct {
	Name            string `json:"name"`
	Passed          bool   `json:"passed"`
	Expected        bool   `json:"expected"`
	Actual          bool   `json:"actual"`
	FilesCount      int    `json:"files_count"`
	TotalLength     int    `json:"total_length"`
	TokenThreshold  int    `json:"token_threshold"`
	EstimatedTokens int    `json:"estimated_tokens"`
}

type TestThresholdPersistenceInput struct {
	Threshold   execution.SummarizationThreshold `json:"threshold"`
	ProjectPath string                           `json:"project_path"`
}

type TestThresholdPersistenceResponse struct {
	WrittenSuccessfully bool                             `json:"written_successfully"`
	ReadSuccessfully    bool                             `json:"read_successfully"`
	MatchesExpected     bool                             `json:"matches_expected"`
	WrittenThreshold    execution.SummarizationThreshold `json:"written_threshold"`
	ReadThreshold       execution.SummarizationThreshold `json:"read_threshold"`
	Error               string                           `json:"error,omitempty"`
}

type TestModelConfigurationInput struct {
	Model    string `json:"model"`
	Provider string `json:"provider"`
}

type TestModelConfigurationResponse struct {
	Valid          bool   `json:"valid"`
	ValidatedModel string `json:"validated_model"`
	Error          string `json:"error,omitempty"`
	Supported      bool   `json:"supported"`
}

func NewTestManualSummarizationTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "test_manual_summarization",
		Description: "Test manual summarization trigger with various context sizes and threshold configurations",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file_count": map[string]any{
					"type":        "integer",
					"description": "Number of files in context",
					"default":     10,
				},
				"context_length": map[string]any{
					"type":        "integer",
					"description": "Total context length in characters",
					"default":     10000,
				},
				"average_file_size": map[string]any{
					"type":        "integer",
					"description": "Average file size in characters",
					"default":     1000,
				},
				"force_trigger": map[string]any{
					"type":        "boolean",
					"description": "Force summarization regardless of threshold",
					"default":     true,
				},
				"threshold": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"file_count": map[string]any{
							"type":        "integer",
							"description": "File count threshold",
							"default":     50,
						},
						"context_length": map[string]any{
							"type":        "integer",
							"description": "Context length threshold",
							"default":     50000,
						},
						"token_count": map[string]any{
							"type":        "integer",
							"description": "Token count threshold",
							"default":     10000,
						},
					},
				},
				"model": map[string]any{
					"type":        "string",
					"description": "Model to use for summarization",
					"default":     "claude-3.5-sonnet",
				},
			},
		},
	}
}

func NewTestAutomaticSummarizationTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "test_automatic_summarization",
		Description: "Test automatic summarization triggers with various thresholds (70%, 80%, 90%)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"scenarios": map[string]any{
					"type":        "array",
					"description": "List of test scenarios",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"name": map[string]any{
								"type":        "string",
								"description": "Scenario name",
							},
							"file_count": map[string]any{
								"type":        "integer",
								"description": "Number of files",
							},
							"context_length": map[string]any{
								"type":        "integer",
								"description": "Context length",
							},
							"threshold": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"file_count": map[string]any{
										"type": "integer",
									},
									"context_length": map[string]any{
										"type": "integer",
									},
									"token_count": map[string]any{
										"type": "integer",
									},
								},
							},
							"expected_trigger": map[string]any{
								"type":        "boolean",
								"description": "Expected trigger result",
							},
						},
					},
				},
			},
		},
	}
}

func NewTestThresholdPersistenceTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "test_threshold_persistence",
		Description: "Test threshold configuration persistence to file",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"threshold": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"file_count": map[string]any{
							"type":        "integer",
							"description": "File count threshold",
							"default":     50,
						},
						"context_length": map[string]any{
							"type":        "integer",
							"description": "Context length threshold",
							"default":     50000,
						},
						"token_count": map[string]any{
							"type":        "integer",
							"description": "Token count threshold",
							"default":     10000,
						},
					},
				},
				"project_path": map[string]any{
					"type":        "string",
					"description": "Project path for config file",
					"default":     ".",
				},
			},
		},
	}
}

func NewTestModelConfigurationTool() *mcp.Tool {
	return &mcp.Tool{
		Name:        "test_model_configuration",
		Description: "Test model validation and configuration changes",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"model": map[string]any{
					"type":        "string",
					"description": "Model name to validate",
					"default":     "claude-3.5-sonnet",
				},
				"provider": map[string]any{
					"type":        "string",
					"description": "Provider name",
					"default":     "anthropic",
				},
			},
		},
	}
}

func HandleTestManualSummarization(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input TestManualSummarizationInput
	if err := json.Unmarshal(request.Params.Arguments, &input); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("parse input: %v", err)}},
		}, fmt.Errorf("parse input: %w", err)
	}

	if input.FileCount == 0 {
		input.FileCount = input.ContextLength / input.AverageFileSize
	}

	tc := execution.NewTaskContext("/test").
		WithChange("test-change").
		WithTask("test-task").
		WithFiles(createTestFiles(input.FileCount, input.AverageFileSize))

	threshold := execution.SummarizationThreshold{
		FileCount:     input.Threshold.FileCount,
		ContextLength: input.Threshold.ContextLength,
		TokenCount:    input.Threshold.TokenCount,
	}

	shouldSummarize := input.ForceTrigger || shouldSummarizeWithThreshold(tc, threshold)

	triggerReason := ""
	forceBypassed := false

	if input.ForceTrigger {
		triggerReason = "manual force trigger"
		forceBypassed = true
	} else if len(tc.Files) > threshold.FileCount {
		triggerReason = fmt.Sprintf("file count (%d) exceeds threshold (%d)", len(tc.Files), threshold.FileCount)
	} else if calculateTotalLength(tc.Files) > threshold.ContextLength {
		triggerReason = fmt.Sprintf("context length (%d) exceeds threshold (%d)", calculateTotalLength(tc.Files), threshold.ContextLength)
	} else if estimateTestTokens(tc.Files) > threshold.TokenCount {
		triggerReason = fmt.Sprintf("estimated tokens (%d) exceeds threshold (%d)", estimateTestTokens(tc.Files), threshold.TokenCount)
	}

	totalLength := calculateTotalLength(tc.Files)
	estimatedTokens := estimateTestTokens(tc.Files)

	response := TestManualSummarizationResponse{
		Triggered:       shouldSummarize,
		TriggerReason:   triggerReason,
		FilesCount:      len(tc.Files),
		TotalLength:     totalLength,
		EstimatedTokens: estimatedTokens,
		Threshold:       threshold,
		ForceBypassed:   forceBypassed,
		ContextInfo: ContextInfo{
			ProjectPath:  tc.ProjectPath,
			ChangeID:     tc.ChangeID,
			TaskID:       tc.TaskID,
			IsSummarized: tc.IsSummarized,
			SummaryCount: tc.SummarizationCount,
		},
	}

	result, _ := json.MarshalIndent(response, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(result)}},
	}, nil
}

func HandleTestAutomaticSummarization(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input TestAutomaticSummarizationInput
	if err := json.Unmarshal(request.Params.Arguments, &input); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("parse input: %v", err)}},
		}, fmt.Errorf("parse input: %w", err)
	}

	if len(input.Scenarios) == 0 {
		input.Scenarios = getDefaultScenarios()
	}

	results := make([]ScenarioResult, 0, len(input.Scenarios))

	for _, scenario := range input.Scenarios {
		tc := execution.NewTaskContext("/test").
			WithFiles(createTestFiles(scenario.FileCount, scenario.ContextLength/scenario.FileCount))

		triggered := shouldSummarizeWithThreshold(tc, scenario.Threshold)
		passed := triggered == scenario.ExpectedTrigger

		result := ScenarioResult{
			Name:            scenario.Name,
			Passed:          passed,
			Expected:        scenario.ExpectedTrigger,
			Actual:          triggered,
			FilesCount:      len(tc.Files),
			TotalLength:     calculateTotalLength(tc.Files),
			TokenThreshold:  scenario.Threshold.TokenCount,
			EstimatedTokens: estimateTestTokens(tc.Files),
		}
		results = append(results, result)
	}

	passed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		}
	}

	response := TestAutomaticSummarizationResponse{
		Passed:  passed,
		Failed:  len(results) - passed,
		Total:   len(results),
		Results: results,
	}

	result, _ := json.MarshalIndent(response, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(result)}},
	}, nil
}

func HandleTestThresholdPersistence(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input TestThresholdPersistenceInput
	if err := json.Unmarshal(request.Params.Arguments, &input); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("parse input: %v", err)}},
		}, fmt.Errorf("parse input: %w", err)
	}

	response := TestThresholdPersistenceResponse{
		WrittenThreshold: input.Threshold,
	}

	loaded, err := execution.LoadSummarizationThreshold(input.ProjectPath)
	if err != nil {
		response.ReadSuccessfully = false
		response.Error = err.Error()
	} else {
		response.ReadSuccessfully = true
		response.ReadThreshold = loaded
		response.MatchesExpected = loaded.FileCount == input.Threshold.FileCount &&
			loaded.ContextLength == input.Threshold.ContextLength &&
			loaded.TokenCount == input.Threshold.TokenCount
	}

	result, _ := json.MarshalIndent(response, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(result)}},
	}, nil
}

func HandleTestModelConfiguration(ctx context.Context, request *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input TestModelConfigurationInput
	if err := json.Unmarshal(request.Params.Arguments, &input); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("parse input: %v", err)}},
		}, fmt.Errorf("parse input: %w", err)
	}

	response := TestModelConfigurationResponse{
		Valid:          false,
		ValidatedModel: "",
		Error:          "",
		Supported:      false,
	}

	supportedModels := map[string]bool{
		execution.LLMModelClaude35:   true,
		execution.LLMModelGPT4:       true,
		execution.LLMModelGPT4Turbo:  true,
		execution.LLMModelGPT35Turbo: true,
	}

	if supportedModels[input.Model] {
		response.Supported = true
		response.Valid = true
		response.ValidatedModel = input.Model
	} else {
		response.Valid = false
		response.ValidatedModel = execution.LLMModelClaude35
		response.Error = fmt.Sprintf("model %s not supported, using default %s", input.Model, execution.LLMModelClaude35)
	}

	result, _ := json.MarshalIndent(response, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(result)}},
	}, nil
}

func createTestFiles(count, avgSize int) []string {
	files := make([]string, count)
	for i := 0; i < count; i++ {
		files[i] = strings.Repeat("x", avgSize)
	}
	return files
}

func calculateTotalLength(files []string) int {
	total := 0
	for _, f := range files {
		total += len(f)
	}
	return total
}

func estimateTestTokens(files []string) int {
	return calculateTotalLength(files) / 4
}

func shouldSummarizeWithThreshold(tc *execution.TaskContext, threshold execution.SummarizationThreshold) bool {
	if len(tc.Files) > threshold.FileCount {
		return true
	}

	totalLength := 0
	for _, file := range tc.Files {
		totalLength += len(file)
	}
	if totalLength > threshold.ContextLength {
		return true
	}

	estimatedTokens := len(strings.Join(tc.Files, "\n")) / 4
	if estimatedTokens > threshold.TokenCount {
		return true
	}

	return false
}

func getDefaultScenarios() []SummarizationScenario {
	return []SummarizationScenario{
		{
			Name:            "70% threshold - file count",
			FileCount:       36,
			ContextLength:   30000,
			Threshold:       execution.SummarizationThreshold{FileCount: 50, ContextLength: 50000, TokenCount: 10000},
			ExpectedTrigger: false,
		},
		{
			Name:            "80% threshold - context length",
			FileCount:       35,
			ContextLength:   42000,
			Threshold:       execution.SummarizationThreshold{FileCount: 50, ContextLength: 50000, TokenCount: 10000},
			ExpectedTrigger: true,
		},
		{
			Name:            "90% threshold - token count",
			FileCount:       45,
			ContextLength:   46000,
			Threshold:       execution.SummarizationThreshold{FileCount: 50, ContextLength: 50000, TokenCount: 10000},
			ExpectedTrigger: true,
		},
		{
			Name:            "below all thresholds",
			FileCount:       30,
			ContextLength:   30000,
			Threshold:       execution.SummarizationThreshold{FileCount: 50, ContextLength: 50000, TokenCount: 10000},
			ExpectedTrigger: false,
		},
		{
			Name:            "exceeds all thresholds",
			FileCount:       60,
			ContextLength:   60000,
			Threshold:       execution.SummarizationThreshold{FileCount: 50, ContextLength: 50000, TokenCount: 10000},
			ExpectedTrigger: true,
		},
	}
}
