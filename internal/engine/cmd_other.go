//go:build !windows

package engine

import (
	"os/exec"
)

func prepareCmd(cmd *exec.Cmd) {
	// No-op for non-Windows platforms
}
