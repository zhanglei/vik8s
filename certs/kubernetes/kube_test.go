package kubecerts

import (
	"github.com/ihaiker/vik8s/install/tools"
	"testing"
	"time"
)

func TestCreatePKIAssets(t *testing.T) {
	node := Node{
		Name:                "vm11",
		Host:                "10.24.1.11",
		ApiServer:           "vik8s-api-server",
		SvcCIDR:             "10.96.0.0/12",
		CertificateValidity: time.Hour * 24 * 365 * 10,
	}
	dir := tools.Join("kube/pki2")
	CreatePKIAssets(dir, node)
}
