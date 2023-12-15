package pkgmanager

import "os/exec"

func hasCommand(cmdName string) bool {
	cmd := exec.Command("powershell", "/C", "Get-Command "+cmdName+" -ea SilentlyContinue")
	if err := cmd.Run(); err != nil {
		return false
	}

	return cmd.ProcessState.ExitCode() == 0
}
