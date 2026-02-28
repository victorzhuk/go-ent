package runtime

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/victorzhuk/go-ent/internal/domain"
)

type mockSkill struct {
	name        string
	description string
}

func (m *mockSkill) Name() string {
	return m.name
}

func (m *mockSkill) Description() string {
	return m.description
}

func (m *mockSkill) CanHandle(ctx domain.SkillContext) bool {
	return true
}

func (m *mockSkill) Execute(ctx context.Context, req domain.SkillRequest) (domain.SkillResult, error) {
	return domain.SkillResult{
		Success: true,
		Output:  "mock output",
	}, nil
}

type mockFailingSkill struct {
	name        string
	description string
}

func (m *mockFailingSkill) Name() string {
	return m.name
}

func (m *mockFailingSkill) Description() string {
	return m.description
}

func (m *mockFailingSkill) CanHandle(ctx domain.SkillContext) bool {
	return true
}

func (m *mockFailingSkill) Execute(ctx context.Context, req domain.SkillRequest) (domain.SkillResult, error) {
	return domain.SkillResult{}, assert.AnError
}

func TestNewRuntime(t *testing.T) {
	t.Parallel()

	runtime := NewRuntime()

	require.NotNil(t, runtime)
}

func TestSkillRuntime_RegisterSkill(t *testing.T) {
	t.Parallel()

	t.Run("valid skill", func(t *testing.T) {
		t.Parallel()

		runtime := NewRuntime().(*skillRuntime)
		skill := &mockSkill{name: "test-skill", description: "A test skill"}

		err := runtime.RegisterSkill(skill)

		require.NoError(t, err)
		assert.True(t, runtime.Exists("test-skill"))
	})

	t.Run("nil skill", func(t *testing.T) {
		t.Parallel()

		runtime := NewRuntime().(*skillRuntime)

		err := runtime.RegisterSkill(nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "skill cannot be nil")
	})

	t.Run("empty name", func(t *testing.T) {
		t.Parallel()

		runtime := NewRuntime().(*skillRuntime)
		skill := &mockSkill{name: "", description: "A test skill"}

		err := runtime.RegisterSkill(skill)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "skill name cannot be empty")
	})

	t.Run("duplicate skill", func(t *testing.T) {
		t.Parallel()

		runtime := NewRuntime().(*skillRuntime)
		skill1 := &mockSkill{name: "test-skill", description: "First skill"}
		skill2 := &mockSkill{name: "test-skill", description: "Second skill"}

		require.NoError(t, runtime.RegisterSkill(skill1))
		err := runtime.RegisterSkill(skill2)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrDuplicateSkill)
	})
}

func TestSkillRuntime_UnregisterSkill(t *testing.T) {
	t.Parallel()

	t.Run("unregister existing", func(t *testing.T) {
		t.Parallel()

		runtime := NewRuntime().(*skillRuntime)
		skill := &mockSkill{name: "test-skill", description: "A test skill"}
		require.NoError(t, runtime.RegisterSkill(skill))

		err := runtime.UnregisterSkill("test-skill")

		require.NoError(t, err)
		assert.False(t, runtime.Exists("test-skill"))
	})

	t.Run("unregister non-existent", func(t *testing.T) {
		t.Parallel()

		runtime := NewRuntime().(*skillRuntime)

		err := runtime.UnregisterSkill("nonexistent")

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrSkillNotFound)
	})
}

func TestSkillRuntime_GetSkill(t *testing.T) {
	t.Parallel()

	t.Run("get existing", func(t *testing.T) {
		t.Parallel()

		runtime := NewRuntime().(*skillRuntime)
		skill := &mockSkill{name: "test-skill", description: "A test skill"}
		require.NoError(t, runtime.RegisterSkill(skill))

		retrieved, err := runtime.GetSkill("test-skill")

		require.NoError(t, err)
		assert.Equal(t, skill, retrieved)
	})

	t.Run("get non-existent", func(t *testing.T) {
		t.Parallel()

		runtime := NewRuntime().(*skillRuntime)

		_, err := runtime.GetSkill("nonexistent")

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrSkillNotFound)
	})
}

func TestSkillRuntime_Execute(t *testing.T) {
	t.Parallel()

	t.Run("execute successfully", func(t *testing.T) {
		t.Parallel()

		runtime := NewRuntime().(*skillRuntime)
		skill := &mockSkill{name: "test-skill", description: "A test skill"}
		require.NoError(t, runtime.RegisterSkill(skill))

		req := domain.SkillRequest{
			Input:      "test input",
			Parameters: map[string]any{},
			Context:    domain.SkillContext{},
		}

		result, err := runtime.Execute(context.Background(), "test-skill", req)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		assert.Equal(t, "mock output", result.Output)
	})

	t.Run("execute non-existent skill", func(t *testing.T) {
		t.Parallel()

		runtime := NewRuntime().(*skillRuntime)
		req := domain.SkillRequest{}

		_, err := runtime.Execute(context.Background(), "nonexistent", req)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrSkillNotFound)
	})

	t.Run("execute failing skill", func(t *testing.T) {
		t.Parallel()

		runtime := NewRuntime().(*skillRuntime)
		skill := &mockFailingSkill{name: "failing-skill", description: "A failing skill"}
		require.NoError(t, runtime.RegisterSkill(skill))

		req := domain.SkillRequest{}

		_, err := runtime.Execute(context.Background(), "failing-skill", req)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "execute skill")
	})
}

func TestSkillRuntime_ListSkills(t *testing.T) {
	t.Parallel()

	t.Run("empty runtime", func(t *testing.T) {
		t.Parallel()

		runtime := NewRuntime()
		skills := runtime.ListSkills()

		assert.NotNil(t, skills)
		assert.Empty(t, skills)
	})

	t.Run("multiple skills", func(t *testing.T) {
		t.Parallel()

		runtime := NewRuntime().(*skillRuntime)

		skill1 := &mockSkill{name: "skill-1", description: "First skill"}
		skill2 := &mockSkill{name: "skill-2", description: "Second skill"}
		skill3 := &mockSkill{name: "skill-3", description: "Third skill"}

		require.NoError(t, runtime.RegisterSkill(skill1))
		require.NoError(t, runtime.RegisterSkill(skill2))
		require.NoError(t, runtime.RegisterSkill(skill3))

		skills := runtime.ListSkills()

		assert.Len(t, skills, 3)
	})
}

func TestSkillRuntime_Exists(t *testing.T) {
	t.Parallel()

	t.Run("exists", func(t *testing.T) {
		t.Parallel()

		runtime := NewRuntime().(*skillRuntime)
		skill := &mockSkill{name: "test-skill", description: "A test skill"}
		require.NoError(t, runtime.RegisterSkill(skill))

		assert.True(t, runtime.Exists("test-skill"))
		assert.False(t, runtime.Exists("nonexistent"))
	})

	t.Run("empty name", func(t *testing.T) {
		t.Parallel()

		runtime := NewRuntime()
		assert.False(t, runtime.Exists(""))
	})
}

func TestSkillRuntime_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping concurrent test in short mode")
	}

	runtime := NewRuntime().(*skillRuntime)
	done := make(chan bool, 10)

	for i := 0; i < 5; i++ {
		go func(idx int) {
			skill := &mockSkill{
				name:        "skill-" + string(rune('a'+idx)),
				description: "A test skill",
			}
			_ = runtime.RegisterSkill(skill)
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		go func() {
			_ = runtime.ListSkills()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
