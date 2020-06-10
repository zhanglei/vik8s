package storage

import (
	"github.com/gookit/color"
	"github.com/ihaiker/vik8s/install/hosts"
	"github.com/ihaiker/vik8s/install/k8s"
	"github.com/ihaiker/vik8s/install/repo"
	"github.com/ihaiker/vik8s/install/tools"
	"github.com/ihaiker/vik8s/libs/flags"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"github.com/spf13/cobra"
	"os"
	"time"
)

type (
	openEBS struct {
		Repo string `help:"{{quay.io.desc}}"`
		//Version string `help:"the image tag version" def:"1.10.0"`

		ApiServer struct {
			Replicas     int               `def:"1" h:"Number of API Server Replicas"`
			NodeSelector map[string]string `help:"NodeSelecter for API Server"`
			Nodes        []string          `help:"label: sc.vik8s.io/openebs-apiserver=viki8s-openebs-apiserver. see --xxx.nodes"`
		} `flag:"apiserver"`

		Provisioner struct {
			Replicas     int               `def:"1" h:"Number of Provisioner Replicas"`
			NodeSelector map[string]string `help:"NodeSelecter for Provisioner"`
			Nodes        []string          `help:"label: sc.vik8s.io/openebs-provisioner=viki8s-openebs-provisioner. see --xxx.nodes"`
		}

		SnapshotOperator struct {
			Enable       bool              `help:"Enable Snapshot Provisioner" def:"true"`
			Replicas     int               `help:"Number of Snapshot Operator Replicas" def:"1"`
			NodeSelector map[string]string `help:"NodeSelecter for Snapshot Operator"`
			Nodes        []string          `help:"abel: sc.vik8s.io/openebs-snapshot-operator=viki8s-openebs-snapshot-operator. see --xxx.nodes"`
		}

		NDM struct {
			NodeSelector map[string]string `help:"NDM DaemonSet nodeSelector, for example: openebs.io/nodegroup=storage-node"`
			Nodes        []string          `help:"label: sc.vik8s.io/openebs-storage-node=viki8s-openebs-storage. see --xxx.nodes "`

			Sparse struct {
				Size  int64 `help:"Size of the sparse file in bytes" def:"10737418240"`
				Count int   `help:"Number of sparse files to be created" def:"0"`
			}

			Filters struct {
				ExcludeVendors []string `help:"Exclude devices with specified vendor. This parameter is a supplement。（CLOUDBYT,OpenEBS）" def:""`
				ExcludePaths   []string `help:"Exclude devices with specified path patterns. This parameter is a supplement。（loop, /dev/fd0, /dev/sr0, /dev/ram, /dev/dm-, /dev/md）" def:""`
				IncludePaths   []string `help:" Include devices with specified path patterns." def:""`
			}
		}

		NdmOperator struct {
			Replicas     int               `help:"Number of Ndm Operator Replicas" def:"1"`
			NodeSelector map[string]string `help:"NodeSelecter for Ndm Operator"`
			Nodes        []string          `help:"label: sc.vik8s.io/openebs-ndmoperator=viki8s-openebs-ndmoperator. see --xxx.nodes"`
		}

		Admission struct {
			Enable       bool              `def:"true" help:"Enable admission server"`
			Replicas     int               `def:"1" h:"Number of admission server Replicas"`
			NodeSelector map[string]string `help:"NodeSelecter for Local admission"`
			Nodes        []string          `help:"label: sc.vik8s.io/openebs-admission=viki8s-openebs-admission. see --xxx.nodes"`
		}

		LocalPVProvisioner struct {
			Replicas     int               `def:"1" h:"Number of localProvisioner Replicas"`
			NodeSelector map[string]string `help:"NodeSelecter for Local Provisioner"`
			Nodes        []string          `help:"abel: sc.vik8s.io/openebs-localpv-provisioner=viki8s-openebs-localpv-provisioner. see --xxx.nodes"`
		}

		Analytics struct {
			Enabled      bool          `help:"Enable sending stats to Google Analytics" def:"true"`
			PingInterval time.Duration `help:"Duration(hours) between sending ping stat" def:"24h"`
		}

		Jiva struct {
			Replicas int `help:"Number of Jiva Replicas" def:"3"`
		}

		HealthCheck struct {
			InitialDelaySeconds int `help:"Delay before liveness probe is initiated" def:"30"`
			PeriodSeconds       int `help:"How often to perform the liveness probe" def:"60"`
		}

		/*//https://velero.io/docs/v1.0.0/
		Velero struct {
			Enable bool `help:"install velero"`
		}*/
	}
)

func (ebs *openEBS) Name() string {
	return "openebs"
}

func (ebs *openEBS) Description() string {
	return `install OpenEBS StorageClass

Parameter Description:

--xxx.nodes :
Choose a set of nodes, if this parameter is set, the --xxx.nodeselector parameter will be modified to 'xxx.label=xxx.value', 
and this label will be automatically added to these nodes. 

for example: 
	
	--ndm.nodes=10.24.0.10 --ndm.nodes=10.24.0.13-10.24.0.15 --ndm.nodes=hostname1 --ndm.nodes=hostname2

OpenEBS Documentation:
	https://docs.openebs.io/docs/next/overview.html
`
}

func (ebs *openEBS) Flags(cmd *cobra.Command) {
	flags.Flags(cmd.Flags(), ebs, "", repo.Template)
}

func (ebs *openEBS) Apply() {
	ebs.Repo = repo.Suffix(repo.QuayIO(ebs.Repo))

	nodes := hosts.Gets(k8s.Config.Masters, k8s.Config.Nodes)
	ssh.Sync(nodes, func(node *ssh.Node) {
		tools.Install("iscsi-initiator-utils", "", node)
		tools.EnableAndStartService("iscsid", node)
	})

	ebs.NDM.Filters.ExcludeVendors = append([]string{"CLOUDBYT", "OpenEBS"}, ebs.NDM.Filters.ExcludeVendors...)
	ebs.NDM.Filters.ExcludePaths = append([]string{"loop", "/dev/fd0", "/dev/sr0", "/dev/ram", "/dev/dm-", "/dev/md"},
		ebs.NDM.Filters.ExcludePaths...)

	ebs.autoLabelNodes(nodes, &ebs.ApiServer.Nodes, &ebs.ApiServer.NodeSelector,
		"sc.vik8s.io/openebs-apiserver", "viki8s-openebs-apiserver",
	)
	ebs.autoLabelNodes(nodes, &ebs.Provisioner.Nodes, &ebs.Provisioner.NodeSelector,
		"sc.vik8s.io/openebs-provisioner", "viki8s-openebs-provisioner",
	)
	ebs.autoLabelNodes(nodes, &ebs.SnapshotOperator.Nodes, &ebs.SnapshotOperator.NodeSelector,
		"sc.vik8s.io/openebs-snapshot-operator", "viki8s-openebs-snapshot-operator",
	)
	ebs.autoLabelNodes(nodes, &ebs.NDM.Nodes, &ebs.NDM.NodeSelector,
		"sc.vik8s.io/openebs-storage-node", "viki8s-openebs-storage",
	)
	ebs.autoLabelNodes(nodes, &ebs.NdmOperator.Nodes, &ebs.NdmOperator.NodeSelector,
		"sc.vik8s.io/openebs-ndmoperator", "viki8s-openebs-ndmoperator",
	)
	ebs.autoLabelNodes(nodes, &ebs.Admission.Nodes, &ebs.Admission.NodeSelector,
		"sc.vik8s.io/openebs-admission", "viki8s-openebs-admission",
	)
	ebs.autoLabelNodes(nodes, &ebs.LocalPVProvisioner.Nodes, &ebs.LocalPVProvisioner.NodeSelector,
		"sc.vik8s.io/openebs-localpv-provisioner", "viki8s-openebs-localpv-provisioner",
	)
	tools.MustScpAndApplyAssert(nodes[0], "yaml/storage/openebs.yaml", ebs)
}

//根据提供的labelkey,labelvalue和selectNode自动选择节点并打标签，并返回标签和主机hostname
//如果未提供selectNodes 将直接返回，用户默认的 seleectNodes和nodeSelector
//如果指定标签在其他主机上存在将报异常
func (ebs *openEBS) autoLabelNodes(nodes []*ssh.Node, selectNodes *[]string, nodeSelector *map[string]string, labelKey, labelValue string) {
	if len(*selectNodes) == 0 {
		return
	}

	labeledNodes := tools.SearchLabelNode(nodes[0], map[string]string{labelKey: labelValue})
	*selectNodes = tools.SelectNodeNames(nodes, *selectNodes)

	//check label
	for _, labeledNode := range labeledNodes {
		utils.Assert(utils.Search(*selectNodes, labeledNode) >= 0,
			color.FgRed.Render("node %s include label %s=%s \nPlease run the following command to delete the corresponding label: kubectl label nodes %s %s="),
			labeledNode, labelKey, labelValue, labeledNode, labelKey,
		)
	}

	//add label
	for _, selectNode := range *selectNodes {
		if utils.Search(labeledNodes, selectNode) == -1 {
			tools.AddNodeLabel(nodes[0], map[string]string{labelKey: labelValue}, selectNode)
		}
	}

	for k, _ := range *nodeSelector {
		delete(*nodeSelector, k)
	}
	(*nodeSelector)[labelKey] = labelValue
}

func (ebs *openEBS) Delete(data bool) {
	master := hosts.Get(k8s.Config.Masters[0])
	err := master.CmdStd("kubectl delete namespaces openebs", os.Stdout)
	utils.Panic(err, "delete namespaces %s", ebs.Name())
}
