package marker

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/lba-studio/n-cli/pkg/formatter"
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
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Panic encountered while processing process information. Skipping analytics.", r)
		}
	}()
	elapsed := time.Since(m.StartedFrom)
	// elapsedMs := elapsed.Milliseconds()
	prettyCommand := strings.Join(m.Command.Args, " ")
	cpuTimeNano, err := monitor.GetCPU()
	if err != nil {
		fmt.Printf("Cannot get cpuTimeNano: %s\n", err.Error())
		return
	}
	cpuTime := time.Duration(time.Duration(cpuTimeNano) * time.Nanosecond)

	memoryUsageInt64, err := monitor.GetMemoryFromCmd(m.Command)
	if err != nil {
		fmt.Printf("Cannot get memoryUsage: %s\n", err.Error())
		return
	}
	memoryUsage := formatter.PrettyPrintInt64(memoryUsageInt64)
	// memoryUsage := fmt.Sprintf("%d", memoryUsageInt64)

	exitCode := m.Command.ProcessState.ExitCode()
	msg := ""
	if exitCode < 0 {
		return
	} else if exitCode == 0 {
		msg = fmt.Sprintf("Command `%s` COMPLETE.\nElapsed: %s\nCPU Time: %s\nMaxrss: %s", prettyCommand, elapsed.String(), cpuTime.String(), memoryUsage)
	} else {
		msg = fmt.Sprintf("Command `%s` FAILED. Status=%d.\nElapsed: %s\nCPU Time: %s\nMaxrss: %s", prettyCommand, exitCode, elapsed.String(), cpuTime.String(), memoryUsage)
	}

	err = notifier.Notify(msg)
	if err != nil {
		fmt.Printf("Error encountered when sending notification: %s\n", err.Error())
	}
}
