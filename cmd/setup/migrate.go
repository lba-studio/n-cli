package setup

import "strings"

func isOldShellHook(command string) bool {
	return strings.Contains(command, "n-cli-notify")
}
