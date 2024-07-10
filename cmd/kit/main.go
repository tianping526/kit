package main

import (
	"log"

	"github.com/spf13/cobra"

	bts "github.com/tianping526/kit/cmd/kit/internal/bts_gen"
	"github.com/tianping526/kit/cmd/kit/internal/project_gen/new"
	"github.com/tianping526/kit/cmd/kit/internal/project_gen/proto"
	rc "github.com/tianping526/kit/cmd/kit/internal/rc_gen"
)

var rootCmd = &cobra.Command{
	Use:     "kit",
	Short:   "kit: An elegant toolkit for Go microservices.",
	Long:    `kit: An elegant toolkit for Go microservices.`,
	Version: release,
}

func init() {
	rootCmd.AddCommand(proto.CmdProto)
	rootCmd.AddCommand(new.CmdNew)
	rootCmd.AddCommand(bts.BtsCmd)
	rootCmd.AddCommand(rc.RcCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
