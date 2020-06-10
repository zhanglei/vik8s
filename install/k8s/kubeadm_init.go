package k8s

import (
	"bytes"
	"fmt"
	"github.com/ihaiker/vik8s/install"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	yamls "github.com/ihaiker/vik8s/yaml"
	"io/ioutil"
	"os"
	"text/template"
)

func InitCluster(node *ssh.Node) *ssh.Node {
	utils.Line("init kubernetes cluster %s", node.Host)

	preCheck(node)
	//fix 这里需要先行加入，因为在初始化模板文件中需要使用
	Config.JoinNode(true, node.Host)

	node.Logger("init cluster")
	{
		setHosts(node, node.Host, Config.Kubernetes.ApiServer)
		install.InstallNTPServices(node, Config.Timezone, Config.NTPServices...)
		makeCerts(node)
		initKubernetes(node)
		applyApiServerEndpoint(node)
	}
	return node
}

func ResetNode(node *ssh.Node) {
	_, _ = node.Cmd("kubeadm reset -f")
	Config.RemoveNode(node.Host)
}

func setHosts(node *ssh.Node, ip, domain string) {
	_ = node.MustCmd2String(fmt.Sprintf("sed -i /%s/d /etc/hosts", domain))
	_ = node.MustCmd2String(fmt.Sprintf("echo '%s %s' >> /etc/hosts", ip, domain))
}

func scpKubeConfig(node *ssh.Node) string {
	config := string(yamls.MustAsset("yaml/kubeadm-config.yaml"))

	if Config.Kubernetes.KubeadmConfig != "" {
		configBytes, err := ioutil.ReadFile(Config.Kubernetes.KubeadmConfig)
		utils.Panic(err, "read kubeadm-config file %s", Config.Kubernetes.KubeadmConfig)
		config = string(configBytes)
	}

	remote := node.Vik8s("apply/kubeadm-config.yaml")
	node.Logger("scp kubeadm.yaml %s", remote)

	kubeadmConfig := parseTemplate(config)
	err := node.ScpContent(kubeadmConfig, remote)
	utils.Panic(err, "scp kubeadm-config file")
	return remote
}

func parseTemplate(templateFile string) []byte {
	data := tools.Json{
		"Etcd":    Config.ETCD,
		"Masters": hosts.Gets(Config.Masters), "Workers": hosts.Gets(Config.Nodes),
		"Nodes":   hosts.Gets(Config.Masters, Config.Nodes),
		"Kubeadm": Config.Kubernetes,
	}
	out := bytes.NewBufferString("")
	t, err := template.New("").Parse(templateFile)
	utils.Panic(err, "template file error")
	err = t.Execute(out, data)
	utils.Panic(err, "template file error")
	return out.Bytes()
}

func initKubernetes(node *ssh.Node) {
	remote := scpKubeConfig(node)
	err := node.CmdStd(fmt.Sprintf("kubeadm init --config=%s --upload-certs", remote), os.Stdout)
	utils.Panic(err, "kubeadm init")
	copyKubeletConfg(node)
}

func copyKubeletConfg(node *ssh.Node) {
	kubeDir := node.HomeDir(".kube")
	kubeConfig := node.HomeDir(".kube/config")
	_ = node.MustCmd2String(fmt.Sprintf("mkdir -p %s  && cp -f /etc/kubernetes/admin.conf %s", kubeDir, kubeConfig))
}

func applyApiServerEndpoint(node *ssh.Node) {
	name := "yaml/vik8s-api-server.yaml"
	content := parseTemplate(string(yamls.MustAsset(name)))

	apiServerEndpoint := node.Vik8s("apply/vik8s-api-server.yaml")
	err := node.ScpContent(content, apiServerEndpoint)
	utils.Panic(err, "scp api-server-endpoint.yaml file")

	_, err = node.Cmd("kubectl apply -f " + apiServerEndpoint)
	utils.Panic(err, "kubectl apply -f api-server-endpoint.yaml")
}
