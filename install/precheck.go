package install

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
)

func PreCheck(node *ssh.Node) {
	setAliRepo(node)
	checkDistribution(node)
	disableSELinuxAndSwap(node)
}

//经检查系统类型是否满足
func support(node *ssh.Node) {
	ss := []string{
		"CentOS 7",
		"CentOS 8",
	}
	utils.Assert(node.ReleaseName != "unsupport", "unsupport system")
	for _, s := range ss {
		if fmt.Sprintf("%s %s", node.ReleaseName, node.MajorVersion) == s {
			return
		}
	}
	node.Logger("[warn] Unstrictly tested system")
}

func setAliRepo(node *ssh.Node) {
	tools.Installs(node, "epel-release")

	if tools.China {
		repoUrl := fmt.Sprintf("http://mirrors.aliyun.com/repo/Centos-%s.repo", node.MajorVersion)
		node.MustCmd("curl --silent -o /etc/yum.repos.d/CentOS-vik8s.repo " + repoUrl)
		if node.MajorVersion == "7" {
			node.MustCmd("curl --silent -o /etc/yum.repos.d/epel.repo http://mirrors.aliyun.com/repo/epel-7.repo")
		}
	}
	tools.Installs(node, "yum-utils", "lvm2", "device-mapper-persistent-data")
}

func checkDistribution(node *ssh.Node) {
	v1, _ := version.NewVersion("4.1")
	support(node)
	v2, _ := version.NewVersion(node.KernelVersion)
	utils.Assert(v1.LessThanOrEqual(v2), "[%s,%s] The kernel version is too low, please upgrade the kernel first, "+
		"your current version is: %s, the minimum requirement is %s", node.Address(), node.Hostname, v2.String(), v1.String())
}

func disableSELinuxAndSwap(node *ssh.Node) {
	utils.Line("disable SELinux and swap")
	_, _ = node.Cmd("setenforce 0")
	_, _ = node.Cmd("swapoff -a")
	_, _ = node.Cmd("sed -i '/swap/ s$\\/^\\(.*\\)$#\\1$g' /etc/fstab")
}

func InstallNTPServices(node *ssh.Node, timezone string, timeServices ...string) {
	defer func() {
		tools.EnableAndStartService("chronyd", node)
		_, _ = node.Cmd("timedatectl set-ntp true")
	}()

	tools.Install("chrony", "", node)
	node.MustCmd2String(fmt.Sprintf("timedatectl set-timezone %s", timezone))
	config := "allow all\n"
	for _, service := range timeServices {
		config += fmt.Sprintf("server %s iburst\n", service)
	}
	err := node.ScpContent([]byte(config), "/etc/chrony.conf")
	utils.Panic(err, "send ntp config")
}
