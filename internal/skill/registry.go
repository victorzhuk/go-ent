package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/victorzhuk/go-ent/internal/domain"
	skdomain "github.com/victorzhuk/go-ent/internal/skill/domain"
	skillmatcher "github.com/victorzhuk/go-ent/internal/skill/matcher"
	skillruntime "github.com/victorzhuk/go-ent/internal/skill/runtime"
)

type DelegationHint struct {
	ToSkill string
	Reason  string
}

type MatchResult struct {
	Skill       *skdomain.Info
	Score       float64
	MatchedBy   []MatchReason
	Delegations []DelegationHint
}

type MatchReason struct {
	Type   string
	Value  string
	Weight float64
}

type Registry struct {
	mu        sync.RWMutex
	skills    map[string]*skdomain.Info
	names     map[string]string
	matcher   skillmatcher.Matcher
	runtime   skillruntime.Runtime
	parser    *Parser
	validator *Validator
}

func NewRegistry() *Registry {
	r := &Registry{
		skills:    make(map[string]*skdomain.Info),
		names:     make(map[string]string),
		runtime:   skillruntime.NewRuntime(),
		parser:    NewParser(),
		validator: NewValidator(),
	}
	r.matcher = skillmatcher.NewMatcher(r)
	return r
}

func (r *Registry) Register(skill domain.Skill) error {
	return r.runtime.RegisterSkill(skill)
}

func (r *Registry) Unregister(name string) error {
	return r.runtime.UnregisterSkill(name)
}

func (r *Registry) GetSkill(name string) (domain.Skill, error) {
	return r.runtime.GetSkill(name)
}

func (r *Registry) Load(skillsPath string) error {
	err := filepath.Walk(skillsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || info.Name() != "SKILL.md" {
			return nil
		}

		skill, err := r.parser.ParseSkillFile(path)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		if r.exists(skill.Name) {
			if err := r.update(skill); err != nil {
				return fmt.Errorf("update skill %s: %w", skill.Name, err)
			}
		} else {
			if err := r.save(skill); err != nil {
				return fmt.Errorf("save skill %s: %w", skill.Name, err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("load skills from %s: %w", skillsPath, err)
	}
	return nil
}

func (r *Registry) RegisterSkill(name, path string) error {
	skill, err := r.parser.ParseSkillFile(path)
	if err != nil {
		return fmt.Errorf("parse skill file: %w", err)
	}

	if skill.Name != name {
		return fmt.Errorf("skill name mismatch: expected %s, got %s", name, skill.Name)
	}

	if r.exists(name) {
		return fmt.Errorf("skill %s already registered", name)
	}

	return r.save(skill)
}

func (r *Registry) UnregisterSkill(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id, exists := r.names[name]
	if !exists {
		return fmt.Errorf("skill not found: %s", name)
	}

	delete(r.skills, id)
	delete(r.names, name)
	return nil
}

func (r *Registry) MatchForContext(ctx domain.SkillContext) []string {
	var matched []string

	for _, skill := range r.runtime.ListSkills() {
		if skill.CanHandle(ctx) {
			matched = append(matched, skill.Name())
		}
	}

	r.mu.RLock()
	skills := r.skills
	r.mu.RUnlock()

	terms := r.buildSearchTerms(ctx)
	for _, skill := range skills {
		if r.matchesContext(skill, terms) {
			matched = append(matched, skill.Name)
		}
	}

	return matched
}

func (r *Registry) FindMatchingSkills(query string, context ...*skillmatcher.MatchContext) []MatchResult {
	if len(context) == 0 || context[0] == nil {
		return r.matchByQuery(query)
	}

	ctx := context[0]
	matcherResults, _ := r.matcher.MatchWithScoring(query, ctx)

	var results []MatchResult
	for _, mr := range matcherResults {
		results = append(results, MatchResult{
			Skill:       mr.Skill,
			Score:       mr.Score,
			MatchedBy:   convertMatchReasons(mr.MatchedBy),
			Delegations: convertDelegations(mr.Delegations),
		})
	}

	for _, skill := range r.runtime.ListSkills() {
		if strings.Contains(strings.ToLower(skill.Name()), strings.ToLower(query)) {
			results = append(results, MatchResult{
				Skill:       nil,
				Score:       1.0,
				MatchedBy:   []MatchReason{{Type: "runtime", Value: skill.Name(), Weight: 1.0}},
				Delegations: []DelegationHint{},
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

func (r *Registry) matchByQuery(query string) []MatchResult {
	var results []MatchResult
	queryLower := strings.ToLower(query)

	matcherResults, _ := r.matcher.MatchByQuery(query)
	for _, mr := range matcherResults {
		results = append(results, MatchResult{
			Skill:       mr.Skill,
			Score:       mr.Score,
			MatchedBy:   convertMatchReasons(mr.MatchedBy),
			Delegations: convertDelegations(mr.Delegations),
		})
	}

	for _, skill := range r.runtime.ListSkills() {
		if strings.Contains(strings.ToLower(skill.Name()), queryLower) {
			results = append(results, MatchResult{
				Skill:       nil,
				Score:       0.5,
				MatchedBy:   []MatchReason{{Type: "runtime", Value: skill.Name(), Weight: 0.5}},
				Delegations: []DelegationHint{},
			})
		}
	}

	return results
}

func (r *Registry) Get(name string) (*skdomain.Info, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.names[name]
	if !exists {
		return nil, fmt.Errorf("skill not found: %s", name)
	}
	return r.skills[id], nil
}

func (r *Registry) All() []*skdomain.Info {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*skdomain.Info, 0, len(r.skills))
	for _, skill := range r.skills {
		result = append(result, skill)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func (r *Registry) ValidateSkill(name string) (*ValidationResult, error) {
	info, err := r.Get(name)
	if err != nil {
		return nil, fmt.Errorf("get skill metadata: %w", err)
	}

	content, err := os.ReadFile(info.FilePath)
	if err != nil {
		return nil, fmt.Errorf("read skill file: %w", err)
	}

	result := r.validator.Validate(info, string(content))
	return result, nil
}

func (r *Registry) ListAll() ([]*skdomain.Info, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills := make([]*skdomain.Info, 0, len(r.skills))
	for _, skill := range r.skills {
		skills = append(skills, skill)
	}
	return skills, nil
}

func (r *Registry) save(skill *skdomain.Info) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if skill.ID == "" {
		return fmt.Errorf("skill ID cannot be empty")
	}
	if skill.Name == "" {
		return fmt.Errorf("skill name cannot be empty")
	}

	if _, exists := r.skills[skill.ID]; exists {
		return fmt.Errorf("%w: %s", skdomain.ErrDuplicate, skill.Name)
	}

	if _, exists := r.names[skill.Name]; exists {
		return fmt.Errorf("%w: %s", skdomain.ErrDuplicate, skill.Name)
	}

	r.skills[skill.ID] = skill
	r.names[skill.Name] = skill.ID
	return nil
}

func (r *Registry) update(skill *skdomain.Info) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if skill.ID == "" {
		return fmt.Errorf("skill ID cannot be empty")
	}
	if skill.Name == "" {
		return fmt.Errorf("skill name cannot be empty")
	}

	if _, exists := r.skills[skill.ID]; !exists {
		return fmt.Errorf("%w: %s", skdomain.ErrNotFound, skill.ID)
	}

	oldName := ""
	for name, id := range r.names {
		if id == skill.ID {
			oldName = name
			break
		}
	}

	if oldName != "" && oldName != skill.Name {
		delete(r.names, oldName)
	}

	r.skills[skill.ID] = skill
	r.names[skill.Name] = skill.ID
	return nil
}

func (r *Registry) exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.names[name]
	return exists
}

func (r *Registry) buildSearchTerms(ctx domain.SkillContext) []string {
	var terms []string

	if ctx.Action != "" {
		terms = append(terms, strings.ToLower(string(ctx.Action)))
	}

	if ctx.Phase != "" {
		terms = append(terms, strings.ToLower(string(ctx.Phase)))
	}

	if ctx.Agent != "" {
		terms = append(terms, strings.ToLower(string(ctx.Agent)))
	}

	if ctx.Metadata != nil {
		for key, val := range ctx.Metadata {
			keyLower := strings.ToLower(key)
			terms = append(terms, keyLower)

			if strVal, ok := val.(string); ok {
				terms = append(terms, strings.ToLower(strVal))
			}
		}
	}

	return terms
}

func (r *Registry) matchesContext(skill *skdomain.Info, terms []string) bool {
	if len(skill.Triggers) == 0 {
		return false
	}

	for _, trigger := range skill.Triggers {
		for _, term := range terms {
			if trigger == term {
				return true
			}

			if strings.Contains(term, trigger) {
				return true
			}

			if strings.Contains(trigger, term) {
				return true
			}
		}
	}

	return false
}

func convertMatchReasons(reasons []skillmatcher.MatchReason) []MatchReason {
	result := make([]MatchReason, len(reasons))
	for i, rsn := range reasons {
		result[i] = MatchReason{
			Type:   rsn.Type,
			Value:  rsn.Value,
			Weight: rsn.Weight,
		}
	}
	return result
}

func convertDelegations(hints []skillmatcher.DelegationHint) []DelegationHint {
	result := make([]DelegationHint, len(hints))
	for i, h := range hints {
		result[i] = DelegationHint{
			ToSkill: h.ToSkill,
			Reason:  h.Reason,
		}
	}
	return result
}
