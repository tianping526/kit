package proto

import (
	"github.com/spf13/cobra"

	"github.com/tianping526/kit/cmd/kit/internal/project_gen/proto/add"
	"github.com/tianping526/kit/cmd/kit/internal/project_gen/proto/server"
)

// CmdProto represents the proto command.
var CmdProto = &cobra.Command{
	Use:   "proto",
	Short: "Generate the proto files",
	Long:  "Generate the proto files.",
	Run:   run,
}

func init() {
	CmdProto.AddCommand(add.CmdAdd)
	CmdProto.AddCommand(server.CmdServer)
}

func run(_ *cobra.Command, _ []string) {
}
