package router

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/victorzhuk/go-ent/internal/config"
)

const (
	defaultRoutingFile = ".goent/routing.yaml"
	defaultPriority    = 100
)

type RoutingConfig struct {
	Rules []RoutingRule `yaml:"rules,omitempty"`
}

type RoutingRule struct {
	ID       string          `yaml:"id"`
	Priority int             `yaml:"priority"`
	Match    MatchConditions `yaml:"match"`
	Action   RouteAction     `yaml:"action"`
}

type MatchConditions struct {
	Type        []string `yaml:"type,omitempty"`
	Complexity  string   `yaml:"complexity,omitempty"`
	FileCount   *int     `yaml:"file_count,omitempty"`
	ContextSize *int     `yaml:"context_size,omitempty"`
	Keywords    []string `yaml:"keywords,omitempty"`
	MaxCost     *float64 `yaml:"max_cost,omitempty"`
}

type RouteAction struct {
	Method   string `yaml:"method"`
	Provider string `yaml:"provider"`
	Model    string `yaml:"model,omitempty"`
}

func LoadRoutingConfig(path string) (*RoutingConfig, error) {
	if path == "" {
		path = defaultRoutingFile
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &RoutingConfig{
				Rules: DefaultRoutingRules(),
			}, nil
		}
		return nil, fmt.Errorf("read routing config: %w", err)
	}

	var config RoutingConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse routing config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validate routing config: %w", err)
	}

	if len(config.Rules) == 0 {
		config.Rules = DefaultRoutingRules()
	}

	return &config, nil
}

func (c *RoutingConfig) Validate() error {
	seen := make(map[string]bool)

	for i, rule := range c.Rules {
		if rule.ID == "" {
			return fmt.Errorf("rule at index %d: id is required", i)
		}

		if seen[rule.ID] {
			return fmt.Errorf("rule %s: duplicate id", rule.ID)
		}
		seen[rule.ID] = true

		if err := rule.Validate(); err != nil {
			return fmt.Errorf("rule %s: %w", rule.ID, err)
		}
	}

	return nil
}

func (r *RoutingRule) Validate() error {
	if r.Priority < 0 {
		return fmt.Errorf("priority must be non-negative")
	}

	if err := r.Action.Validate(); err != nil {
		return fmt.Errorf("action: %w", err)
	}

	validTypes := []string{"implement", "refactor", "analyze", "fix", "test", "feature", "bugfix", "documentation"}
	for _, t := range r.Match.Type {
		if !slices.Contains(validTypes, t) {
			return fmt.Errorf("invalid type %q, must be one of: %s", t, strings.Join(validTypes, ", "))
		}
	}

	validComplexity := []string{"trivial", "simple", "medium", "complex", "high"}
	if r.Match.Complexity != "" && !slices.Contains(validComplexity, r.Match.Complexity) {
		return fmt.Errorf("invalid complexity %q, must be one of: %s", r.Match.Complexity, strings.Join(validComplexity, ", "))
	}

	if r.Match.FileCount != nil && *r.Match.FileCount < 0 {
		return fmt.Errorf("file_count must be non-negative")
	}

	if r.Match.ContextSize != nil && *r.Match.ContextSize < 0 {
		return fmt.Errorf("context_size must be non-negative")
	}

	if r.Match.MaxCost != nil && *r.Match.MaxCost < 0 {
		return fmt.Errorf("max_cost must be non-negative")
	}

	return nil
}

func (a *RouteAction) Validate() error {
	method := config.CommunicationMethod(a.Method)
	if !method.Valid() {
		return fmt.Errorf("invalid method %q, must be one of: acp, cli, api", a.Method)
	}

	if a.Provider == "" {
		return fmt.Errorf("provider is required")
	}

	return nil
}

func DefaultRoutingRules() []RoutingRule {
	return []RoutingRule{
		{
			ID:       "simple-tasks-cli-haiku",
			Priority: 1000,
			Match: MatchConditions{
				Complexity:  "trivial",
				FileCount:   intPtr(0),
				ContextSize: intPtr(20000),
			},
			Action: RouteAction{
				Method:   "cli",
				Provider: "anthropic",
				Model:    "claude-3-haiku",
			},
		},
		{
			ID:       "bulk-implementation-acp-glm",
			Priority: 900,
			Match: MatchConditions{
				Type:      []string{"implement", "feature"},
				FileCount: intPtr(3),
			},
			Action: RouteAction{
				Method:   "acp",
				Provider: "moonshot",
				Model:    "glm-4",
			},
		},
		{
			ID:       "large-context-acp-kimi",
			Priority: 950,
			Match: MatchConditions{
				ContextSize: intPtr(50000),
			},
			Action: RouteAction{
				Method:   "acp",
				Provider: "moonshot",
				Model:    "kimi-k2",
			},
		},
		{
			ID:       "complex-refactor-acp-deepseek",
			Priority: 980,
			Match: MatchConditions{
				Type:       []string{"refactor"},
				Complexity: "complex",
			},
			Action: RouteAction{
				Method:   "acp",
				Provider: "deepseek",
				Model:    "deepseek-coder",
			},
		},
		{
			ID:       "test-generation-cli-haiku",
			Priority: 970,
			Match: MatchConditions{
				Type:       []string{"test"},
				Complexity: "simple",
			},
			Action: RouteAction{
				Method:   "cli",
				Provider: "anthropic",
				Model:    "claude-3-haiku",
			},
		},
		{
			ID:       "bugfix-simple-cli",
			Priority: 960,
			Match: MatchConditions{
				Type:       []string{"fix", "bugfix"},
				Complexity: "simple",
				FileCount:  intPtr(2),
			},
			Action: RouteAction{
				Method:   "cli",
				Provider: "anthropic",
				Model:    "claude-3-5-sonnet",
			},
		},
		{
			ID:       "fallback-default",
			Priority: 0,
			Match:    MatchConditions{},
			Action: RouteAction{
				Method:   "cli",
				Provider: "moonshot",
				Model:    "glm-4",
			},
		},
	}
}

func intPtr(i int) *int {
	return &i
}

func (c *RoutingConfig) SortRules() {
	slices.SortFunc(c.Rules, func(a, b RoutingRule) int {
		return b.Priority - a.Priority
	})
}

func (r *RoutingRule) MatchKeywords(description string) bool {
	if len(r.Match.Keywords) == 0 {
		return true
	}

	descLower := strings.ToLower(description)
	for _, kw := range r.Match.Keywords {
		if strings.Contains(descLower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}
