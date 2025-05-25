package Handler

import (
	"os"
	"os/exec"
	"path/filepath"
)

func executeFile(fileName string) error {
	dir, _ := os.Getwd()
	cmd := exec.Command(filepath.Join(dir, fileName))
	err := cmd.Start()
	if err != nil {
		return err
	}

	return nil
}
