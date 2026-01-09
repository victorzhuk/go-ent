package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/victorzhuk/go-ent/internal/domain"
)

// Registry manages skill metadata and matching.
type Registry struct {
	skills []SkillMeta
	parser *Parser
}

// NewRegistry creates a new skill registry.
func NewRegistry() *Registry {
	return &Registry{
		skills: make([]SkillMeta, 0),
		parser: NewParser(),
	}
}

// Load scans a directory for SKILL.md files and loads their metadata.
func (r *Registry) Load(skillsPath string) error {
	return filepath.Walk(skillsPath, func(path string, info os.FileInfo, err error) error {
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

		r.skills = append(r.skills, *meta)
		return nil
	})
}

// MatchForContext returns skill names that match the given context.
func (r *Registry) MatchForContext(ctx domain.SkillContext) []string {
	var matched []string

	// Build search terms from context
	terms := r.buildSearchTerms(ctx)

	for _, skill := range r.skills {
		if r.matchesContext(skill, terms) {
			matched = append(matched, skill.Name)
		}
	}

	return matched
}

// Get retrieves a skill by name.
func (r *Registry) Get(name string) (*SkillMeta, error) {
	for _, skill := range r.skills {
		if skill.Name == name {
			return &skill, nil
		}
	}
	return nil, fmt.Errorf("skill not found: %s", name)
}

// All returns all loaded skills.
func (r *Registry) All() []SkillMeta {
	return r.skills
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
