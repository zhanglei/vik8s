package ssh

import (
	"fmt"
	"github.com/ihaiker/vik8s/libs/utils"
	"sync"
)

func Sync(nodes []*Node, run func(node *Node)) {
	hasError := ""
	gw := new(sync.WaitGroup)
	for _, node := range nodes {
		gw.Add(1)
		go func(node *Node) {
			defer gw.Done()
			defer utils.Catch(func(err error) {
				hasError += fmt.Sprintf("%s %s\n", node.Host, err.Error())
			})
			run(node)
		}(node)
	}
	gw.Wait()
	utils.Assert(hasError == "", hasError)
}
