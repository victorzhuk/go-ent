package domain

import "fmt"

type Info struct {
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

type Trigger struct {
	Patterns     []string
	Keywords     []string
	FilePatterns []string
	Weight       float64
}

func NewInfo(name, description, version string) (*Info, error) {
	if name == "" {
		return nil, ErrInvalidName
	}
	if description == "" {
		return nil, ErrInvalidDescription
	}
	if version == "" {
		version = "1.0.0"
	}

	return &Info{
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
