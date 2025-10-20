package helper

import (
	"os"
	"strings"
)

func ReplaceTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, _ := os.UserHomeDir()
		path = strings.Replace(path, "~", homeDir, 1)
	}

	return path
}
