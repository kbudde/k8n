package ytt

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

// Render the input with the ytt files in the folder using ytt cmd.
func Render(input, folder string) ([]byte, error) {
	yttFiles, err := filepath.Glob(filepath.Join(folder, "*.ytt.yaml"))
	if err != nil {
		return nil, err
	}

	if len(yttFiles) == 0 {
		return nil, fmt.Errorf("no ytt files found in %s", folder)
	}

	args := []string{"--data-values-file", input}
	for _, f := range yttFiles {
		args = append(args, "-f", f)
	}
	args = append(args, "--allow-symlink-destination", "/")

	return runYtt(args)
}

// Run ytt with the given args.
func runYtt(args []string) ([]byte, error) {
	cmd := exec.Command("ytt", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("error running ytt: %w", err)
	}

	return output, nil
}
