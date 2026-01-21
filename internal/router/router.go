package router

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/execution"
	"github.com/victorzhuk/go-ent/internal/memory"
	"github.com/victorzhuk/go-ent/internal/worker"
)

var (
	// ErrNoMatchingRule indicates no routing rule matched the task.
	ErrNoMatchingRule = fmt.Errorf("no matching routing rule")

	// ErrInvalidTask indicates the task is invalid for routing.
	ErrInvalidTask = fmt.Errorf("invalid task for routing")

	// ErrProviderNotFound indicates the requested provider is not available.
	ErrProviderNotFound = fmt.Errorf("provider not found")

	// ErrInvalidOverride indicates the override configuration is invalid.
	ErrInvalidOverride = fmt.Errorf("invalid override configuration")

	// ErrInvalidMethod indicates the requested method is not available.
	ErrInvalidMethod = fmt.Errorf("invalid method")

	// ErrInvalidModel indicates the requested model is not available.
	ErrInvalidModel = fmt.Errorf("invalid model")
)

type DefaultRoutes struct {
	// SimpleTasks provider for quick one-shot tasks.
	SimpleTasks string

	// Implementation provider for bulk implementation.
	Implementation string

	// LargeContext provider for files with large context.
	LargeContext string

	// ComplexTasks provider for complex refactoring.
	ComplexTasks string

	// Research provider stays in Claude Code.
	Research string

	// Planning provider stays in Claude Code.
	Planning string

	// Review provider stays in Claude Code.
	Review string
}

type RoutingDecision struct {
	// Method is the selected communication method.
	Method config.CommunicationMethod

	// Provider is the selected provider name.
	Provider string

	// Model is the selected model.
	Model string

	// Reason explains why this routing was chosen.
	Reason string

	// EstimatedCost is the predicted cost in USD.
	EstimatedCost float64

	// RuleName indicates which rule produced this decision.
	RuleName string
}

type Router struct {
	rules           []RoutingRule
	defaults        DefaultRoutes
	providers       map[string]worker.ProviderDefinition
	costBudget      float64
	remainingBudget float64
	costsByProvider map[string]float64
	memory          *memory.MemoryStore
	logger          *slog.Logger
	learningEnabled bool
}

func NewRouter(config *worker.Config, memoryStore *memory.MemoryStore, opts ...RouterOption) (*Router, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	router := &Router{
		rules:           make([]RoutingRule, 0),
		providers:       config.Providers,
		costBudget:      0,
		remainingBudget: 0,
		costsByProvider: make(map[string]float64),
		memory:          memoryStore,
		logger:          slog.Default(),
		learningEnabled: memoryStore != nil,
	}

	for _, opt := range opts {
		opt(router)
	}

	if config.CostTracking != nil && config.CostTracking.Enabled {
		router.SetCostBudget(config.CostTracking.GlobalBudget)
	}

	return router, nil
}

type RouterOption func(*Router)

func WithRoutingConfig(path string) RouterOption {
	return func(r *Router) {
		config, err := LoadRoutingConfig(path)
		if err != nil {
			r.logger.Warn("load routing config failed, using defaults", "error", err, "path", path)
			return
		}

		r.SetRules(config.Rules)
		r.logger.Info("loaded routing config", "path", path, "rules", len(config.Rules))
	}
}

func (r *Router) SetRules(rules []RoutingRule) {
	slices.SortFunc(rules, func(a, b RoutingRule) int {
		return b.Priority - a.Priority
	})
	r.rules = rules
}

func (r *Router) SetDefaults(defaults DefaultRoutes) {
	r.defaults = defaults
}

func (r *Router) SetCostBudget(budget float64) {
	r.costBudget = budget
	r.remainingBudget = budget
}

func (r *Router) Route(ctx context.Context, task *execution.Task) (*RoutingDecision, error) {
	if task == nil {
		return nil, fmt.Errorf("route: %w", ErrInvalidTask)
	}

	var decision *RoutingDecision
	var err error

	if task.ForceProvider != "" || task.ForceModel != "" || task.ForceMethod != "" || task.ForceAgent != "" {
		decision, err = r.routeOverride(ctx, task)
	} else if r.learningEnabled && r.memory != nil {
		decision, err = r.routeWithLearning(ctx, task)
		if decision == nil || err != nil {
			decision, err = r.applyRoutingRules(ctx, task)
		}
	} else {
		decision, err = r.applyRoutingRules(ctx, task)
	}

	if err != nil {
		return nil, err
	}

	if err := r.checkBudget(decision); err != nil {
		return nil, fmt.Errorf("budget check: %w", err)
	}

	r.logRoutingDecision(task, decision)

	return decision, nil
}

func (r *Router) applyRoutingRules(ctx context.Context, task *execution.Task) (*RoutingDecision, error) {
	taskCharacteristics := r.analyzeTask(task)

	for _, rule := range r.rules {
		if r.ruleMatches(rule, task, taskCharacteristics) {
			return r.createDecision(rule, task, taskCharacteristics), nil
		}
	}

	return r.applyDefaultRouting(ctx, task, taskCharacteristics)
}

func (r *Router) routeOverride(ctx context.Context, task *execution.Task) (*RoutingDecision, error) {
	decision := &RoutingDecision{
		RuleName: "override",
	}

	chars := r.analyzeTask(task)

	if task.ForceProvider != "" {
		providerDef, exists := r.providers[task.ForceProvider]
		if !exists {
			return nil, fmt.Errorf("override provider %s not found: %w", task.ForceProvider, ErrProviderNotFound)
		}
		decision.Provider = task.ForceProvider
		decision.Method = providerDef.Method

		if task.ForceModel == "" {
			decision.Model = providerDef.Model
		}

		if task.ForceMethod == "" {
			decision.Method = providerDef.Method
		}
	} else if task.ForceAgent != "" {
		providerDef, exists := r.providers["glm"]
		if !exists {
			providerDef, exists = r.providers["kimi"]
		}
		if !exists {
			return nil, fmt.Errorf("no provider available for agent override: %w", ErrProviderNotFound)
		}
		decision.Provider = providerDef.Provider
		decision.Method = providerDef.Method
		decision.Model = providerDef.Model
	}

	if task.ForceModel != "" {
		if decision.Provider == "" {
			for name, provider := range r.providers {
				if strings.Contains(provider.Model, task.ForceModel) || provider.Model == task.ForceModel {
					decision.Provider = name
					decision.Method = provider.Method
					decision.Model = task.ForceModel
					break
				}
			}
			if decision.Provider == "" {
				return nil, fmt.Errorf("override model %s not found in any provider: %w", task.ForceModel, ErrInvalidModel)
			}
		} else {
			providerDef, exists := r.providers[decision.Provider]
			if !exists {
				return nil, fmt.Errorf("provider %s not found: %w", decision.Provider, ErrProviderNotFound)
			}

			modelExists := providerDef.Model == task.ForceModel || strings.Contains(providerDef.Model, task.ForceModel)
			if !modelExists {
				return nil, fmt.Errorf("model %s not available for provider %s: %w", task.ForceModel, decision.Provider, ErrInvalidModel)
			}
			decision.Model = task.ForceModel
		}
	}

	if task.ForceMethod != "" {
		method := config.CommunicationMethod(task.ForceMethod)
		if method != config.MethodACP && method != config.MethodCLI && method != config.MethodAPI {
			return nil, fmt.Errorf("invalid method %s: %w", task.ForceMethod, ErrInvalidMethod)
		}
		decision.Method = method
	}

	estimatedCost := 0.01
	if decision.Provider != "" {
		if providerDef, exists := r.providers[decision.Provider]; exists {
			estimatedCost = r.estimateCost(providerDef, chars, 1.0)
		}
	}

	decision.EstimatedCost = estimatedCost

	overrideReasons := []string{}
	if task.ForceProvider != "" {
		overrideReasons = append(overrideReasons, fmt.Sprintf("provider=%s", task.ForceProvider))
	}
	if task.ForceModel != "" {
		overrideReasons = append(overrideReasons, fmt.Sprintf("model=%s", task.ForceModel))
	}
	if task.ForceMethod != "" {
		overrideReasons = append(overrideReasons, fmt.Sprintf("method=%s", task.ForceMethod))
	}
	if task.ForceAgent != "" {
		overrideReasons = append(overrideReasons, fmt.Sprintf("agent=%s", task.ForceAgent))
	}

	decision.Reason = "manual override: " + strings.Join(overrideReasons, ", ")

	r.logger.Info("using manual override", "reasons", strings.Join(overrideReasons, ", "))

	return decision, nil
}

func (r *Router) routeWithLearning(ctx context.Context, task *execution.Task) (*RoutingDecision, error) {
	chars := r.analyzeTask(task)

	if r.memory == nil || r.memory.GetTotalPatterns() < 5 {
		return nil, fmt.Errorf("insufficient pattern data")
	}

	recommendation := r.memory.Query(task.Type, chars.fileCount, chars.estimatedTokens)
	if recommendation == nil {
		return nil, fmt.Errorf("no matching pattern")
	}

	_, exists := r.providers[recommendation.Provider]
	if !exists {
		return nil, fmt.Errorf("recommended provider %s not found: %w", recommendation.Provider, ErrProviderNotFound)
	}

	decision := &RoutingDecision{
		Method:        config.CommunicationMethod(recommendation.Method),
		Provider:      recommendation.Provider,
		Model:         recommendation.Model,
		Reason:        fmt.Sprintf("learned pattern: %s (confidence: %.1f%%)", recommendation.Reason, recommendation.Confidence*100),
		EstimatedCost: recommendation.EstimatedCost,
		RuleName:      "learned_pattern",
	}

	r.logger.Info("using learned pattern",
		"task_type", task.Type,
		"provider", decision.Provider,
		"model", decision.Model,
		"confidence", fmt.Sprintf("%.1f%%", recommendation.Confidence*100),
		"reason", recommendation.Reason,
	)

	return decision, nil
}

func (r *Router) ruleMatches(rule RoutingRule, task *execution.Task, chars taskCharacteristics) bool {
	if rule.Match.Complexity != "" && rule.Match.Complexity != chars.complexity {
		return false
	}

	if len(rule.Match.Type) > 0 {
		matches := false
		for _, tt := range rule.Match.Type {
			if strings.EqualFold(tt, task.Type) {
				matches = true
				break
			}
		}
		if !matches {
			return false
		}
	}

	if rule.Match.FileCount != nil && chars.fileCount < *rule.Match.FileCount {
		return false
	}

	if rule.Match.ContextSize != nil && chars.estimatedTokens < *rule.Match.ContextSize {
		return false
	}

	if !rule.MatchKeywords(task.Description) {
		return false
	}

	return true
}

func (r *Router) createDecision(rule RoutingRule, task *execution.Task, chars taskCharacteristics) *RoutingDecision {
	providerDef := r.providers[rule.Action.Provider]
	model := rule.Action.Model
	if model == "" {
		model = providerDef.Model
	}

	method := config.CommunicationMethod(rule.Action.Method)

	estimatedCost := r.estimateCost(providerDef, chars, 1.0)

	decision := &RoutingDecision{
		Method:        method,
		Provider:      rule.Action.Provider,
		Model:         model,
		Reason:        fmt.Sprintf("matched rule %s: %s", rule.ID, r.buildReason(rule, chars)),
		EstimatedCost: estimatedCost,
		RuleName:      rule.ID,
	}

	return decision
}

func (r *Router) applyDefaultRouting(ctx context.Context, task *execution.Task, _ taskCharacteristics) (*RoutingDecision, error) {
	chars := r.analyzeTask(task)

	var providerName string
	var reason string

	switch {
	case chars.complexity == "simple" && chars.fileCount <= 1:
		providerName = r.defaults.SimpleTasks
		reason = "simple task with single file, using simple_tasks default"

	case chars.estimatedTokens > 50000:
		providerName = r.defaults.LargeContext
		reason = fmt.Sprintf("large context (%d tokens), using large_context default", chars.estimatedTokens)

	case chars.complexity == "complex":
		providerName = r.defaults.ComplexTasks
		reason = "complex task, using complex_tasks default"

	case chars.fileCount > 3:
		providerName = r.defaults.Implementation
		reason = fmt.Sprintf("bulk implementation (%d files), using implementation default", chars.fileCount)

	default:
		providerName = r.defaults.Implementation
		reason = "no specific rule matched, using implementation default"
	}

	if providerName == "" {
		providerName = "glm"
		reason = "no default configured, using glm fallback"
	}

	providerDef, exists := r.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("default provider %s not found: %w", providerName, ErrProviderNotFound)
	}

	decision := &RoutingDecision{
		Method:        providerDef.Method,
		Provider:      providerName,
		Model:         providerDef.Model,
		Reason:        reason,
		EstimatedCost: r.estimateCost(providerDef, chars, 1.0),
		RuleName:      "default",
	}

	return decision, nil
}

type taskCharacteristics struct {
	complexity      string
	fileCount       int
	estimatedTokens int
	hasFiles        bool
}

func (r *Router) analyzeTask(task *execution.Task) taskCharacteristics {
	chars := taskCharacteristics{}

	if task.Context != nil {
		chars.fileCount = len(task.Context.Files)
		chars.hasFiles = chars.fileCount > 0
	}

	chars.estimatedTokens = r.estimateTokens(task)
	chars.complexity = r.assessComplexity(task, chars)

	return chars
}

func (r *Router) assessComplexity(task *execution.Task, chars taskCharacteristics) string {
	if task.ForceModel != "" || task.ForceAgent != "" {
		return "forced"
	}

	complexityScore := 0

	if task.Type == "refactor" || task.Type == "architecture" {
		complexityScore += 3
	}

	if task.Type == "feature" {
		complexityScore += 2
	}

	if chars.fileCount > 5 {
		complexityScore += 2
	} else if chars.fileCount > 2 {
		complexityScore += 1
	}

	if chars.estimatedTokens > 100000 {
		complexityScore += 3
	} else if chars.estimatedTokens > 50000 {
		complexityScore += 2
	} else if chars.estimatedTokens > 20000 {
		complexityScore += 1
	}

	if len(task.Skills) > 2 {
		complexityScore += 1
	}

	switch {
	case complexityScore >= 5:
		return "complex"
	case complexityScore >= 3:
		return "medium"
	default:
		return "simple"
	}
}

func (r *Router) estimateTokens(task *execution.Task) int {
	if task.Context == nil || !task.Context.HasFiles() {
		return 1000
	}

	const tokensPerFile = 2000
	const tokensPerWord = 4

	fileTokens := len(task.Context.Files) * tokensPerFile
	descriptionTokens := len(strings.Fields(task.Description)) * tokensPerWord

	return fileTokens + descriptionTokens
}

func (r *Router) estimateCost(providerDef worker.ProviderDefinition, chars taskCharacteristics, multiplier float64) float64 {
	baseCost := 0.01

	if chars.estimatedTokens > 100000 {
		baseCost = 0.05
	} else if chars.estimatedTokens > 50000 {
		baseCost = 0.03
	} else if chars.estimatedTokens > 20000 {
		baseCost = 0.02
	}

	costMultiplier := 1.0

	if providerDef.Method == config.MethodACP {
		costMultiplier = 1.5
	} else if providerDef.Method == config.MethodAPI {
		costMultiplier = 1.0
	} else if providerDef.Method == config.MethodCLI {
		costMultiplier = 0.5
	}

	if providerDef.Provider == "anthropic" {
		costMultiplier *= 2.0
	} else if providerDef.Provider == "moonshot" {
		costMultiplier *= 1.0
	} else if providerDef.Provider == "deepseek" {
		costMultiplier *= 0.5
	}

	return baseCost * multiplier * costMultiplier
}

func (r *Router) buildReason(rule RoutingRule, chars taskCharacteristics) string {
	reasons := []string{}

	if rule.Match.Complexity != "" {
		reasons = append(reasons, fmt.Sprintf("complexity=%s", rule.Match.Complexity))
	}

	if len(rule.Match.Type) > 0 {
		reasons = append(reasons, fmt.Sprintf("type=%s", strings.Join(rule.Match.Type, "|")))
	}

	if rule.Match.FileCount != nil {
		reasons = append(reasons, fmt.Sprintf("files>=%d", *rule.Match.FileCount))
	}

	if rule.Match.ContextSize != nil {
		reasons = append(reasons, fmt.Sprintf("tokens>=%d", *rule.Match.ContextSize))
	}

	if len(rule.Match.Keywords) > 0 {
		reasons = append(reasons, fmt.Sprintf("keywords=%s", strings.Join(rule.Match.Keywords, "|")))
	}

	if len(reasons) == 0 {
		return "default match"
	}

	return strings.Join(reasons, ", ")
}

func (r *Router) checkBudget(decision *RoutingDecision) error {
	if r.costBudget <= 0 {
		return nil
	}

	if decision.EstimatedCost > r.remainingBudget {
		fallback, err := r.findCheaperProvider(decision)
		if err != nil {
			return fmt.Errorf("insufficient budget ($%.4f remaining, need $%.4f for %s): %w",
				r.remainingBudget, decision.EstimatedCost, decision.Provider, err)
		}

		decision.Provider = fallback.Provider
		decision.Model = fallback.Model
		decision.Method = fallback.Method
		decision.Reason = fmt.Sprintf("budget fallback: %s (original %s too expensive at $%.4f)",
			decision.Provider, fallback.OriginalProvider, decision.EstimatedCost)
		decision.EstimatedCost = fallback.EstimatedCost

		r.logger.Warn("budget exceeded, using cheaper provider",
			"original_provider", fallback.OriginalProvider,
			"fallback_provider", decision.Provider,
			"original_cost", fmt.Sprintf("$%.4f", fallback.OriginalCost),
			"fallback_cost", fmt.Sprintf("$%.4f", decision.EstimatedCost),
			"remaining_budget", fmt.Sprintf("$%.4f", r.remainingBudget),
		)
	}

	return nil
}

type providerFallback struct {
	Provider         string
	Model            string
	Method           config.CommunicationMethod
	EstimatedCost    float64
	OriginalProvider string
	OriginalCost     float64
}

func (r *Router) findCheaperProvider(decision *RoutingDecision) (*providerFallback, error) {
	var cheapest *providerFallback
	var chars taskCharacteristics

	for name, provider := range r.providers {
		cost := r.estimateCost(provider, chars, 1.0)
		if cost < decision.EstimatedCost && cost <= r.remainingBudget {
			if cheapest == nil || cost < cheapest.EstimatedCost {
				cheapest = &providerFallback{
					Provider:         name,
					Model:            provider.Model,
					Method:           provider.Method,
					EstimatedCost:    cost,
					OriginalProvider: decision.Provider,
					OriginalCost:     decision.EstimatedCost,
				}
			}
		}
	}

	if cheapest == nil {
		return nil, fmt.Errorf("no cheaper provider available within budget")
	}

	return cheapest, nil
}

func (r *Router) RecordCost(provider string, cost float64) {
	r.costsByProvider[provider] += cost
	r.remainingBudget -= cost

	r.logger.Info("cost recorded",
		"provider", provider,
		"cost", fmt.Sprintf("$%.4f", cost),
		"remaining_budget", fmt.Sprintf("$%.4f", r.remainingBudget),
		"total_by_provider", fmt.Sprintf("$%.4f", r.costsByProvider[provider]),
	)
}

func (r *Router) GetRemainingBudget() float64 {
	return r.remainingBudget
}

func (r *Router) GetCostsByProvider() map[string]float64 {
	costs := make(map[string]float64, len(r.costsByProvider))
	for k, v := range r.costsByProvider {
		costs[k] = v
	}
	return costs
}

func (r *Router) ResetBudget() {
	r.remainingBudget = r.costBudget
	for k := range r.costsByProvider {
		delete(r.costsByProvider, k)
	}
	r.logger.Info("budget reset", "budget", fmt.Sprintf("$%.4f", r.costBudget))
}

func (r *Router) logRoutingDecision(task *execution.Task, decision *RoutingDecision) {
	r.logger.Info("routing decision",
		"task_type", task.Type,
		"method", decision.Method,
		"provider", decision.Provider,
		"model", decision.Model,
		"reason", decision.Reason,
		"estimated_cost", fmt.Sprintf("$%.4f", decision.EstimatedCost),
		"rule", decision.RuleName,
		"remaining_budget", fmt.Sprintf("$%.4f", r.remainingBudget),
	)
}

func (r *Router) SetMemoryStore(memoryStore *memory.MemoryStore) {
	r.memory = memoryStore
	r.learningEnabled = memoryStore != nil
}

func (r *Router) GetMemoryStore() *memory.MemoryStore {
	return r.memory
}

func (r *Router) EnableLearning(enabled bool) {
	r.learningEnabled = enabled
}

func (r *Router) IsLearningEnabled() bool {
	return r.learningEnabled
}
