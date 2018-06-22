package tmux

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Result is a commadn result
type Result struct {
	Stdout string
	Stderr string
}

// Exec runs a tmux command
func Exec(args ...string) (Result, error) {
	var stdin bytes.Buffer
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	cmd := exec.Command("tmux", args...)
	cmd.Stdin = &stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return Result{}, fmt.Errorf("Error running command \"tmux %v\", %s", args, stderr.String())
	}

	return Result{stdout.String(), stderr.String()}, nil
}

// SessionInfo infos about a running tmux session
type SessionInfo struct{}

// ListSessions returns the list of sessions currently running
func ListSessions() (map[string]SessionInfo, error) {
	sessionMap := make(map[string]SessionInfo)

	res, err := Exec("ls")
	if err != nil {
		return sessionMap, nil
	}

	splits := strings.Split(res.Stdout, "\n")
	for _, sess := range splits {
		sessSplits := strings.Split(sess, ":")
		if len(sessSplits) > 1 {
			sessionMap[sessSplits[0]] = SessionInfo{}
		}
	}

	return sessionMap, nil
}

type Options struct {
	BaseIndex     int
	PaneBaseIndex int
}

func GetOptions() (*Options, error) {
	options := &Options{
		BaseIndex:     0,
		PaneBaseIndex: 0,
	}

	var stderr bytes.Buffer
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "tmux start-server\\; show-options -g\\; show-window-options -g")
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		return options, fmt.Errorf("Error gettings tmux options %v, %s", err, stderr.String())
	}

	optionsString := strings.Split(stdout.String(), "\n")
	for _, option := range optionsString {
		optionSplits := strings.Split(option, " ")
		if len(optionSplits) == 2 {
			name := optionSplits[0]
			if name == "base-index" {
				if v, err := strconv.Atoi(optionSplits[1]); err == nil {
					options.BaseIndex = v
				}
			} else if name == "pane-base-index" {
				if v, err := strconv.Atoi(optionSplits[1]); err == nil {
					options.PaneBaseIndex = v
				}

			}
		}
	}

	return options, nil
}
