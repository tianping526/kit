package new

import (
	"github.com/spf13/cobra"

	"github.com/tianping526/kit/cmd/kit/internal/project_gen/new/cmd"
	"github.com/tianping526/kit/cmd/kit/internal/project_gen/new/configs"
	"github.com/tianping526/kit/cmd/kit/internal/project_gen/new/trivial"
)

// CmdNew represents the new command.
var CmdNew = &cobra.Command{
	Use:   "new",
	Short: "new project",
	Long:  "Generate new project.",
	Run:   run,
}

func init() {
	CmdNew.AddCommand(cmd.CmdCmd)
	CmdNew.AddCommand(configs.CmdConfigs)
	CmdNew.AddCommand(trivial.CmdTrivial)
}

func run(_ *cobra.Command, _ []string) {
}
