package genspec

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatorBasic(t *testing.T) {
	// Create validator
	v, err := NewValidator("claude")
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}

	// Create temp dir
	tmpDir := t.TempDir()

	// Write valid agent file
	validAgent := `---
name: test
description: Test agent
model: sonnet
skills:
  - go-code
---

Test content`

	validPath := filepath.Join(tmpDir, "valid.md")
	if err := os.WriteFile(validPath, []byte(validAgent), 0o644); err != nil {
		t.Fatalf("write valid file: %v", err)
	}

	// Validate
	result := v.ValidateAgent(validPath)
	if !result.IsValid() {
		t.Errorf("valid agent should pass validation, got errors: %v", result.Errors)
	}
}

func TestValidatorInvalidModel(t *testing.T) {
	v, err := NewValidator("claude")
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}

	tmpDir := t.TempDir()

	// Write agent with invalid model
	invalidAgent := `---
name: test
description: Test agent
model: invalid-model
---

Test content`

	invalidPath := filepath.Join(tmpDir, "invalid.md")
	if err := os.WriteFile(invalidPath, []byte(invalidAgent), 0o644); err != nil {
		t.Fatalf("write invalid file: %v", err)
	}

	result := v.ValidateAgent(invalidPath)
	if result.IsValid() {
		t.Error("invalid model should fail validation")
	}

	// Check error message
	found := false
	for _, e := range result.Errors {
		if e.Field == "model" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error on 'model' field")
	}
}

func TestValidatorMissingRequired(t *testing.T) {
	v, err := NewValidator("claude")
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}

	tmpDir := t.TempDir()

	// Write agent missing required 'description'
	missingAgent := `---
name: test
model: sonnet
---

Test content`

	missingPath := filepath.Join(tmpDir, "missing.md")
	if err := os.WriteFile(missingPath, []byte(missingAgent), 0o644); err != nil {
		t.Fatalf("write missing file: %v", err)
	}

	result := v.ValidateAgent(missingPath)
	if result.IsValid() {
		t.Error("missing required field should fail validation")
	}

	// Check for description error
	found := false
	for _, e := range result.Errors {
		if e.Field == "description" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error on 'description' field")
	}
}

func TestValidatorUnknownField(t *testing.T) {
	v, err := NewValidator("claude")
	if err != nil {
		t.Fatalf("NewValidator: %v", err)
	}

	tmpDir := t.TempDir()

	// Write agent with unknown field
	unknownAgent := `---
name: test
description: Test agent
unknownField: value
---

Test content`

	unknownPath := filepath.Join(tmpDir, "unknown.md")
	if err := os.WriteFile(unknownPath, []byte(unknownAgent), 0o644); err != nil {
		t.Fatalf("write unknown file: %v", err)
	}

	result := v.ValidateAgent(unknownPath)
	if result.IsValid() {
		t.Error("unknown field should fail validation in strict mode")
	}

	// Check for unknownField error
	found := false
	for _, e := range result.Errors {
		if e.Field == "unknownField" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error on 'unknownField'")
	}
}
