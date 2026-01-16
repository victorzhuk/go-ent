package rules

import (
	"context"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type RuleDefinition struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Enabled     bool        `yaml:"enabled"`
	Conditions  []Condition `yaml:"conditions"`
	Actions     []ActionDef `yaml:"actions"`
}

type Condition struct {
	Type     string                 `yaml:"type"`
	Field    string                 `yaml:"field,omitempty"`
	Operator string                 `yaml:"operator"`
	Value    interface{}            `yaml:"value,omitempty"`
	Params   map[string]interface{} `yaml:"params,omitempty"`
}

type ActionDef struct {
	Type    string                 `yaml:"type"`
	Params  map[string]interface{} `yaml:"params"`
	Comment string                 `yaml:"comment,omitempty"`
}

func LoadRuleDefinition(path string) (*RuleDefinition, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- controlled config/template file path
	if err != nil {
		return nil, fmt.Errorf("read rule: %w", err)
	}

	var def RuleDefinition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("parse rule: %w", err)
	}

	if err := def.Validate(); err != nil {
		return nil, fmt.Errorf("validate rule: %w", err)
	}

	return &def, nil
}

func (d *RuleDefinition) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("name is required")
	}

	if d.Description == "" {
		return fmt.Errorf("description is required")
	}

	if len(d.Conditions) == 0 {
		return fmt.Errorf("at least one condition is required")
	}

	if len(d.Actions) == 0 {
		return fmt.Errorf("at least one action is required")
	}

	for i, cond := range d.Conditions {
		if err := cond.Validate(); err != nil {
			return fmt.Errorf("condition %d: %w", i, err)
		}
	}

	for i, act := range d.Actions {
		if err := act.Validate(); err != nil {
			return fmt.Errorf("action %d: %w", i, err)
		}
	}

	return nil
}

func (c *Condition) Validate() error {
	if c.Type == "" {
		return fmt.Errorf("type is required")
	}

	if c.Operator == "" {
		return fmt.Errorf("operator is required")
	}

	validOperators := map[string]bool{
		"eq": true, "ne": true, "gt": true, "lt": true,
		"gte": true, "lte": true, "contains": true,
		"not_contains": true, "regex": true, "in": true,
	}

	if !validOperators[c.Operator] {
		return fmt.Errorf("invalid operator: %s", c.Operator)
	}

	return nil
}

func (a *ActionDef) Validate() error {
	if a.Type == "" {
		return fmt.Errorf("type is required")
	}

	return nil
}

func (d *RuleDefinition) CreateRule() Rule {
	return &YAMLRule{definition: d}
}

type YAMLRule struct {
	definition *RuleDefinition
}

func (r *YAMLRule) Name() string {
	return r.definition.Name
}

func (r *YAMLRule) Description() string {
	return r.definition.Description
}

func (r *YAMLRule) CanHandle(event Event) bool {
	if !r.definition.Enabled {
		return false
	}

	for _, condition := range r.definition.Conditions {
		if !matchesCondition(event, condition) {
			return false
		}
	}

	return true
}

func (r *YAMLRule) Evaluate(ctx context.Context, event Event) ([]Action, error) {
	actions := make([]Action, 0, len(r.definition.Actions))

	for _, actionDef := range r.definition.Actions {
		action := Action(actionDef)
		actions = append(actions, action)
	}

	return actions, nil
}

func matchesCondition(event Event, cond Condition) bool {
	if cond.Type == "event_type" {
		if cond.Operator == "eq" {
			if val, ok := cond.Value.(string); ok && val == event.Type {
				return true
			}
		}
	}

	if cond.Type == "source" {
		if cond.Operator == "eq" {
			if val, ok := cond.Value.(string); ok && val == event.Source {
				return true
			}
		}
	}

	return false
}
