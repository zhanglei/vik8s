package cmd

import (
	"fmt"
	"github.com/carmark/pseudo-terminal-go/terminal"
	"github.com/gookit/color"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/kvz/logstreamer"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
)

var colors = []color.Color{
	color.FgRed, color.FgGreen, color.FgYellow,
	color.FgBlue, color.FgMagenta, color.FgCyan,
}
var colorsSize = len(colors)

func filterNode(node *ssh.Node, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	for _, filter := range filters {
		if filter == node.Hostname {
			return true
		}
	}
	return false
}

func runCmd(cmd string, nodes []*ssh.Node, filters ...string) {
	for i, node := range nodes {
		if !filterNode(node, filters) {
			continue
		}
		prefix := colors[i%colorsSize].Sprintf("[%s] ", node.Hostname)
		out := logstreamer.NewLogstreamerForStdout(prefix)
		if err := node.CmdStd(cmd, out, true); err != nil {
			_, _ = out.Write([]byte(err.Error()))
		}
		_ = out.Close()
	}
}

var bashCmd = &cobra.Command{
	Use: "bash", Short: "Run commands uniformly in the cluster",
	SilenceErrors: true, SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		nodes := hosts.Nodes()
		utils.Assert(len(nodes) > 0, "not found any host, use `vik8s host <node>` to add.")

		if len(args) > 0 {
			runCmd(strings.Join(args, " "), nodes)
			return
		}

		filters := make([]string, 0)

		if term, err := terminal.NewWithStdInOut(); err != nil {
			fmt.Println(err.Error())
		} else {
			defer term.ReleaseFromStdInOut()
			term.SetPrompt("vik8s> ")
		ILP:
			for {
				line, err := term.ReadLine()
				if err == io.EOF {
					break
				} else if (err != nil && strings.Contains(err.Error(), "control-c break")) || len(line) == 0 {
					continue
				}
				line = strings.TrimSpace(line)

				if line[0] == '@' {
					filters = strings.Split(strings.TrimSpace(line[1:]), " ")
					term.SetPrompt(fmt.Sprintf("vik8s %s> ", strings.Join(filters, "|")))
					continue
				} else if line == "-" {
					filters = filters[0:0]
					term.SetPrompt("vik8s> ")
					continue
				}

				switch line {
				case "":
				case "clear":
					_, _ = os.Stdout.Write([]byte("\x1b[2J\x1b[0;0H"))
				case "exit", "quit":
					break ILP
				default:
					runCmd(line, nodes, filters...)
				}
			}
		}
	},
}
