package k8s

import (
	"encoding/json"
	"fmt"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
)

func daemonJson() map[string]interface{} {
	daemon := map[string]interface{}{
		"exec-opts": []string{"native.cgroupdriver=systemd"},
		"hosts": []string{
			"fd://",
			"tcp://0.0.0.0:2375",
			"unix:///var/run/docker.sock",
		},
		"registry-mirrors": []string{},
	}
	if tools.China {
		daemon["registry-mirrors"] = append(daemon["registry-mirrors"].([]string), []string{
			"https://dockerhub.azk8s.cn",
			"https://docker.mirrors.ustc.edu.cn",
			"http://hub-mirror.c.163.com",
			"https://registry.cn-hangzhou.aliyuncs.com",
		}...)
	}
	if Config.Docker.Registry != "" {
		daemon["insecure-registries"] = []string{Config.Docker.Registry}
		daemon["registry-mirrors"] = append(daemon["registry-mirrors"].([]string), Config.Docker.Registry)
	}
	return daemon
}

func checkDocker(node *ssh.Node) {
	defer tools.EnableAndStartService("docker", node)

	dockerVersion := node.MustCmd2String("rpm -qi docker-ce | grep Version | awk '{printf $3}'")
	if dockerVersion != "" && (dockerVersion == Config.Docker.Version || !Config.Docker.CheckVersion) {
		node.Logger("docker installd %s", dockerVersion)
		return
	}
	var err error
	//Install containerd.io
	if node.ReleaseName == "CentOS" && node.MajorVersion == "8" {
		node.Logger("CentOS 8 check container.io")
		containerIO := node.MustCmd2String("rpm -qa | grep containerd.io || echo NOT_FOUND")
		if containerIO == "NOT_FOUND" {
			node.Logger("Install containerd.io")
			_, err = node.Cmd(fmt.Sprintf("dnf clean packages && dnf Install -y %s", repo.Containerd()))
			utils.Panic(err, "Install containerd.io")
		}
	}

	tools.AddRepo(repo.Docker(), node)
	tools.Install("docker-ce", Config.Docker.Version, node)
	tools.Install("docker-ce-cli", Config.Docker.Version, node)

	//set docker daemon.json
	_ = node.MustCmd2String("mkdir -p /etc/docker")
	if Config.Docker.DaemonJson != "" {
		err := node.Scp(Config.Docker.DaemonJson, "/etc/docker/daemin.json")
		utils.Panic(err, "scp daemon.json")
	} else {
		bs, _ := json.MarshalIndent(daemonJson(), "", "    ")
		err := node.ScpContent(bs, "/etc/docker/daemon.json")
		utils.Panic(err, "scp daemon.json")
	}

	// set docker.service
	_ = node.MustCmd2String("sed -i 's/-H fd:\\/\\///g' /usr/lib/systemd/system/docker.service ")
}
