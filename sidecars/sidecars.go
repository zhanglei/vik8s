package sidecars

import "github.com/spf13/cobra"

//Metrics
//从 v1.8 开始，资源使用情况的度量（如容器的 CPU 和内存使用）可以通过 Metrics API 获取。注意

//Prometheus
//Prometheus 是另外一个监控和时间序列数据库，并且还提供了告警的功能。它提供了强大的查询语言和HTTP接口，也支持将数据导出到Grafana中展示。

// kubesphere
// https://kubesphere.io/zh-CN/

//helm

func AddCommand(sd *cobra.Command) {
	sd.AddCommand(dashboardCmd)
}
