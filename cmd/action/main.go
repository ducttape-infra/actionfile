package main

import (
	"fmt"
	"path/filepath"
	"os"
	"strings"

	"actionfile/pkg/actionfile"

	"github.com/spf13/cobra"
)

var version = "1.0.0"

var (
	shell       string
	background  bool
	interactive bool
	subshell    bool
	evaluate    bool
	sourced     bool
	listMode    bool
	listActions bool
	argVars     []string
)

func main() {
	rootCmd := &cobra.Command{
		Version: version,
		Use:   "action [file] [act] [ctx]",
		Short: "Execute actions from an Actionfile",
		Long: `action executes script blocks from Actionfile.md, Actfile.md, or README.md.

  action                    # run "default" action
  action install            # run "install" section
  action setup-cross        # run "setup-cross" section
  action make cross         # run "make" with context "cross"
  action path/to/file.md install  # explicit file
  ./ directory/ install     # search in directory
  action --list             # list all sections
  action --list-actions     # list as "action context" pairs`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var file string
			var searchDir string
			var act, ctx string

			// Parse positional args: [dir] [file] [act] [ctx]
			for _, a := range args {
				if a == "." || strings.HasSuffix(a, "/") {
					searchDir = a
				} else if strings.HasSuffix(a, ".md") {
					file = a
				} else if act == "" {
					act = a
				} else if ctx == "" {
					ctx = a
				}
			}

			// Resolve file
			if file == "" {
				dir := searchDir
				if dir == "" {
					dir = "."
				}
				var err error
				file, err = actionfile.FindFile(dir)
				if err != nil {
					return err
				}
			}
			if fi, err := os.Stat(file); err != nil || fi.IsDir() {
				return fmt.Errorf("file not found: %s", file)
			}

			// Parse
			sections, config, vars, shared, err := actionfile.ExtractSections(file)
			if err != nil {
				return fmt.Errorf("parse: %w", err)
			}

			// List modes
			if listMode {
				for _, k := range actionfile.ListKeys(sections) {
					fmt.Println(k)
				}
				return nil
			}
			if listActions {
				for _, a := range actionfile.ListActions(sections) {
					fmt.Println(a)
				}
				return nil
			}

			// Resolve action
			result, err := actionfile.Resolve(sections, act, ctx)
			if err != nil {
				return err
			}

			// Add --arg overrides to vars
			for _, kv := range argVars {
				if idx := strings.Index(kv, "="); idx > 0 {
					k := kv[:idx]
					v := kv[idx+1:]
					vars += fmt.Sprintf("\n%s=\"%s\"", k, v)
				}
			}

			// Determine mode
			mode := result.Mode
			if background {
				mode = "background"
			} else if interactive {
				mode = "interactive"
			} else if subshell {
				mode = "subshell"
			} else if evaluate {
				mode = "evaluate"
			} else if sourced {
				mode = "sourced"
			}

			// Predefined variables
			vars = fmt.Sprintf("%s\nFILENAME=%q\nDIR=%q", vars, file, filepath.Dir(file))

			// Execute
			return actionfile.Execute(actionfile.ExecuteOpts{
				Dir:         filepath.Dir(file),
				Shell:       shell,
				Background:  background,
				Interactive: interactive,
				Subshell:    true,
				Evaluate:    evaluate,
				Sourced:     sourced,
				Vars:        vars,
				Config:      actionfile.ParseIni(config),
				Shared:      shared,
				Script:      result.Script,
				Mode:        mode,
			})
		},
	}

	rootCmd.Flags().StringVar(&shell, "shell", "bash", "Shell to use for execution")
	rootCmd.Flags().BoolVar(&background, "background", false, "Run in background")
	rootCmd.Flags().BoolVar(&interactive, "interactive", false, "Run interactively")
	rootCmd.Flags().BoolVar(&subshell, "subshell", false, "Run in subshell (default)")
	rootCmd.Flags().BoolVar(&evaluate, "evaluate", false, "Evaluate in current shell")
	rootCmd.Flags().BoolVar(&sourced, "sourced", false, "Source via temporary file")
	rootCmd.Flags().BoolVar(&listMode, "list", false, "List all sections")
	rootCmd.Flags().BoolVar(&listActions, "list-actions", false, "List actions as 'action context' pairs")
	rootCmd.Flags().StringSliceVar(&argVars, "arg", nil, "Override variables (--arg KEY=VAL)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
