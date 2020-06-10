package cmd

import (
	"github.com/ihaiker/vik8s/install/k8s"
	"github.com/ihaiker/vik8s/sidecars"
	"github.com/spf13/cobra"
)

var sidecarsCmd = &cobra.Command{
	Use: "sidecars", Aliases: []string{"ss"},
	PersistentPreRun: k8s.Config.LoadCmd,
}

func init() {
	sidecars.AddCommand(sidecarsCmd)
}
