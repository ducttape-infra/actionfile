package actionfile

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ExecuteOpts controls execution behavior.
type ExecuteOpts struct {
	Shell       string // shell to use (default: bash)
	Background  bool   // run in background
	Interactive bool   // run interactively
	Subshell    bool   // run in subshell (default)
	Evaluate    bool   // eval in current shell
	Sourced     bool   // source via tempfile
	Vars        string // variable assignments (newline-separated)
	Config      string // config exports
	Shared      string // shared script
	Script      string // action script
	Mode        string // execution mode hint
}

// Execute runs the action script according to the options.
func Execute(opts ExecuteOpts) error {
	if opts.Shell == "" {
		opts.Shell = "bash"
	}

	// Build preamble: config + vars + shared
	var preamble string
	if opts.Config != "" {
		preamble += opts.Config + "\n"
	}
	if opts.Vars != "" {
		preamble += opts.Vars + "\n"
	}
	if opts.Shared != "" {
		preamble += opts.Shared + "\n"
	}
	fullScript := preamble + opts.Script

	// Mode from opts, fallback to Mode field
	mode := opts.Mode
	if mode == "" {
		mode = "subshell"
	}

	switch {
	case opts.Background:
		return executeBackground(opts.Shell, fullScript)
	case opts.Interactive:
		return executeInteractive(opts.Shell, fullScript)
	case opts.Evaluate:
		return executeEvaluate(fullScript)
	case opts.Sourced:
		return executeSourced(opts.Shell, fullScript)
	default: // subshell
		return executeSubshell(opts.Shell, fullScript)
	}
}

func executeSubshell(shell, script string) error {
	cmd := exec.Command(shell, "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func executeBackground(shell, script string) error {
	cmd := exec.Command(shell, "-c", script)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}

func executeInteractive(shell, script string) error {
	cmd := exec.Command(shell, "-i")
	cmd.Stdin = strings.NewReader(script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func executeEvaluate(script string) error {
	// eval in current process (limited — Go can't modify parent shell)
	// Run via sh -c as a subshell equivalent
	cmd := exec.Command("sh", "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func executeSourced(shell, script string) error {
	// Write to tempfile and source it
	tmpFile, err := os.CreateTemp("", "actionfile-*.sh")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(script); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write temp: %w", err)
	}
	tmpFile.Close()

	cmd := exec.Command(shell, "-c", "source "+tmpFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
