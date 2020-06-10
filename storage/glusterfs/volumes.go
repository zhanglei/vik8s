package glusterfs

import (
	"fmt"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/k8s"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var glusterFSCreateVolumeCmd = &cobra.Command{
	Use: "create",
	Example: "vik8s storage glusterfs volume create volumeName " +
		"[stripe <COUNT>] [[replica <COUNT> [arbiter <COUNT>]]|[replica 2 thin-arbiter 1]] [disperse [<COUNT>]] [disperse-data <COUNT>] [redundancy <COUNT>]",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
			return err
		}
		return cobra.OnlyValidArgs(cmd, args[1:])
	},
	ValidArgs: []string{
		"stripe", "replica", "arbiter", "thin-arbiter",
		"disperse", "disperse-data", "redundancy",
		"2", "3", "4", "5",
	},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		nodes := hosts.Gets(k8s.Config.Masters, k8s.Config.Nodes)
		master := nodes.Master()
		volumes := strings.Split(master.MustCmd2String(gluster+" volume list", true), "\n")
		if utils.Search(volumes, name) >= 0 {
			fmt.Println("the volume already exists")
			return
		}

		podNodes := strings.Split(master.MustCmd2String("kubectl get pod -l glusterfs=pod -n glusterfs -o jsonpath={.items[*].spec.nodeName}", true), " ")
		volumePath := master.MustCmd2String(`kubectl get daemonsets.apps glusterfs -o go-template='{{range .spec.template.spec.volumes}}{{if eq .name "vik8s-glusterfs-volumes"}}{{.hostPath.path}}{{end}}{{end}}'`, true)
		for _, nodeName := range podNodes {
			nodes.Get(nodeName).MustCmdStd(fmt.Sprintf("mkdir -p %s/%s", volumePath, name), os.Stdout, false)
		}

		options := strings.Join(args[1:], " ")

		brick := ""
		for _, node := range podNodes {
			brick += fmt.Sprintf(" %s:/data/%s", node, name)
		}
		master.MustCmdStd(fmt.Sprintf("%s volume create %s %s transport tcp %s force", gluster, name, options, brick), os.Stdout)
		master.MustCmdStd(fmt.Sprintf("%s volume start %s", gluster, name), os.Stdout)
	},
}
