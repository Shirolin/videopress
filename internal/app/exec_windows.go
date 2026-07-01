package app

import (
	"os"
	"os/exec"
)

func execLookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func runCommand(name string, args []string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
