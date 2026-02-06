package domain

import "fmt"

// Skill represents a skill's metadata and capabilities.
// This is the core domain entity for skill registry operations.
type Skill struct {
	ID               string
	Name             string
	Description      string
	Version          string
	Author           string
	Category         string
	Tags             []string
	AllowedTools     []string
	StructureVersion string
	DependsOn        []string
	FilePath         string
	Triggers         []string
	ExplicitTriggers []Trigger
	DelegatesTo      map[string]string
	Role             string
	Instructions     string
	Examples         string
	References       []string
}

// Trigger represents an explicit trigger for skill activation.
type Trigger struct {
	Patterns     []string
	Keywords     []string
	FilePatterns []string
	Weight       float64
}

// NewSkill creates a new Skill instance.
func NewSkill(name, description, version string) (*Skill, error) {
	if name == "" {
		return nil, ErrInvalidSkillName
	}
	if description == "" {
		return nil, ErrInvalidSkillDescription
	}
	if version == "" {
		version = "1.0.0"
	}

	return &Skill{
		ID:               fmt.Sprintf("%s@%s", name, version),
		Name:             name,
		Description:      description,
		Version:          version,
		Triggers:         make([]string, 0),
		ExplicitTriggers: make([]Trigger, 0),
		Tags:             make([]string, 0),
		AllowedTools:     make([]string, 0),
		DependsOn:        make([]string, 0),
		References:       make([]string, 0),
		DelegatesTo:      make(map[string]string),
	}, nil
}

// AddTrigger adds a trigger string to the skill.
func (s *Skill) AddTrigger(trigger string) {
	s.Triggers = append(s.Triggers, trigger)
}

// AddExplicitTrigger adds an explicit trigger to the skill.
func (s *Skill) AddExplicitTrigger(trigger Trigger) {
	s.ExplicitTriggers = append(s.ExplicitTriggers, trigger)
}

// AddTag adds a tag to the skill.
func (s *Skill) AddTag(tag string) {
	s.Tags = append(s.Tags, tag)
}

// AddAllowedTool adds an allowed tool to the skill.
func (s *Skill) AddAllowedTool(tool string) {
	s.AllowedTools = append(s.AllowedTools, tool)
}

// AddDependency adds a skill dependency.
func (s *Skill) AddDependency(skillName string) {
	s.DependsOn = append(s.DependsOn, skillName)
}

// AddReference adds a reference file path.
func (s *Skill) AddReference(ref string) {
	s.References = append(s.References, ref)
}

// AddDelegation adds a delegation mapping to another skill.
func (s *Skill) AddDelegation(toSkill, reason string) {
	if s.DelegatesTo == nil {
		s.DelegatesTo = make(map[string]string)
	}
	s.DelegatesTo[toSkill] = reason
}
