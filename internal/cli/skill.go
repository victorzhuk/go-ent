package cli

import (
	"github.com/spf13/cobra"
	cliskill "github.com/victorzhuk/go-ent/internal/cli/skill"
)

func newSkillCmd() *cobra.Command {
	return cliskill.NewCmd()
}
