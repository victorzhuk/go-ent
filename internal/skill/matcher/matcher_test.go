package matcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	skilldomain "github.com/victorzhuk/go-ent/internal/skill/domain"
)

type mockMatcherRepository struct {
	skills []*skilldomain.Skill
}

func (m *mockMatcherRepository) ListAll() ([]*skilldomain.Skill, error) {
	return m.skills, nil
}

func TestNewMatcher(t *testing.T) {
	t.Parallel()

	repo := &mockMatcherRepository{}
	matcher := NewMatcher(repo)

	require.NotNil(t, matcher)
}

func TestMatcher_MatchByQuery(t *testing.T) {
	t.Parallel()

	t.Run("empty repository", func(t *testing.T) {
		t.Parallel()

		repo := &mockMatcherRepository{skills: []*skilldomain.Skill{}}
		matcher := NewMatcher(repo)

		results, err := matcher.MatchByQuery("test")

		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("matching skill", func(t *testing.T) {
		t.Parallel()

		s, _ := skilldomain.NewSkill("test-skill", "A test skill", "1.0.0")
		repo := &mockMatcherRepository{skills: []*skilldomain.Skill{s}}
		matcher := NewMatcher(repo)

		results, err := matcher.MatchByQuery("test")

		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "test-skill", results[0].Skill.Name)
		assert.Equal(t, 0.5, results[0].Score)
	})

	t.Run("no match", func(t *testing.T) {
		t.Parallel()

		s, _ := skilldomain.NewSkill("different-skill", "A different skill", "1.0.0")
		repo := &mockMatcherRepository{skills: []*skilldomain.Skill{s}}
		matcher := NewMatcher(repo)

		results, err := matcher.MatchByQuery("test")

		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("case insensitive", func(t *testing.T) {
		t.Parallel()

		s, _ := skilldomain.NewSkill("Test-Skill", "A test skill", "1.0.0")
		repo := &mockMatcherRepository{skills: []*skilldomain.Skill{s}}
		matcher := NewMatcher(repo)

		results, err := matcher.MatchByQuery("TEST")

		require.NoError(t, err)
		assert.Len(t, results, 1)
	})
}

func TestMatcher_MatchByContext(t *testing.T) {
	t.Parallel()

	t.Run("nil context falls back to query match", func(t *testing.T) {
		t.Parallel()

		s, _ := skilldomain.NewSkill("test-skill", "A test skill", "1.0.0")
		repo := &mockMatcherRepository{skills: []*skilldomain.Skill{s}}
		matcher := NewMatcher(repo)

		results, err := matcher.MatchByContext("test", nil)

		require.NoError(t, err)
		assert.Len(t, results, 1)
	})

	t.Run("with context uses scoring", func(t *testing.T) {
		t.Parallel()

		s, _ := skilldomain.NewSkill("test-skill", "A test skill", "1.0.0")
		repo := &mockMatcherRepository{skills: []*skilldomain.Skill{s}}
		matcher := NewMatcher(repo)

		ctx := &MatchContext{
			Query:     "test",
			FileTypes: []string{".go"},
		}

		results, err := matcher.MatchByContext("test", ctx)

		require.NoError(t, err)
		assert.NotNil(t, results)
	})
}

func TestMatcher_MatchWithScoring(t *testing.T) {
	t.Parallel()

	t.Run("scores skill by name", func(t *testing.T) {
		t.Parallel()

		s, _ := skilldomain.NewSkill("test-skill", "A test skill", "1.0.0")
		repo := &mockMatcherRepository{skills: []*skilldomain.Skill{s}}
		matcher := NewMatcher(repo)

		ctx := &MatchContext{Query: "test"}

		results, err := matcher.MatchWithScoring("test", ctx)

		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.GreaterOrEqual(t, results[0].Score, 0.5)
		assert.Len(t, results[0].MatchedBy, 1)
		assert.Equal(t, "name", results[0].MatchedBy[0].Type)
	})

	t.Run("scores skill by explicit trigger", func(t *testing.T) {
		t.Parallel()

		s, _ := skilldomain.NewSkill("test-skill", "A test skill", "1.0.0")
		s.AddExplicitTrigger(skilldomain.Trigger{
			Keywords: []string{"implement"},
			Weight:   0.7,
		})
		repo := &mockMatcherRepository{skills: []*skilldomain.Skill{s}}
		matcher := NewMatcher(repo)

		ctx := &MatchContext{Query: "implement feature"}

		results, err := matcher.MatchWithScoring("implement feature", ctx)

		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.True(t, results[0].Score > 0.7)
	})

	t.Run("results sorted by score", func(t *testing.T) {
		t.Parallel()

		s1, _ := skilldomain.NewSkill("low-skill", "Low skill", "1.0.0")
		s2, _ := skilldomain.NewSkill("high-skill", "High skill", "1.0.0")
		s2.AddExplicitTrigger(skilldomain.Trigger{
			Keywords: []string{"implement"},
			Weight:   0.9,
		})

		repo := &mockMatcherRepository{skills: []*skilldomain.Skill{s1, s2}}
		matcher := NewMatcher(repo)

		ctx := &MatchContext{Query: "implement"}

		results, err := matcher.MatchWithScoring("implement", ctx)

		require.NoError(t, err)
		assert.True(t, results[0].Score >= results[len(results)-1].Score)
	})
}

func TestMatcher_MatchByCapability(t *testing.T) {
	t.Parallel()

	t.Run("matches by capability", func(t *testing.T) {
		t.Parallel()

		s, _ := skilldomain.NewSkill("code-skill", "Code skill", "1.0.0")
		s.AddTrigger("code")
		repo := &mockMatcherRepository{skills: []*skilldomain.Skill{s}}
		matcher := NewMatcher(repo)

		results, err := matcher.MatchByCapability(skilldomain.CapabilityTypeCode)

		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "code-skill", results[0].Name)
	})

	t.Run("no match", func(t *testing.T) {
		t.Parallel()

		s, _ := skilldomain.NewSkill("test-skill", "Test skill", "1.0.0")
		repo := &mockMatcherRepository{skills: []*skilldomain.Skill{s}}
		matcher := NewMatcher(repo)

		results, err := matcher.MatchByCapability(skilldomain.CapabilityTypeCode)

		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestMatcher_MatchMultipleCapabilities(t *testing.T) {
	t.Parallel()

	t.Run("matches multiple capabilities", func(t *testing.T) {
		t.Parallel()

		s1, _ := skilldomain.NewSkill("code-skill", "Code skill", "1.0.0")
		s1.AddTrigger("code")

		s2, _ := skilldomain.NewSkill("review-skill", "Review skill", "1.0.0")
		s2.AddTrigger("review")

		s3, _ := skilldomain.NewSkill("test-skill", "Test skill", "1.0.0")
		s3.AddTrigger("code")
		s3.AddTrigger("test")

		repo := &mockMatcherRepository{skills: []*skilldomain.Skill{s1, s2, s3}}
		matcher := NewMatcher(repo)

		results, err := matcher.MatchMultipleCapabilities([]skilldomain.CapabilityType{
			skilldomain.CapabilityTypeCode,
			skilldomain.CapabilityTypeTest,
		})

		require.NoError(t, err)
		assert.Len(t, results, 2)
	})
}

func TestMatcher_MatchToTaskRequirements(t *testing.T) {
	t.Parallel()

	t.Run("matches requirements", func(t *testing.T) {
		t.Parallel()

		s, _ := skilldomain.NewSkill("code-skill", "Code skill", "1.0.0")
		s.AddTrigger("code")
		repo := &mockMatcherRepository{skills: []*skilldomain.Skill{s}}
		matcher := NewMatcher(repo)

		requirements := map[skilldomain.CapabilityType]float64{
			skilldomain.CapabilityTypeCode: 0.8,
		}

		results, err := matcher.MatchToTaskRequirements(requirements)

		require.NoError(t, err)
		assert.Len(t, results, 1)
	})
}

func TestMatcher_ContextBoosts(t *testing.T) {
	t.Parallel()

	t.Run("file type boost", func(t *testing.T) {
		t.Parallel()

		s, _ := skilldomain.NewSkill("test-skill", "A test skill", "1.0.0")
		s.AddExplicitTrigger(skilldomain.Trigger{
			FilePatterns: []string{"*.go"},
			Weight:       0.5,
		})
		repo := &mockMatcherRepository{skills: []*skilldomain.Skill{s}}
		matcher := NewMatcher(repo)

		ctx := &MatchContext{
			Query:     "test",
			FileTypes: []string{".go"},
		}

		results, err := matcher.MatchWithScoring("test", ctx)

		require.NoError(t, err)
		assert.True(t, results[0].Score > 0.5)
	})

	t.Run("task type boost", func(t *testing.T) {
		t.Parallel()

		s, _ := skilldomain.NewSkill("test-skill", "A test skill for implementation", "1.0.0")
		s.AddTrigger("implement")
		repo := &mockMatcherRepository{skills: []*skilldomain.Skill{s}}
		matcher := NewMatcher(repo)

		ctx := &MatchContext{
			Query:    "test",
			TaskType: "implement",
		}

		results, err := matcher.MatchWithScoring("test", ctx)

		require.NoError(t, err)
		assert.True(t, results[0].Score > 0.5)
	})

	t.Run("affinity boost", func(t *testing.T) {
		t.Parallel()

		s, _ := skilldomain.NewSkill("test-skill", "A test skill", "1.0.0")
		repo := &mockMatcherRepository{skills: []*skilldomain.Skill{s}}
		matcher := NewMatcher(repo)

		ctx := &MatchContext{
			Query:        "test",
			ActiveSkills: []string{"test-skill"},
		}

		results, err := matcher.MatchWithScoring("test", ctx)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, results[0].Score, 0.6)
	})
}

func TestMatcher_PatternMatching(t *testing.T) {
	t.Parallel()

	t.Run("matches regex pattern", func(t *testing.T) {
		t.Parallel()

		s, _ := skilldomain.NewSkill("test-skill", "A test skill", "1.0.0")
		s.AddExplicitTrigger(skilldomain.Trigger{
			Patterns: []string{`^implement\s+\w+`},
			Weight:   0.8,
		})
		repo := &mockMatcherRepository{skills: []*skilldomain.Skill{s}}
		matcher := NewMatcher(repo)

		ctx := &MatchContext{Query: "implement feature"}

		results, err := matcher.MatchWithScoring("implement feature", ctx)

		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, 0.8, results[0].Score)
	})
}

func TestMatcher_Delegation(t *testing.T) {
	t.Parallel()

	t.Run("extracts delegations", func(t *testing.T) {
		t.Parallel()

		s, _ := skilldomain.NewSkill("test-skill", "A test skill", "1.0.0")
		s.AddDelegation("other-skill", "reason for delegation")
		repo := &mockMatcherRepository{skills: []*skilldomain.Skill{s}}
		matcher := NewMatcher(repo)

		ctx := &MatchContext{Query: "test"}

		results, err := matcher.MatchWithScoring("test", ctx)

		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Len(t, results[0].Delegations, 1)
		assert.Equal(t, "other-skill", results[0].Delegations[0].ToSkill)
	})
}
