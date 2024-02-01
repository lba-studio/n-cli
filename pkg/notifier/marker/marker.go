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

type printedMarkerInfo struct {
	exitCode    int
	elapsed     string
	cpuTime     string
	memoryUsage int64
}

func (m *NotificationMarkerImpl) formatMessage(info printedMarkerInfo) string {
	status := "COMPLETE"
	if info.exitCode > 0 {
		status = "FAILED"
	}
	prettyCommand := strings.Join(m.Command.Args, " ")

	infoStrings := []string{
		fmt.Sprintf("Command `%s` %s.", prettyCommand, status),
		fmt.Sprintf("Elapsed: %s", info.elapsed),
	}
	if !monitor.IsWindows() {
		infoStrings = append(infoStrings,
			fmt.Sprintf("CPU Time: %s", info.cpuTime),
			fmt.Sprintf("Memory Usage: %s", formatter.PrettyPrintInt64(info.memoryUsage)),
		)
	}

	return strings.Join(infoStrings, "\n")
}

func (m *NotificationMarkerImpl) Done() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Panic encountered while processing process information. Skipping analytics.", r)
		}
	}()
	elapsed := time.Since(m.StartedFrom)
	cpuTimeNano, err := monitor.GetCPU()
	if err != nil && err != monitor.ErrIsWindows {
		fmt.Printf("Cannot get cpuTimeNano: %s\n", err.Error())
		return
	}
	cpuTime := time.Duration(time.Duration(cpuTimeNano) * time.Nanosecond)

	memoryUsage, err := monitor.GetMemoryFromCmd(m.Command)
	if err != nil && err != monitor.ErrIsWindows {
		fmt.Printf("Cannot get memoryUsage: %s\n", err.Error())
		return
	}

	exitCode := m.Command.ProcessState.ExitCode()
	if exitCode < 0 {
		return
	}
	msg := m.formatMessage(printedMarkerInfo{
		memoryUsage: memoryUsage,
		cpuTime:     cpuTime.String(),
		elapsed:     elapsed.String(),
		exitCode:    exitCode,
	})

	err = notifier.Notify(msg)
	if err != nil {
		fmt.Printf("Error encountered when sending notification: %s\n", err.Error())
	}
}
