package router

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadRoutingConfig_MissingFile(t *testing.T) {
	t.Parallel()

	config, err := LoadRoutingConfig("nonexistent.yaml")
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 7, len(config.Rules))
}

func TestLoadRoutingConfig_EmptyFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "routing.yaml")

	err := os.WriteFile(configPath, []byte(""), 0644)
	require.NoError(t, err)

	config, err := LoadRoutingConfig(configPath)
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 7, len(config.Rules))
}

func TestLoadRoutingConfig_ValidYAML(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "routing.yaml")

	content := `
rules:
  - id: "rule1"
    priority: 100
    match:
      complexity: "trivial"
      file_count: 1
    action:
      method: "cli"
      provider: "anthropic"
      model: "claude-3-haiku-3-5"

  - id: "rule2"
    priority: 200
    match:
      type:
        - "implement"
        - "feature"
      context_size: 50000
    action:
      method: "acp"
      provider: "moonshot"
      model: "glm-4"
`
	err := os.WriteFile(configPath, []byte(content), 0644)
	require.NoError(t, err)

	config, err := LoadRoutingConfig(configPath)
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 2, len(config.Rules))
	assert.Equal(t, "rule1", config.Rules[0].ID)
	assert.Equal(t, "rule2", config.Rules[1].ID)
	assert.Equal(t, "trivial", config.Rules[0].Match.Complexity)
	assert.Equal(t, intPtr(1), config.Rules[0].Match.FileCount)
	assert.Equal(t, "cli", config.Rules[0].Action.Method)
	assert.Equal(t, "anthropic", config.Rules[0].Action.Provider)
}

func TestLoadRoutingConfig_InvalidYAML(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "routing.yaml")

	invalidContent := `
rules:
  - id: "test"
    priority: [invalid yaml]
`
	err := os.WriteFile(configPath, []byte(invalidContent), 0644)
	require.NoError(t, err)

	config, err := LoadRoutingConfig(configPath)
	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestLoadRoutingConfig_ValidationErrors(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	tests := []struct {
		name     string
		content  string
		errorMsg string
	}{
		{
			name: "duplicate rule id",
			content: `
rules:
  - id: "duplicate"
    priority: 100
    match:
      complexity: "trivial"
    action:
      method: "cli"
      provider: "anthropic"
  - id: "duplicate"
    priority: 200
    match:
      complexity: "simple"
    action:
      method: "acp"
      provider: "moonshot"
`,
			errorMsg: "duplicate id",
		},
		{
			name: "invalid method",
			content: `
rules:
  - id: "test"
    priority: 100
    match:
      complexity: "trivial"
    action:
      method: "invalid-method"
      provider: "anthropic"
`,
			errorMsg: "invalid method",
		},
		{
			name: "missing provider",
			content: `
rules:
  - id: "test"
    priority: 100
    match:
      complexity: "trivial"
    action:
      method: "cli"
`,
			errorMsg: "provider is required",
		},
		{
			name: "invalid type",
			content: `
rules:
  - id: "test"
    priority: 100
    match:
      type:
        - "invalid-type"
    action:
      method: "cli"
      provider: "anthropic"
`,
			errorMsg: "invalid type",
		},
		{
			name: "invalid complexity",
			content: `
rules:
  - id: "test"
    priority: 100
    match:
      complexity: "invalid-complexity"
    action:
      method: "cli"
      provider: "anthropic"
`,
			errorMsg: "invalid complexity",
		},
		{
			name: "negative file_count",
			content: `
rules:
  - id: "test"
    priority: 100
    match:
      file_count: -1
    action:
      method: "cli"
      provider: "anthropic"
`,
			errorMsg: "file_count must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testPath := filepath.Join(tempDir, tt.name+".yaml")
			err := os.WriteFile(testPath, []byte(tt.content), 0644)
			require.NoError(t, err)

			config, err := LoadRoutingConfig(testPath)
			assert.Error(t, err)
			assert.Nil(t, config)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

func TestRoutingConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  RoutingConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: RoutingConfig{
				Rules: []RoutingRule{
					{
						ID:       "rule1",
						Priority: 100,
						Match: MatchConditions{
							Complexity: "simple",
						},
						Action: RouteAction{
							Method:   "cli",
							Provider: "anthropic",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing rule id",
			config: RoutingConfig{
				Rules: []RoutingRule{
					{
						Priority: 100,
						Match: MatchConditions{
							Complexity: "simple",
						},
						Action: RouteAction{
							Method:   "cli",
							Provider: "anthropic",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate ids",
			config: RoutingConfig{
				Rules: []RoutingRule{
					{
						ID:       "duplicate",
						Priority: 100,
						Match:    MatchConditions{},
						Action: RouteAction{
							Method:   "cli",
							Provider: "anthropic",
						},
					},
					{
						ID:       "duplicate",
						Priority: 200,
						Match:    MatchConditions{},
						Action: RouteAction{
							Method:   "acp",
							Provider: "moonshot",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name:    "empty config",
			config:  RoutingConfig{Rules: []RoutingRule{}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRoutingRule_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rule    RoutingRule
		wantErr bool
	}{
		{
			name: "valid rule",
			rule: RoutingRule{
				ID:       "test-rule",
				Priority: 100,
				Match: MatchConditions{
					Type:       []string{"implement", "feature"},
					Complexity: "simple",
					FileCount:  intPtr(3),
				},
				Action: RouteAction{
					Method:   "acp",
					Provider: "moonshot",
					Model:    "glm-4",
				},
			},
			wantErr: false,
		},
		{
			name: "negative priority",
			rule: RoutingRule{
				ID:       "test",
				Priority: -1,
				Match:    MatchConditions{},
				Action: RouteAction{
					Method:   "cli",
					Provider: "anthropic",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid action method",
			rule: RoutingRule{
				ID:       "test",
				Priority: 100,
				Match:    MatchConditions{},
				Action: RouteAction{
					Method:   "invalid",
					Provider: "anthropic",
				},
			},
			wantErr: true,
		},
		{
			name: "missing provider",
			rule: RoutingRule{
				ID:       "test",
				Priority: 100,
				Match:    MatchConditions{},
				Action: RouteAction{
					Method: "cli",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			rule: RoutingRule{
				ID:       "test",
				Priority: 100,
				Match: MatchConditions{
					Type: []string{"invalid-type"},
				},
				Action: RouteAction{
					Method:   "cli",
					Provider: "anthropic",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid complexity",
			rule: RoutingRule{
				ID:       "test",
				Priority: 100,
				Match: MatchConditions{
					Complexity: "invalid",
				},
				Action: RouteAction{
					Method:   "cli",
					Provider: "anthropic",
				},
			},
			wantErr: true,
		},
		{
			name: "negative file_count",
			rule: RoutingRule{
				ID:       "test",
				Priority: 100,
				Match: MatchConditions{
					FileCount: intPtr(-1),
				},
				Action: RouteAction{
					Method:   "cli",
					Provider: "anthropic",
				},
			},
			wantErr: true,
		},
		{
			name: "negative context_size",
			rule: RoutingRule{
				ID:       "test",
				Priority: 100,
				Match: MatchConditions{
					ContextSize: intPtr(-1),
				},
				Action: RouteAction{
					Method:   "cli",
					Provider: "anthropic",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.rule.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRouteAction_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		action  RouteAction
		wantErr bool
	}{
		{
			name: "valid action",
			action: RouteAction{
				Method:   "cli",
				Provider: "anthropic",
				Model:    "claude-3-haiku-3-5",
			},
			wantErr: false,
		},
		{
			name: "valid action without model",
			action: RouteAction{
				Method:   "acp",
				Provider: "moonshot",
			},
			wantErr: false,
		},
		{
			name: "invalid method",
			action: RouteAction{
				Method:   "invalid",
				Provider: "anthropic",
			},
			wantErr: true,
		},
		{
			name: "missing provider",
			action: RouteAction{
				Method: "cli",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.action.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMatchKeywords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		rule        RoutingRule
		description string
		wantMatch   bool
	}{
		{
			name: "matches keyword",
			rule: RoutingRule{
				Match: MatchConditions{
					Keywords: []string{"urgent", "critical"},
				},
			},
			description: "URGENT: Fix critical bug",
			wantMatch:   true,
		},
		{
			name: "no keywords specified",
			rule: RoutingRule{
				Match: MatchConditions{},
			},
			description: "Any description",
			wantMatch:   true,
		},
		{
			name: "no keyword match",
			rule: RoutingRule{
				Match: MatchConditions{
					Keywords: []string{"urgent", "critical"},
				},
			},
			description: "Add new feature",
			wantMatch:   false,
		},
		{
			name: "case insensitive match",
			rule: RoutingRule{
				Match: MatchConditions{
					Keywords: []string{"URGENT"},
				},
			},
			description: "urgent: fix bug",
			wantMatch:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := tt.rule.MatchKeywords(tt.description)
			assert.Equal(t, tt.wantMatch, result)
		})
	}
}

func TestSortRules(t *testing.T) {
	t.Parallel()

	config := RoutingConfig{
		Rules: []RoutingRule{
			{ID: "low", Priority: 10},
			{ID: "high", Priority: 1000},
			{ID: "medium", Priority: 100},
			{ID: "default", Priority: 0},
		},
	}

	config.SortRules()

	assert.Equal(t, "high", config.Rules[0].ID)
	assert.Equal(t, "medium", config.Rules[1].ID)
	assert.Equal(t, "low", config.Rules[2].ID)
	assert.Equal(t, "default", config.Rules[3].ID)
}
