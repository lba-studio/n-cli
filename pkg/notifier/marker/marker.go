package marker

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/lba-studio/n-cli/pkg/notifier"
)

type NotificationMarker interface {
	Done()
}

type NotificationMarkerImpl struct {
	StartedFrom time.Time
	Command     *exec.Cmd
}

func NewNotificationMarker(cmd *exec.Cmd) NotificationMarker {
	return &NotificationMarkerImpl{
		StartedFrom: time.Now(),
		Command:     cmd,
	}
}

func (m *NotificationMarkerImpl) Done() {
	elapsed := time.Since(m.StartedFrom)
	// elapsedMs := elapsed.Milliseconds()
	prettyCommand := strings.Join(m.Command.Args, " ")

	exitCode := m.Command.ProcessState.ExitCode()
	msg := ""
	if exitCode < 0 {
		return
	} else if exitCode == 0 {
		msg = fmt.Sprintf("Command `%s` COMPLETE. Elapsed: %s", prettyCommand, elapsed.String())
	} else {
		msg = fmt.Sprintf("Command `%s` FAILED. Status=%d. Elapsed: %s", prettyCommand, exitCode, elapsed.String())
	}

	err := notifier.Notify(msg)
	if err != nil {
		fmt.Printf("Error encountered when sending notification: %s\n", err.Error())
	}
}
