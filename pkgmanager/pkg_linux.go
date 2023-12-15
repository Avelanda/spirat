package pkgmanager

import "os/exec"

func hasCommand(cmdName string) bool {
	cmd := exec.Command("sh", "-c", "command -v "+cmdName)
	if err := cmd.Run(); err != nil {
		return false
	}

	return cmd.ProcessState.ExitCode() == 0
}
