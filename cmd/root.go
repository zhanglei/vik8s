package cmd

import (
	"fmt"
	"github.com/gookit/color"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
	"os"
	"runtime"
	"strings"
)

var (
	VERSION        = "v0.0.0"
	BUILD_TIME     = "2012-12-12 12:12:12"
	GITLOG_VERSION = "0000"
)

var rootCmd = &cobra.Command{
	Version: strings.TrimSpace(VERSION),
	Use:     "vik8s", Short: "very easy install HA k8s",
	Long: fmt.Sprintf(`very easy install k8s。
Build: %s, Go: %s, GitLog: %s`,
		BUILD_TIME, runtime.Version(), GITLOG_VERSION),
}

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates bash completion scripts",
	Run: func(cmd *cobra.Command, args []string) {
		_ = rootCmd.GenBashCompletion(os.Stdout)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&tools.ConfigDir, "config", "f",
		tools.ConfigDir, "The folder where the configuration file is located")
	rootCmd.PersistentFlags().BoolVar(&tools.China, "china", true, "Whether domestic network")

	rootCmd.AddCommand(dataCmd, hostsCmd, etcdCmd)
	rootCmd.AddCommand(configCmd, initCmd, joinCmd, resetCmd, cleanCmd)
	rootCmd.AddCommand(ingressRootCmd, storageClassCmd, sidecarsCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(bashCmd)
	rootCmd.Flags().SortFlags = false
}

func Execute() {
	defer utils.Catch(func(err error) {
		if serr, match := err.(*utils.WrapError); match {
			color.FgRed.Println(serr.Error())
		} else {
			color.FgRed.Println(err)
			color.FgRed.Println(utils.Stack())
		}
	})

	if runCommand, args, err := rootCmd.Find(os.Args[1:]); err == nil {
		if runCommand.Name() == rootCmd.Name() {
			for i, arg := range args {
				if arg == "--" {
					param := strings.Join(args[i+1:], " ")
					os.Args = append(os.Args[0:i+1], "bash", param)
					break
				}
			}
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
