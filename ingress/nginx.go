package ingress

import (
	"fmt"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
	"os"
)

const (
	NginxIngressVersion = "0.30.0"
)

type nginx struct {
	Repo          string
	Version       string
	Hostnetwork   bool
	NodePortHttp  int
	NodePortHttps int
	Replicas      int

	NodeSelectors map[string]string
}

func (n *nginx) Name() string {
	return "nginx"
}

func (n *nginx) Description() string {
	return fmt.Sprintf("install kubernetes/ingress-nginx ( v%s ), more info see https://github.com/kubernetes/ingress-nginx", NginxIngressVersion)
}

func (n *nginx) Flags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&n.Repo, "repo", "", repo.QuayIODesc())
	cmd.Flags().StringVar(&n.Version, "version", NginxIngressVersion, "")
	cmd.Flags().BoolVar(&n.Hostnetwork, "host-network", false,
		"Whether to enable the host network method")

	cmd.Flags().IntVar(&n.NodePortHttp, "nodeport", -1, "the ingress-nginx http 80 service nodeport, 0: automatic allocation, -1: disable")
	cmd.Flags().IntVar(&n.NodePortHttps, "nodeport-https", -1, "the ingress-nginx https 443 service nodeport, 0: automatic allocation, -1: disable")

	cmd.Flags().StringToStringVar(&n.NodeSelectors, "node-selector", map[string]string{"kubernetes.io/os": "linux"}, "Deployment.nodeSelector")
	cmd.Flags().IntVar(&n.Replicas, "replicas", 1, "ingress-nginx pod replicas number")
}

func (n *nginx) Apply(master *ssh.Node) {
	data := tools.Json{
		"Repo": repo.QuayIO(n.Repo), "Version": n.Version,
		"HostNetwork":  n.Hostnetwork,
		"NodePortHttp": n.NodePortHttp, "NodePortHttps": n.NodePortHttps,
		"Replicas": n.Replicas, "NodeSelectors": n.NodeSelectors,
	}
	name := "yaml/ingress/nginx.yaml"
	tools.MustScpAndApplyAssert(master, name, data)
}

func (n *nginx) Delete(master *ssh.Node) {
	err := master.CmdStd("kubectl delete namespaces ingress-nginx ", os.Stdout)
	utils.Panic(err, "delete nginx ingress")
}
