package domain

//nolint:gosec // test file with necessary file operations

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockSkill struct {
	name        string
	description string
	canHandle   bool
	result      SkillResult
	err         error
}

func (m *mockSkill) Name() string {
	return m.name
}

func (m *mockSkill) Description() string {
	return m.description
}

func (m *mockSkill) CanHandle(ctx SkillContext) bool {
	return m.canHandle
}

func (m *mockSkill) Execute(ctx context.Context, req SkillRequest) (SkillResult, error) {
	if m.err != nil {
		return SkillResult{}, m.err
	}
	return m.result, nil
}

func TestSkill_Interface(t *testing.T) {
	skill := &mockSkill{
		name:        "test-skill",
		description: "A test skill",
		canHandle:   true,
	}

	assert.Equal(t, "test-skill", skill.Name())
	assert.Equal(t, "A test skill", skill.Description())
}

func TestSkill_CanHandle(t *testing.T) {
	tests := []struct {
		name  string
		skill *mockSkill
		ctx   SkillContext
		want  bool
	}{
		{
			name: "can handle",
			skill: &mockSkill{
				canHandle: true,
			},
			ctx: SkillContext{
				Action: SpecActionImplement,
			},
			want: true,
		},
		{
			name: "cannot handle",
			skill: &mockSkill{
				canHandle: false,
			},
			ctx: SkillContext{
				Action: SpecActionReview,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.skill.CanHandle(tt.ctx)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestSkill_Execute_Success(t *testing.T) {
	skill := &mockSkill{
		result: SkillResult{
			Success: true,
			Output:  "execution successful",
		},
	}

	req := SkillRequest{
		Input: "test input",
		Parameters: map[string]interface{}{
			"param1": "value1",
		},
	}

	result, err := skill.Execute(context.Background(), req)
	assert.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "execution successful", result.Output)
}

func TestSkill_Execute_Failure(t *testing.T) {
	skill := &mockSkill{
		err: errors.New("execution failed"),
	}

	req := SkillRequest{
		Input: "test input",
	}

	_, err := skill.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, "execution failed", err.Error())
}

func TestSkillMetadata(t *testing.T) {
	metadata := SkillMetadata{
		Name:        "go-code",
		Description: "Go code generation and manipulation",
		Version:     "1.0.0",
		Author:      "go-ent",
		Tags:        []string{"code", "go", "generation"},
	}

	assert.Equal(t, "go-code", metadata.Name)
	assert.Equal(t, "Go code generation and manipulation", metadata.Description)
	assert.Equal(t, "1.0.0", metadata.Version)
	assert.Equal(t, "go-ent", metadata.Author)
	assert.Len(t, metadata.Tags, 3)
	assert.Contains(t, metadata.Tags, "code")
	assert.Contains(t, metadata.Tags, "go")
	assert.Contains(t, metadata.Tags, "generation")
}

func TestSkillContext(t *testing.T) {
	ctx := SkillContext{
		Action:  SpecActionImplement,
		Phase:   ActionPhaseExecution,
		Runtime: RuntimeClaudeCode,
		Agent:   AgentRoleDeveloper,
		Metadata: map[string]interface{}{
			"change_id": "add-feature",
			"priority":  "high",
		},
	}

	assert.Equal(t, SpecActionImplement, ctx.Action)
	assert.Equal(t, ActionPhaseExecution, ctx.Phase)
	assert.Equal(t, RuntimeClaudeCode, ctx.Runtime)
	assert.Equal(t, AgentRoleDeveloper, ctx.Agent)
	assert.Equal(t, "add-feature", ctx.Metadata["change_id"])
	assert.Equal(t, "high", ctx.Metadata["priority"])
}

func TestSkillRequest(t *testing.T) {
	req := SkillRequest{
		Input: "implement user authentication",
		Parameters: map[string]interface{}{
			"method": "jwt",
			"expiry": 3600,
		},
		Context: SkillContext{
			Action: SpecActionImplement,
			Agent:  AgentRoleDeveloper,
		},
	}

	assert.Equal(t, "implement user authentication", req.Input)
	assert.Equal(t, "jwt", req.Parameters["method"])
	assert.Equal(t, 3600, req.Parameters["expiry"])
	assert.Equal(t, SpecActionImplement, req.Context.Action)
	assert.Equal(t, AgentRoleDeveloper, req.Context.Agent)
}

func TestSkillResult_Success(t *testing.T) {
	result := SkillResult{
		Success: true,
		Output:  "user authentication implemented",
		Metadata: map[string]interface{}{
			"files_created": []string{"auth.go", "auth_test.go"},
			"lines_of_code": 250,
		},
	}

	assert.True(t, result.Success)
	assert.Equal(t, "user authentication implemented", result.Output)
	assert.Empty(t, result.Error)

	files := result.Metadata["files_created"].([]string)
	assert.Len(t, files, 2)
	assert.Contains(t, files, "auth.go")

	lines := result.Metadata["lines_of_code"].(int)
	assert.Equal(t, 250, lines)
}

func TestSkillResult_Failure(t *testing.T) {
	result := SkillResult{
		Success: false,
		Error:   "compilation failed",
		Metadata: map[string]interface{}{
			"error_type": "syntax_error",
			"line":       42,
		},
	}

	assert.False(t, result.Success)
	assert.Empty(t, result.Output)
	assert.Equal(t, "compilation failed", result.Error)
	assert.Equal(t, "syntax_error", result.Metadata["error_type"])
	assert.Equal(t, 42, result.Metadata["line"])
}

func TestSkillContext_Empty(t *testing.T) {
	ctx := SkillContext{}

	assert.Equal(t, SpecAction(""), ctx.Action)
	assert.Equal(t, ActionPhase(""), ctx.Phase)
	assert.Equal(t, Runtime(""), ctx.Runtime)
	assert.Equal(t, AgentRole(""), ctx.Agent)
	assert.Nil(t, ctx.Metadata)
}

func TestSkillRequest_Empty(t *testing.T) {
	req := SkillRequest{}

	assert.Empty(t, req.Input)
	assert.Nil(t, req.Parameters)
	assert.Equal(t, SkillContext{}, req.Context)
}

func TestSkillResult_Empty(t *testing.T) {
	result := SkillResult{}

	assert.False(t, result.Success)
	assert.Empty(t, result.Output)
	assert.Empty(t, result.Error)
	assert.Nil(t, result.Metadata)
}
