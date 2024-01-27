package version_test

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/lba-studio/n-cli/pkg/version"
	"github.com/stretchr/testify/assert"
)

func TestGetVersion(t *testing.T) {
	if os.Getenv("VERSION_TEST") != "true" {
		t.Skip()
	}
	gitTagCmd := exec.Command("git", "describe", "--tags")
	var gitTagOut bytes.Buffer
	gitTagCmd.Stdout = &gitTagOut
	if err := gitTagCmd.Run(); err != nil {
		t.Errorf("cannot get git tag: %s", err.Error())
		return
	}
	gitTag := strings.TrimSpace(gitTagOut.String())
	assert.NotEqual(t, "", gitTag, "git tag is empty")
	assert.Equal(t, gitTag, version.GetVersion())
}
