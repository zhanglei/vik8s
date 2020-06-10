package tools

import (
	"os"
	"path/filepath"
)

type Json map[string]interface{}

var ConfigDir = os.ExpandEnv("$HOME/.vik8s")
var China = true

func Join(path ...string) string {
	return filepath.Join(append([]string{ConfigDir}, path...)...)
}
