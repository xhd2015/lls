package run

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/xhd2015/lls/config"
)

func collapseEnv(path string, envs []string) string {
	return config.CollapsePath(path, envs)
}

func expandEnvs(path string, envs []string) string {
	return config.ExpandPath(path, envs)
}

// removeDuplicates removes duplicate entries from a slice
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// selectWithFzf uses fzf to select from a list of options
func selectWithFzf(options []string, query string) (string, error) {
	if len(options) == 0 {
		return "", fmt.Errorf("no options available")
	}

	// Create fzf command
	fzfCmd := exec.Command("fzf", "--no-mouse", "--no-sort")
	if query != "" {
		fzfCmd.Args = append(fzfCmd.Args, "--query="+query)
	}

	// Prepare input
	input := strings.Join(options, "\n")
	fzfCmd.Stdin = strings.NewReader(input)

	// Capture output
	output, err := fzfCmd.Output()
	if err != nil {
		// detect if fzf installed
		if cmdErr, ok := err.(*exec.Error); ok && cmdErr.Err == exec.ErrNotFound {
			return "", fmt.Errorf("fzf not installed, try `brew install fzf`. check https://github.com/junegunn/fzf for more details")
		}

		// Check if it's just user cancellation
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "", nil // User cancelled
		}
		return "", fmt.Errorf("fzf failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// executeCommand executes the selected command
func executeCommand(cmdName string, args []string) error {
	cmd := exec.Command(cmdName, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
