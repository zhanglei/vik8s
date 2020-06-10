package tools

import (
	"bufio"
	"github.com/ihaiker/vik8s/libs/ssh"
	"github.com/ihaiker/vik8s/libs/utils"
	"os"
	"regexp"
	"strings"
)

var hostsPattern = regexp.MustCompile("([^@\\s]*)@([^:\\s]*):(\\d+(\\.\\d+){3}):(\\d+)")

func HostsNodes(hostsFile string) []*ssh.Node {
	f, err := os.Open(hostsFile)
	utils.Panic(err, "read")
	nodes := make([]*ssh.Node, 0)
	reader := bufio.NewReader(f)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		groups := hostsPattern.FindStringSubmatch(string(line))
		user, pawd, ip, port := groups[1], groups[2], groups[3], groups[5]
		node := &ssh.Node{User: user, Host: ip, Port: port}
		if strings.HasPrefix(pawd, "pk;") {
			node.Type = ssh.PrivateKey
			node.Key = os.ExpandEnv(strings.ReplaceAll(pawd[3:], "~", "$HOME"))
		} else {
			node.Type = ssh.Password
			node.Password = pawd
		}
		utils.Panic(node.Info(), "info")
		nodes = append(nodes, node)
	}
	return nodes
}
