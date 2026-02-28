package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	skilldomain "github.com/victorzhuk/go-ent/internal/skill/domain"
)

func TestInMemoryRepository_Save(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc    string
		skill   *skilldomain.Skill
		wantErr bool
		errType error
	}{
		{
			desc: "valid skill",
			skill: func() *skilldomain.Skill {
				s, _ := skilldomain.NewSkill("test-skill", "A test skill", "1.0.0")
				return s
			}(),
			wantErr: false,
		},
		{
			desc:    "nil skill",
			skill:   nil,
			wantErr: true,
		},
		{
			desc: "skill with empty ID",
			skill: func() *skilldomain.Skill {
				s, _ := skilldomain.NewSkill("test-skill", "A test skill", "1.0.0")
				s.ID = ""
				return s
			}(),
			wantErr: true,
		},
		{
			desc: "skill with empty name",
			skill: func() *skilldomain.Skill {
				s, _ := skilldomain.NewSkill("", "A test skill", "1.0.0")
				return s
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			t.Parallel()

			repo := NewInMemoryRepository()
			err := repo.Save(tt.skill)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			require.NoError(t, err)
		})
	}

	t.Run("duplicate skill by ID", func(t *testing.T) {
		t.Parallel()

		repo := NewInMemoryRepository()
		s1, _ := skilldomain.NewSkill("skill-1", "First skill", "1.0.0")
		s1.ID = "duplicate-id"
		require.NoError(t, repo.Save(s1))

		s2, _ := skilldomain.NewSkill("skill-2", "Second skill", "2.0.0")
		s2.ID = "duplicate-id"
		err := repo.Save(s2)

		require.Error(t, err)
		assert.ErrorIs(t, err, skilldomain.ErrDuplicate)
	})

	t.Run("duplicate skill by name", func(t *testing.T) {
		t.Parallel()

		repo := NewInMemoryRepository()
		s1, _ := skilldomain.NewSkill("test-skill", "First skill", "1.0.0")
		require.NoError(t, repo.Save(s1))

		s2, _ := skilldomain.NewSkill("test-skill", "Second skill", "2.0.0")
		err := repo.Save(s2)

		require.Error(t, err)
		assert.ErrorIs(t, err, skilldomain.ErrDuplicate)
	})
}

func TestInMemoryRepository_FindByID(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()

		repo := NewInMemoryRepository()
		s, _ := skilldomain.NewSkill("test-skill", "A test skill", "1.0.0")
		require.NoError(t, repo.Save(s))

		skill, err := repo.FindByID("test-skill@1.0.0")

		require.NoError(t, err)
		assert.NotNil(t, skill)
		assert.Equal(t, "test-skill@1.0.0", skill.ID)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		repo := NewInMemoryRepository()
		_, err := repo.FindByID("nonexistent@1.0.0")

		require.Error(t, err)
		assert.ErrorIs(t, err, skilldomain.ErrNotFound)
	})

	t.Run("empty ID", func(t *testing.T) {
		t.Parallel()

		repo := NewInMemoryRepository()
		_, err := repo.FindByID("")

		require.Error(t, err)
		assert.ErrorIs(t, err, skilldomain.ErrNotFound)
	})
}

func TestInMemoryRepository_FindByName(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()

		repo := NewInMemoryRepository()
		s, _ := skilldomain.NewSkill("test-skill", "A test skill", "1.0.0")
		require.NoError(t, repo.Save(s))

		skill, err := repo.FindByName("test-skill")

		require.NoError(t, err)
		assert.NotNil(t, skill)
		assert.Equal(t, "test-skill", skill.Name)
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		repo := NewInMemoryRepository()
		_, err := repo.FindByName("nonexistent")

		require.Error(t, err)
		assert.ErrorIs(t, err, skilldomain.ErrNotFound)
	})

	t.Run("empty name", func(t *testing.T) {
		t.Parallel()

		repo := NewInMemoryRepository()
		_, err := repo.FindByName("")

		require.Error(t, err)
		assert.ErrorIs(t, err, skilldomain.ErrNotFound)
	})
}

func TestInMemoryRepository_ListAll(t *testing.T) {
	t.Parallel()

	t.Run("empty repository", func(t *testing.T) {
		t.Parallel()

		repo := NewInMemoryRepository()
		skills, err := repo.ListAll()

		require.NoError(t, err)
		assert.NotNil(t, skills)
		assert.Empty(t, skills)
	})

	t.Run("multiple skills", func(t *testing.T) {
		t.Parallel()

		repo := NewInMemoryRepository()

		s1, _ := skilldomain.NewSkill("skill-1", "First skill", "1.0.0")
		s2, _ := skilldomain.NewSkill("skill-2", "Second skill", "2.0.0")
		s3, _ := skilldomain.NewSkill("skill-3", "Third skill", "3.0.0")

		require.NoError(t, repo.Save(s1))
		require.NoError(t, repo.Save(s2))
		require.NoError(t, repo.Save(s3))

		skills, err := repo.ListAll()

		require.NoError(t, err)
		assert.Len(t, skills, 3)
	})
}

func TestInMemoryRepository_Delete(t *testing.T) {
	t.Parallel()

	t.Run("delete existing", func(t *testing.T) {
		t.Parallel()

		repo := NewInMemoryRepository()
		s, _ := skilldomain.NewSkill("test-skill", "A test skill", "1.0.0")
		require.NoError(t, repo.Save(s))

		err := repo.Delete("test-skill@1.0.0")

		require.NoError(t, err)

		_, err = repo.FindByID("test-skill@1.0.0")
		assert.Error(t, err)
		assert.ErrorIs(t, err, skilldomain.ErrNotFound)
	})

	t.Run("delete non-existent", func(t *testing.T) {
		t.Parallel()

		repo := NewInMemoryRepository()
		err := repo.Delete("nonexistent@1.0.0")

		require.Error(t, err)
		assert.ErrorIs(t, err, skilldomain.ErrNotFound)
	})

	t.Run("delete empty ID", func(t *testing.T) {
		t.Parallel()

		repo := NewInMemoryRepository()
		err := repo.Delete("")

		require.Error(t, err)
		assert.ErrorIs(t, err, skilldomain.ErrNotFound)
	})
}

func TestInMemoryRepository_Update(t *testing.T) {
	t.Parallel()

	t.Run("update existing", func(t *testing.T) {
		t.Parallel()

		repo := NewInMemoryRepository()
		s, _ := skilldomain.NewSkill("test-skill", "A test skill", "1.0.0")
		require.NoError(t, repo.Save(s))

		s.Description = "Updated description"
		err := repo.Update(s)

		require.NoError(t, err)

		skill, err := repo.FindByID("test-skill@1.0.0")
		require.NoError(t, err)
		assert.Equal(t, "Updated description", skill.Description)
	})

	t.Run("update non-existent", func(t *testing.T) {
		t.Parallel()

		repo := NewInMemoryRepository()
		s, _ := skilldomain.NewSkill("new-skill", "New description", "1.0.0")

		err := repo.Update(s)

		require.Error(t, err)
		assert.ErrorIs(t, err, skilldomain.ErrNotFound)
	})

	t.Run("update with nil skill", func(t *testing.T) {
		t.Parallel()

		repo := NewInMemoryRepository()
		err := repo.Update(nil)

		require.Error(t, err)
	})

	t.Run("update with empty ID", func(t *testing.T) {
		t.Parallel()

		repo := NewInMemoryRepository()
		s, _ := skilldomain.NewSkill("test-skill", "A test skill", "1.0.0")
		s.ID = ""

		err := repo.Update(s)

		require.Error(t, err)
	})
}

func TestInMemoryRepository_Exists(t *testing.T) {
	t.Parallel()

	t.Run("exists", func(t *testing.T) {
		t.Parallel()

		repo := NewInMemoryRepository()
		s, _ := skilldomain.NewSkill("test-skill", "A test skill", "1.0.0")
		require.NoError(t, repo.Save(s))

		assert.True(t, repo.Exists("test-skill"))
		assert.False(t, repo.Exists("nonexistent"))
	})

	t.Run("empty name", func(t *testing.T) {
		t.Parallel()

		repo := NewInMemoryRepository()
		assert.False(t, repo.Exists(""))
	})
}

func TestInMemoryRepository_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping concurrent test in short mode")
	}

	repo := NewInMemoryRepository()
	done := make(chan bool, 10)

	for i := 0; i < 5; i++ {
		go func(idx int) {
			s, _ := skilldomain.NewSkill(
				"skill-"+string(rune('a'+idx)),
				"A test skill",
				"1.0.0",
			)
			_ = repo.Save(s)
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		go func() {
			_, _ = repo.ListAll()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
