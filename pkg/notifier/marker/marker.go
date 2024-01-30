package marker

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/lba-studio/n-cli/pkg/monitor"
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

	cpuTime := time.Duration(time.Duration(monitor.GetCPU()) * time.Nanosecond)

	exitCode := m.Command.ProcessState.ExitCode()
	msg := ""
	if exitCode < 0 {
		return
	} else if exitCode == 0 {
		msg = fmt.Sprintf("Command `%s` COMPLETE.\nElapsed: %s\nCPU Time: %s", prettyCommand, elapsed.String(), cpuTime.String())
	} else {
		msg = fmt.Sprintf("Command `%s` FAILED. Status=%d.\nElapsed: %s\nCPU Time: %s", prettyCommand, exitCode, elapsed.String(), cpuTime.String())
	}

	err := notifier.Notify(msg)
	if err != nil {
		fmt.Printf("Error encountered when sending notification: %s\n", err.Error())
	}
}
