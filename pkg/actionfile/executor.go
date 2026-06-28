package actionfile

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type ExecuteOpts struct {
	Dir         string
	Shell       string
	Background  bool
	Interactive bool
	Subshell    bool
	Evaluate    bool
	Sourced     bool
	Vars        string
	Config      string
	Shared      string
	Script      string
	Mode        string
}

func Execute(opts ExecuteOpts) error {
	if opts.Shell == "" {
		opts.Shell = "bash"
	}
	var pre string
	if opts.Dir != "" {
		pre += fmt.Sprintf("cd %q\n", opts.Dir)
	}
	if opts.Config != "" {
		pre += opts.Config + "\n"
	}
	if opts.Vars != "" {
		pre += opts.Vars + "\n"
	}
	if opts.Shared != "" {
		pre += opts.Shared + "\n"
	}
	script := pre + opts.Script
	mode := opts.Mode
	if mode == "" {
		mode = "subshell"
	}
	switch {
	case opts.Background:
		return runShell(opts.Shell, script, true)
	case opts.Interactive:
		return runInteractive(opts.Shell, script)
	case opts.Evaluate:
		return runShell("sh", script, false)
	case opts.Sourced:
		return runSourced(opts.Shell, script)
	default:
		return runShell(opts.Shell, script, false)
	}
}

func runShell(shell, script string, bg bool) error {
	cmd := exec.Command(shell, "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if bg {
		return cmd.Start()
	}
	return cmd.Run()
}

func runInteractive(shell, script string) error {
	cmd := exec.Command(shell, "-i")
	cmd.Stdin = strings.NewReader(script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runSourced(shell, script string) error {
	f, err := os.CreateTemp("", "actionfile-*.sh")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(script); err != nil {
		f.Close()
		return fmt.Errorf("write temp: %w", err)
	}
	f.Close()
	cmd := exec.Command(shell, "-c", "source "+f.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
