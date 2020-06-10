package k8s

import (
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"os"
)

func Clean(nodes []*ssh.Node, expFn ...func(node *ssh.Node)) {
	ssh.Sync(nodes, func(node *ssh.Node) {

		_, _ = node.Cmd("kubeadm reset -f")
		_, _ = node.Cmd("ipvsadm -C")
		_, _ = node.Cmd("ifconfig | grep cni0 > /dev/null && ifconfig cni0 down")
		_, _ = node.Cmd("ip link show | grep kube-ipvs0 && ip link delete kube-ipvs0")
		_, _ = node.Cmd("ip link show | grep dummy0 && ip link delete dummy0")

		_, _ = node.Cmd("rm -rf /etc/cni/net.d/* ~/.kube /etc/kubernetes/*")
		_, _ = node.Cmd("rm -rf /var/lib/etcd")
		_, _ = node.Cmd("rm -rf /var/lib/ceph")

		_, _ = node.Shell(`
			iptables -P INPUT ACCEPT
			iptables -P FORWARD ACCEPT
			iptables -P OUTPUT ACCEPT
			iptables -t nat -F
			iptables -t mangle -F
			iptables -F
			iptables -X
			
			ip6tables -P INPUT ACCEPT
			ip6tables -P FORWARD ACCEPT
			ip6tables -P OUTPUT ACCEPT
			ip6tables -t nat -F
			ip6tables -t mangle -F
			ip6tables -F
			ip6tables -X
		`)

		for _, fn := range expFn {
			fn(node)
		}
	})
	config := tools.Join("config.json")
	utils.Panic(os.RemoveAll(config), "remove %s", config)

	kube := tools.Join("kube")
	utils.Panic(os.RemoveAll(kube), "remove %s", kube)
}
