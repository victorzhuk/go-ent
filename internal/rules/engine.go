package rules

import (
	"context"
	"fmt"
	"sync"
)

type Engine struct {
	rules   []Rule
	enabled map[string]bool
	mu      sync.RWMutex
}

func NewEngine() *Engine {
	return &Engine{
		rules:   make([]Rule, 0),
		enabled: make(map[string]bool),
	}
}

func (e *Engine) LoadRule(r Rule) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, existing := range e.rules {
		if existing.Name() == r.Name() {
			return fmt.Errorf("rule %s already loaded", r.Name())
		}
	}

	e.rules = append(e.rules, r)
	e.enabled[r.Name()] = true

	return nil
}

func (e *Engine) UnloadRule(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for i, r := range e.rules {
		if r.Name() == name {
			e.rules = append(e.rules[:i], e.rules[i+1:]...)
			delete(e.enabled, name)
			return nil
		}
	}

	return fmt.Errorf("rule %s not found", name)
}

func (e *Engine) Enable(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, r := range e.rules {
		if r.Name() == name {
			e.enabled[name] = true
			return nil
		}
	}

	return fmt.Errorf("rule %s not found", name)
}

func (e *Engine) Disable(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.enabled, name)
	return nil
}

func (e *Engine) Evaluate(ctx context.Context, event Event) ([]Action, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var actions []Action

	for _, rule := range e.rules {
		if !e.enabled[rule.Name()] {
			continue
		}

		if rule.CanHandle(event) {
			ruleActions, err := rule.Evaluate(ctx, event)
			if err != nil {
				return nil, fmt.Errorf("evaluate rule %s: %w", rule.Name(), err)
			}

			if len(ruleActions) > 0 {
				actions = append(actions, ruleActions...)
			}
		}
	}

	return actions, nil
}

func (e *Engine) List() []RuleInfo {
	e.mu.RLock()
	defer e.mu.RUnlock()

	infos := make([]RuleInfo, 0, len(e.rules))
	for _, r := range e.rules {
		infos = append(infos, RuleInfo{
			Name:        r.Name(),
			Description: r.Description(),
			Enabled:     e.enabled[r.Name()],
		})
	}

	return infos
}

type Event struct {
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Context   map[string]interface{} `json:"context"`
	Timestamp int64                  `json:"timestamp"`
}

type Action struct {
	Type    string                 `json:"type"`
	Params  map[string]interface{} `json:"params"`
	Comment string                 `json:"comment,omitempty"`
}

type Rule interface {
	Name() string
	Description() string
	CanHandle(event Event) bool
	Evaluate(ctx context.Context, event Event) ([]Action, error)
}

type RuleInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}
