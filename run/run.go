package run

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/less-gen/flags"
	"github.com/xhd2015/lls/config"
	"github.com/xhd2015/xgo/support/cmd"
	"golang.org/x/term"
)

// Global debug flag
var debug bool

const help = `
lls helps to quickly navigate through your working projects

Usage: lls                         navgiate through fzf
       lls show                    show predefined project locations
	   lls edit                    edit predefined project locations

Options:
  --config CONFIG_FILE             path to config file (default: ~/.config/lls/config.json)
  --editor EDITOR                  selected editor, default: code(VSCode), vim
  --list                           list all available directories without fzf selection
  -v,--verbose                     show verbose info  

Examples:
  lls
  lls edit
  lls --config /path/to/custom/config.json
  lls --list
`

func Main(args []string) error {
	if len(args) > 0 {
		cmd := args[0]
		cmdArgs := args[1:]

		if cmd == "help" || cmd == "--help" {
			fmt.Print(strings.TrimPrefix(help, "\n"))
			return nil
		}

		switch cmd {
		case "show":
			return handleShow(cmdArgs)
		case "history":
			return handleHistory(cmdArgs)
		case "edit", "config":
			return handleEdit(cmdArgs)
		default:
			if !strings.HasPrefix(cmd, "-") {
				return fmt.Errorf("unrecognized command: %s", cmd)
			}
			// fallback to default
		}
	}

	return lls(args)
}

// debugLog prints debug messages when debug mode is enabled
func debugLog(format string, args ...interface{}) {
	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}

func handleShow(args []string) error {
	var configFile string

	args, err := flags.String("--config", &configFile).
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

	fmt.Println(conf)
	return nil
}

func handleHistory(args []string) error {
	var configFile string
	var stdout bool

	args, err := flags.String("--config", &configFile).
		Bool("--stdout", &stdout).
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

	var cfg config.Config
	content, _ := os.ReadFile(conf)
	json.Unmarshal(content, &cfg)

	if len(cfg.HistoryFiles) == 0 {
		return fmt.Errorf("history_files is empty")
	}

	var lines []string
	for _, file := range cfg.HistoryFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		lines = append(lines, splitLines(string(content))...)
	}

	lines = unique(lines)

	isTerminal := term.IsTerminal(int(os.Stdout.Fd()))
	if stdout || !isTerminal {
		for _, line := range lines {
			fmt.Println(line)
		}
		return nil
	}

	tmpFile, err := os.CreateTemp("", "lls-history-*.txt")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(strings.Join(lines, "\n"))
	tmpFile.Close()

	return cmd.New().Stdin(os.Stdin).Run("bash", "-c", fmt.Sprintf("cat '%s' | fzf", tmpFile.Name()))
}

func unique(lines []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if !seen[line] {
			seen[line] = true
			result = append(result, line)
		}
	}
	return result
}

func splitLines(content string) []string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		s := strings.TrimSpace(line)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

func handleEdit(args []string) error {
	var configFile string

	args, err := flags.String("--config", &configFile).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		return fmt.Errorf("unexpected extra args: %s", strings.Join(args, " "))
	}

	var conf string

	if configFile != "" {
		// Use custom config file
		conf = configFile
		// Ensure the directory exists for custom config file
		confDir := filepath.Dir(conf)
		if err := os.MkdirAll(confDir, 0755); err != nil {
			return err
		}
	} else {
		// Use default config file
		conf, err = getConfigFile(true, "config.json")
		if err != nil {
			return err
		}
	}

	var cfg config.Config
	content, _ := os.ReadFile(conf)
	json.Unmarshal(content, &cfg)

	stat, statErr := os.Stat(conf)
	if statErr != nil {
		if !os.IsNotExist(statErr) {
			return statErr
		}
		simpleConf := config.Config{
			Envs:     []string{},
			Projects: []string{},
		}
		jsonData, err := json.Marshal(simpleConf)
		if err != nil {
			return err
		}
		err = os.WriteFile(conf, jsonData, 0644)
		if err != nil {
			return err
		}
	} else if stat.IsDir() {
		return fmt.Errorf("config file is a directory: %s", conf)
	}

	openCmd := "code"
	if cfg.OpenCmd != "" {
		openCmd = cfg.OpenCmd
	}

	// prefer code, vim
	return cmd.New().Stdin(os.Stdin).Run(openCmd, conf)
}

func getConfigFile(createDir bool, fileName string) (string, error) {
	conf, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(conf, "lls")
	if createDir {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return "", err
		}
	}
	return filepath.Join(configDir, fileName), nil
}
