package run

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/lls/config"
)

func lls(args []string) error {
	var configFile string
	var listOnly bool

	args, err := flags.String("--config", &configFile).
		Bool("--list", &listOnly).
		Bool("-v,--verbose", &debug).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}
	var conf string

	if configFile != "" {
		// Use custom config file
		conf = configFile
	} else {
		// Use default config file
		conf, err = getConfigFile(false, "config.json")
		if err != nil {
			return err
		}
	}

	confData, readErr := os.ReadFile(conf)
	if readErr != nil {
		if !os.IsNotExist(readErr) {
			return readErr
		}
	}

	var cfg config.Config
	if len(confData) > 0 {
		err = json.Unmarshal(confData, &cfg)
		if err != nil {
			return fmt.Errorf("reading config %s: %w", conf, err)
		}
	}

	// Generate list of directories similar to lls_list
	var allDirs []string

	// Process base directories (equivalent to base array in bash)
	for _, dir := range cfg.Projects {
		dirs := llsWorktreeAt(dir, cfg.Envs)
		allDirs = append(allDirs, dirs...)
	}

	// Add worktrees from current directory (equivalent to lls_pwd)
	pwdDirs := llsPwd(cfg.Envs)
	allDirs = append(allDirs, pwdDirs...)

	// Remove duplicates (equivalent to kool lines uniq)
	uniqueDirs := removeDuplicates(allDirs)

	var collapsedDirs []string
	for _, uniqDir := range uniqueDirs {
		dir := collapseEnv(uniqDir, cfg.Envs)
		collapsedDirs = append(collapsedDirs, dir)
	}

	// If --list flag is specified, just print the directories and exit
	if listOnly {
		for _, dir := range collapsedDirs {
			fmt.Println(dir)
		}
		return nil
	}

	// Use fzf for selection
	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	selectedCmd, err := selectWithFzf(collapsedDirs, query)
	if err != nil {
		return err
	}

	if selectedCmd == "" {
		return nil // User cancelled
	}

	expandedCmd := expandEnvs(selectedCmd, cfg.Envs)

	// Execute the selected command
	fmt.Println()
	fmt.Printf("%s -> %s\n", selectedCmd, expandedCmd)

	openCmd := "code"
	if cfg.OpenCmd != "" {
		openCmd = cfg.OpenCmd
	}

	return executeCommand(openCmd, []string{expandedCmd})
}

// llsWorktreeAt replicates the lls_worktree_at function
func llsWorktreeAt(dir string, envs []string) []string {
	expandedDir := expandEnvs(dir, envs)
	if expandedDir == "" {
		return nil
	}

	gitDir := filepath.Join(expandedDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		// Not a git repository, return the directory itself
		return []string{expandedDir}
	}

	// Get git worktree list
	cmd := exec.Command("git", "-C", expandedDir, "worktree", "list")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to the directory itself
		return []string{expandedDir}
	}

	var result []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) > 0 {
			worktreePath := parts[0]
			result = append(result, worktreePath)
		}
	}

	return result
}

// llsWorktree replicates the lls_worktree function
func llsWorktree(dir string, envs []string) []string {
	worktreeDir := filepath.Join(dir, "worktree")
	if _, err := os.Stat(worktreeDir); os.IsNotExist(err) {
		return nil
	}

	var result []string
	err := filepath.Walk(worktreeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		if info.IsDir() && info.Name() == ".git" {
			// Found a .git directory, get the parent directory
			parentDir := filepath.Dir(path)
			result = append(result, parentDir)
		}

		return nil
	})

	if err != nil {
		return nil
	}

	return result
}

// llsPwd replicates the lls_pwd function
func llsPwd(envs []string) []string {

	cmd := exec.Command("git", "worktree", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var result []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) > 0 {
			worktreePath := parts[0]
			result = append(result, worktreePath)
		}
	}

	return result
}
