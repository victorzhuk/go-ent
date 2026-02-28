package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/victorzhuk/go-ent/internal/domain"

	skilldomain "github.com/victorzhuk/go-ent/internal/skill/domain"
	skillmatcher "github.com/victorzhuk/go-ent/internal/skill/matcher"
	skillrepository "github.com/victorzhuk/go-ent/internal/skill/repository"
	skillruntime "github.com/victorzhuk/go-ent/internal/skill/runtime"
)

// Repository defines the interface for skill persistence.
// Defined at the consumer side following Clean Architecture.
type Repository interface {
	Save(skill *skilldomain.Skill) error
	FindByID(id string) (*skilldomain.Skill, error)
	FindByName(name string) (*skilldomain.Skill, error)
	ListAll() ([]*skilldomain.Skill, error)
	Delete(id string) error
	Update(skill *skilldomain.Skill) error
	Exists(name string) bool
}

// Registry manages skill metadata and matching.
// It composes repository, matcher, and runtime components.
type Registry struct {
	repo      Repository
	matcher   skillmatcher.Matcher
	runtime   skillruntime.Runtime
	parser    *Parser
	validator *Validator

	// Internal fields for backward compatibility with tests
	skills        []SkillMeta
	runtimeSkills map[string]domain.Skill
}

// NewRegistry creates a new skill registry.
func NewRegistry() *Registry {
	repo := skillrepository.NewInMemoryRepository()
	matcher := skillmatcher.NewMatcher(repo)
	runtime := skillruntime.NewRuntime()

	return &Registry{
		repo:          repo,
		matcher:       matcher,
		runtime:       runtime,
		parser:        NewParser(),
		validator:     NewValidator(),
		skills:        make([]SkillMeta, 0),
		runtimeSkills: make(map[string]domain.Skill),
	}
}

// skillMetaToDomain converts SkillMeta to skilldomain.Skill.
func (r *Registry) skillMetaToDomain(meta *SkillMeta) (*skilldomain.Skill, error) {
	skill, err := skilldomain.NewSkill(meta.Name, meta.Description, meta.Version)
	if err != nil {
		return nil, err
	}

	skill.Author = meta.Author
	skill.Category = meta.Category
	skill.FilePath = meta.FilePath
	skill.StructureVersion = meta.StructureVersion
	skill.Role = meta.Role
	skill.Instructions = meta.Instructions
	skill.Examples = meta.Examples
	skill.References = meta.References
	skill.Tags = meta.Tags
	skill.AllowedTools = meta.AllowedTools
	skill.DependsOn = meta.DependsOn
	skill.DelegatesTo = meta.DelegatesTo

	for _, trigger := range meta.Triggers {
		skill.AddTrigger(trigger)
	}

	for _, explicitTrigger := range meta.ExplicitTriggers {
		skill.AddExplicitTrigger(skilldomain.Trigger{
			Patterns:     explicitTrigger.Patterns,
			Keywords:     explicitTrigger.Keywords,
			FilePatterns: explicitTrigger.FilePatterns,
			Weight:       explicitTrigger.Weight,
		})
	}

	return skill, nil
}

// domainToSkillMeta converts skilldomain.Skill to *SkillMeta for backward compatibility.
func (r *Registry) domainToSkillMeta(skill *skilldomain.Skill) *SkillMeta {
	return &SkillMeta{
		Name:             skill.Name,
		Description:      skill.Description,
		Version:          skill.Version,
		Author:           skill.Author,
		Category:         skill.Category,
		Tags:             skill.Tags,
		AllowedTools:     skill.AllowedTools,
		StructureVersion: skill.StructureVersion,
		DependsOn:        skill.DependsOn,
		FilePath:         skill.FilePath,
		Triggers:         skill.Triggers,
		DelegatesTo:      skill.DelegatesTo,
		Role:             skill.Role,
		Instructions:     skill.Instructions,
		Examples:         skill.Examples,
		References:       skill.References,
		ExplicitTriggers: r.convertExplicitTriggers(skill.ExplicitTriggers),
	}
}

// convertExplicitTriggers converts skilldomain.Trigger to Trigger.
func (r *Registry) convertExplicitTriggers(triggers []skilldomain.Trigger) []Trigger {
	result := make([]Trigger, len(triggers))
	for i, t := range triggers {
		result[i] = Trigger{
			Patterns:     t.Patterns,
			Keywords:     t.Keywords,
			FilePatterns: t.FilePatterns,
			Weight:       t.Weight,
		}
	}
	return result
}

// convertMatchReasons converts skillmatcher.MatchReason to []MatchReason.
func (r *Registry) convertMatchReasons(reasons []skillmatcher.MatchReason) []MatchReason {
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

// convertDelegations converts skillmatcher.DelegationHint to []DelegationHint.
func (r *Registry) convertDelegations(hints []skillmatcher.DelegationHint) []DelegationHint {
	result := make([]DelegationHint, len(hints))
	for i, h := range hints {
		result[i] = DelegationHint{
			ToSkill: h.ToSkill,
			Reason:  h.Reason,
		}
	}
	return result
}

// DelegationHint provides a hint for delegating work to another skill.
type DelegationHint struct {
	ToSkill string // Name of skill to delegate to
	Reason  string // Why delegation should happen
}

// MatchResult represents a skill match with its confidence score and reasons.
type MatchResult struct {
	Skill       *SkillMeta       // The matched skill
	Score       float64          // 0.0-1.0 confidence score
	MatchedBy   []MatchReason    // List of what triggered the match
	Delegations []DelegationHint // Hints for skill delegation
}

// MatchReason explains why a skill was matched.
type MatchReason struct {
	Type   string  // "keyword", "pattern", "file_type"
	Value  string  // The specific value that matched
	Weight float64 // The trigger weight
}

// Register adds a runtime skill to the registry.
func (r *Registry) Register(skill domain.Skill) error {
	return r.runtime.RegisterSkill(skill)
}

// Unregister removes a runtime skill from the registry.
func (r *Registry) Unregister(name string) error {
	return r.runtime.UnregisterSkill(name)
}

// GetSkill retrieves a runtime skill by name.
func (r *Registry) GetSkill(name string) (domain.Skill, error) {
	return r.runtime.GetSkill(name)
}

// Load scans a directory for SKILL.md files and loads their metadata.
func (r *Registry) Load(skillsPath string) error {
	err := filepath.Walk(skillsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || info.Name() != "SKILL.md" {
			return nil
		}

		meta, err := r.parser.ParseSkillFile(path)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		// Convert SkillMeta to domain.Skill
		skill, err := r.skillMetaToDomain(meta)
		if err != nil {
			return fmt.Errorf("convert skill metadata: %w", err)
		}

		// Try to update if exists, otherwise save
		if r.repo.Exists(skill.Name) {
			if err := r.repo.Update(skill); err != nil {
				return fmt.Errorf("update skill %s: %w", skill.Name, err)
			}
		} else {
			if err := r.repo.Save(skill); err != nil {
				return fmt.Errorf("save skill %s: %w", skill.Name, err)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// RegisterSkill loads a skill from a given path and registers it.
func (r *Registry) RegisterSkill(name, path string) error {
	meta, err := r.parser.ParseSkillFile(path)
	if err != nil {
		return fmt.Errorf("parse skill file: %w", err)
	}

	if meta.Name != name {
		return fmt.Errorf("skill name mismatch: expected %s, got %s", name, meta.Name)
	}

	if r.repo.Exists(name) {
		return fmt.Errorf("skill %s already registered", name)
	}

	// Convert SkillMeta to domain.Skill
	skill, err := r.skillMetaToDomain(meta)
	if err != nil {
		return fmt.Errorf("convert skill metadata: %w", err)
	}

	return r.repo.Save(skill)
}

// UnregisterSkill removes a skill from the metadata list.
func (r *Registry) UnregisterSkill(name string) error {
	skill, err := r.repo.FindByName(name)
	if err != nil {
		return err
	}
	return r.repo.Delete(skill.ID)
}

// MatchForContext returns skill names that match the given context.
func (r *Registry) MatchForContext(ctx domain.SkillContext) []string {
	var matched []string

	// Check runtime skills first
	for _, skill := range r.runtime.ListSkills() {
		if skill.CanHandle(ctx) {
			matched = append(matched, skill.Name())
		}
	}

	// Then check metadata skills
	skills, _ := r.repo.ListAll()
	terms := r.buildSearchTerms(ctx)
	for _, skill := range skills {
		skillMeta := r.domainToSkillMeta(skill)
		if r.matchesContext(*skillMeta, terms) {
			matched = append(matched, skill.Name)
		}
	}

	return matched
}

// FindMatchingSkills returns skills with match scores and reasons based on the given query and context.
func (r *Registry) FindMatchingSkills(query string, context ...*skillmatcher.MatchContext) []MatchResult {
	if len(context) == 0 || context[0] == nil {
		return r.matchByQuery(query)
	}

	ctx := context[0]

	// Use matcher for metadata skills
	matcherResults, _ := r.matcher.MatchWithScoring(query, ctx)

	var results []MatchResult
	for _, mr := range matcherResults {
		results = append(results, MatchResult{
			Skill:       r.domainToSkillMeta(mr.Skill),
			Score:       mr.Score,
			MatchedBy:   r.convertMatchReasons(mr.MatchedBy),
			Delegations: r.convertDelegations(mr.Delegations),
		})
	}

	// Add runtime skills
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

// matchByQuery performs query-based skill matching (backward compatible).
func (r *Registry) matchByQuery(query string) []MatchResult {
	var results []MatchResult
	queryLower := strings.ToLower(query)

	// Use matcher for metadata skills
	matcherResults, _ := r.matcher.MatchByQuery(query)
	for _, mr := range matcherResults {
		results = append(results, MatchResult{
			Skill:       r.domainToSkillMeta(mr.Skill),
			Score:       mr.Score,
			MatchedBy:   r.convertMatchReasons(mr.MatchedBy),
			Delegations: r.convertDelegations(mr.Delegations),
		})
	}

	// Add runtime skills
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

// Get retrieves a skill by name.
func (r *Registry) Get(name string) (*SkillMeta, error) {
	skill, err := r.repo.FindByName(name)
	if err != nil {
		return nil, fmt.Errorf("skill not found: %s", name)
	}
	return r.domainToSkillMeta(skill), nil
}

// All returns all loaded skills.
func (r *Registry) All() []SkillMeta {
	skills, _ := r.repo.ListAll()
	result := make([]SkillMeta, len(skills))
	for i, skill := range skills {
		result[i] = *r.domainToSkillMeta(skill)
	}
	return result
}

// buildSearchTerms extracts searchable terms from SkillContext.
func (r *Registry) buildSearchTerms(ctx domain.SkillContext) []string {
	var terms []string

	// Add action as term
	if ctx.Action != "" {
		terms = append(terms, strings.ToLower(string(ctx.Action)))
	}

	// Add phase as term
	if ctx.Phase != "" {
		terms = append(terms, strings.ToLower(string(ctx.Phase)))
	}

	// Add agent role as term
	if ctx.Agent != "" {
		terms = append(terms, strings.ToLower(string(ctx.Agent)))
	}

	// Extract terms from metadata
	if ctx.Metadata != nil {
		for key, val := range ctx.Metadata {
			keyLower := strings.ToLower(key)
			terms = append(terms, keyLower)

			// Handle string values
			if strVal, ok := val.(string); ok {
				terms = append(terms, strings.ToLower(strVal))
			}
		}
	}

	return terms
}

// matchesContext checks if a skill's triggers match any context terms.
func (r *Registry) matchesContext(skill SkillMeta, terms []string) bool {
	if len(skill.Triggers) == 0 {
		return false
	}

	for _, trigger := range skill.Triggers {
		for _, term := range terms {
			// Exact match
			if trigger == term {
				return true
			}

			// Partial match (term contains trigger)
			if strings.Contains(term, trigger) {
				return true
			}

			// Partial match (trigger contains term)
			if strings.Contains(trigger, term) {
				return true
			}
		}
	}

	return false
}

// ValidateSkill validates a single skill by name.
func (r *Registry) ValidateSkill(name string) (*ValidationResult, error) {
	meta, err := r.Get(name)
	if err != nil {
		return nil, fmt.Errorf("get skill metadata: %w", err)
	}

	content, err := os.ReadFile(meta.FilePath) // #nosec G304 -- controlled skill file path
	if err != nil {
		return nil, fmt.Errorf("read skill file: %w", err)
	}

	result := r.validator.Validate(meta, string(content))
	return result, nil
}
