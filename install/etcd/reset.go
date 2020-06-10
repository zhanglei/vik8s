package etcd

import (
	"fmt"
	"github.com/gookit/color"
	"github.com/ihaiker/vik8s/libs/ssh"
)

func ResetCluster(node *ssh.Node) {
	if !Config.Exists(node.Host) {
		fmt.Printf("%s not in the cluster\n", color.FgRed.Render(node.Host))
		return
	}
	_ = node.MustCmd2String("etcdadm reset")
	Config.Remove(node.Host)
}
