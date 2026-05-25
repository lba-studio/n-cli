package setup

import "runtime"

func pathHint() string {
	if runtime.GOOS == "windows" {
		return "Add n-cli to your PATH via System Environment Variables or your PowerShell profile."
	}
	return "Check your shell profile (e.g. ~/.zshrc) and restart the agent after setup."
}
