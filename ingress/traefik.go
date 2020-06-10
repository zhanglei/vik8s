package ingress

import (
	"encoding/base64"
	"fmt"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
	"math/rand"
	"os"
	"strings"
)

type traefik struct {
	Repo                        string
	Version                     string
	Hostnetwork                 bool
	NodePortHttp, NodePortHttps int
	Replicas                    int
	NodeSelectors               map[string]string
	IngressUI                   string
	AuthUI                      bool
	AuthUser                    string
	AuthPassword                string
}

func (t *traefik) Name() string {
	return "traefik"
}

func (t *traefik) Description() string {
	return "https://docs.traefik.io/v1.7/user-guide/kubernetes/"
}

func (t *traefik) Flags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&t.Repo, "repo", "", "Choose a container registry to images (default docker.io)")

	cmd.Flags().StringVar(&t.Version, "version", "1.7", "")

	cmd.Flags().BoolVar(&t.Hostnetwork, "host-network", false, "Whether to enable the host network method")

	cmd.Flags().IntVar(&t.NodePortHttp, "nodeport", -1, "the ingress-traefik http service nodeport, 0: automatic allocatiot,  -1: disable")
	cmd.Flags().IntVar(&t.NodePortHttps, "nodeport-https", -1, "the ingress-traefik https 443 service nodeport, 0: automatic allocation, -1: disable")

	cmd.Flags().StringToStringVar(&t.NodeSelectors, "node-selector", map[string]string{"kubernetes.io/os": "linux"}, "Deployment.nodeSelector")
	cmd.Flags().IntVar(&t.Replicas, "replicas", 1, "ingress-traefik pod replicas number")
	cmd.Flags().StringVar(&t.IngressUI, "ui-ingress", "", "Creating ingress that will expose the Traefik Web UI.")

	cmd.Flags().BoolVar(&t.AuthUI, "ui-auth", true, "Whether to enable `basic authentication` int traefik web ui ingress")
	cmd.Flags().StringVar(&t.AuthUser, "ui-user", "admin", "web ui `basic authentication` user ")
	cmd.Flags().StringVar(&t.AuthPassword, "ui-passwd", "", "web ui `basic authentication` password (default: randomly generated and pint to console)")
}

func (t *traefik) Apply(master *ssh.Node) {
	data := tools.Json{
		"Repo": t.Repo, "Version": t.Version,
		"HostNetwork":  t.Hostnetwork,
		"NodePortHttp": t.NodePortHttp, "NodePortHttps": t.NodePortHttps,
		"Replicas": t.Replicas, "NodeSelectors": t.NodeSelectors,
		"IngressUI": t.IngressUI, "AuthUI": t.AuthUI,
	}
	if t.Repo != "" && !strings.HasSuffix(t.Repo, "/") {
		data["Repo"] = t.Repo + "/"
	}

	if t.AuthUI {
		if t.AuthPassword == "" {
			t.AuthPassword = fmt.Sprintf("%06d", rand.Int63n(1000000))
			fmt.Println(strings.Repeat("=", 40))
			fmt.Printf("  the web ui ingress default password for `%s` is : %s\n", t.AuthUser, t.AuthPassword)
			fmt.Println(strings.Repeat("=", 40))
		}
		password, _ := utils.HashApr1(t.AuthPassword)
		data["AuthData"] = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", t.AuthUser, password)))
	}

	name := "yaml/ingress/traefik.yaml"
	tools.MustScpAndApplyAssert(master, name, data)
}

func (n *traefik) Delete(master *ssh.Node) {
	err := master.CmdStd("kubectl delete namespaces ingress-traefik ", os.Stdout)
	utils.Panic(err, "delete traefik ingress")
}
