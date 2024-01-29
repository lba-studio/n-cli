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

	if m.Command.ProcessState.ExitCode() < 0 {
		return
	}

	msg := fmt.Sprintf("Command `%s` is complete. Elapsed: %s", strings.Join(m.Command.Args, " "), elapsed.String())
	err := notifier.Notify(msg)
	if err != nil {
		fmt.Printf("Error encountered when sending notification: %s\n", err.Error())
	}
}
