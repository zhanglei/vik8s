package sidecars

import (
	"encoding/base64"
	"fmt"
	"github.com/ihaiker/vik8s/certs"
	"github.com/ihaiker/vik8s/install/k8s"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
	"strconv"
)

var dashboardCmd = &cobra.Command{
	Use: "dashboard", Aliases: []string{"db"}, Short: "Web UI (Dashboard) ",
	Long: `Web UI (Dashboard)
more info : https://github.com/kubernetes/dashboard/blob/master/docs/user/README.md
`,
	Run: func(cmd *cobra.Command, args []string) {

		exposePort, _ := cmd.Flags().GetInt("expose")
		ingress, _ := cmd.Flags().GetString("ingress")
		enableInsecureLogin, _ := cmd.Flags().GetBool("enable-insecure-login")
		insecureHeader, _ := cmd.Flags().GetBool("insecure-header")

		master := k8s.Config.Master()

		//dashboard
		{
			data := tools.Json{"ExposePort": exposePort}
			if enableInsecureLogin {
				tools.MustScpAndApplyAssert(master, "yaml/sidecars/dashboard/alternative.yaml", data)
			} else {
				certPath, _ := cmd.Flags().GetString("ingress-cert")
				keyPath, _ := cmd.Flags().GetString("ingress-key")
				if certPath == "" {
					cert, key := certs.NewCertificateAuthority(certs.NewConfig(ingress))
					data["TLSCert"] = base64.StdEncoding.EncodeToString(certs.EncodeCertPEM(cert))
					data["TLSKey"] = base64.StdEncoding.EncodeToString(certs.EncodePrivateKeyPEM(key))
				} else {
					data["TLSCert"] = utils.Base64File(certPath)
					data["TLSKey"] = utils.Base64File(keyPath)
				}
				tools.MustScpAndApplyAssert(master, "yaml/sidecars/dashboard/recommended.yaml", data)
			}
		}

		//dashboard access control
		token := ""
		{
			tools.MustScpAndApplyAssert(master, "yaml/sidecars/dashboard/user.yaml", tools.Json{})

			token = master.MustCmd2String("kubectl -n kubernetes-dashboard describe secret " +
				" $(kubectl -n kubernetes-dashboard get secret | grep admin-user | awk '{print $1}') " +
				" | grep 'token:' | awk '{printf $2}'")
		}

		if ingress != "" {
			data := tools.Json{
				"Ingress": ingress, "Token": token,
				"EnableInsecureLogin": enableInsecureLogin, "InsecureHeader": insecureHeader,
			}
			tools.MustScpAndApplyAssert(master, "yaml/sidecars/dashboard/ingress.yaml", data)
		}

		//show access function
		if exposePort == 0 {
			allocExposePort := master.MustCmd2String("kubectl describe -n kubernetes-dashboard service kubernetes-dashboard  | grep NodePort: | awk '{printf $3}' | awk -F\"/\" '{printf $1}'")
			exposePort, _ = strconv.Atoi(allocExposePort)
		}

		fmt.Println(`Successful installation.`)

		if ingress != "" || exposePort >= 0 {
			fmt.Println("You can access the address via the URL")
			scheme := "https"
			if enableInsecureLogin {
				scheme = "http"
			}
			if exposePort >= 0 {
				fmt.Printf("\t%s://%s:%d\n", scheme, master.Host, exposePort)
			}
			if ingress != "" {
				fmt.Printf("\t%s://%s\n", scheme, ingress)
			}
		} else {
			fmt.Println(`
				To access Dashboard from your local workstation you must create a secure channel to your Kubernetes cluster. Run the following command:
				 $ kubectl proxy
				Now access Dashboard at:
				  http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/.
			`)
		}
		if enableInsecureLogin && !insecureHeader {
			fmt.Println("To make Dashboard use authorization header you simply need to pass Authorization: Bearer <token> in every request to Dashboard.")
			fmt.Println("How to access, please check the documentation help yourself：\n" +
				"   https://github.com/kubernetes/dashboard/blob/master/docs/user/access-control/README.md#authorization-header")
		}
		fmt.Println("token: ")
		fmt.Println(token)
		fmt.Println(`	
	Accessing Dashboard: 
		https://github.com/kubernetes/dashboard/blob/5e86d6d405df3f85fe13938501689c663fdb9fb0/docs/user/accessing-dashboard/README.md`)
	},
}

var dashboardUninstallCmd = &cobra.Command{
	Use: "uninstall", Aliases: []string{"remove", "delete", "del"}, Short: "uninstall dashboard",
	Run: func(cmd *cobra.Command, args []string) {
		master := k8s.Config.Master()
		fmt.Println(master.MustCmd2String("kubectl delete namespaces kubernetes-dashboard"))
	},
}

func init() {
	dashboardCmd.Flags().Bool("enable-insecure-login", false, "When enabled, Dashboard login view will also be shown when Dashboard is not served over HTTPS.")

	/*
		nginx   : https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/
		traefik : https://docs.traefik.io/v1.7/configuration/backends/kubernetes/
	*/
	dashboardCmd.Flags().Bool("insecure-header", false, "Add secure access control token to the header in ingress.")

	dashboardCmd.Flags().Int("expose", -1, "expose dashboard server nodeport. -1: disable, 0: system allocation, >0: designated port")
	dashboardCmd.Flags().String("ingress", "", "deploy dashboard ingress")
	dashboardCmd.Flags().String("ingress-key", "", "dashboard ingress dashboard.key")
	dashboardCmd.Flags().String("ingress-cert", "", "dashboard ingress dashboard.crt")
	dashboardCmd.AddCommand(dashboardUninstallCmd)
}
