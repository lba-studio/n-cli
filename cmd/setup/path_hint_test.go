package setup

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathHint(t *testing.T) {
	hint := pathHint()
	if runtime.GOOS == "windows" {
		assert.Contains(t, hint, "System Environment Variables")
	} else {
		assert.Contains(t, hint, "~/.zshrc")
	}
}
