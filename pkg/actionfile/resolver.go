package actionfile

import "fmt"

// Resolve finds the matching section for the given act and ctx.
// Resolution order:
//  1. act+ctx (concatenated)
//  2. ctx-act (dashed)
//  3. act (direct)
//  4. "default" if act is empty
func Resolve(sections []Section, act, ctx string) (*ResolveResult, error) {
	if act == "" {
		act = "default"
	}

	bodyMap := make(map[string]Section)
	for _, s := range sections {
		bodyMap[s.Key] = s
	}

	script := ""
	var mode string

	// Step 1: Concatenated: act+ctx
	if ctx != "" {
		joined := act + ctx
		if s, ok := bodyMap[joined]; ok {
			script = s.Body
			mode = s.Mode
		}
	}

	// Step 2: Dashed: ctx-act
	if script == "" && ctx != "" {
		composite := ctx + "-" + act
		if s, ok := bodyMap[composite]; ok {
			script = s.Body
			mode = s.Mode
		}
	}

	// Step 3: Direct
	if script == "" {
		if s, ok := bodyMap[act]; ok {
			script = s.Body
			mode = s.Mode
		}
	}

	if script == "" {
		return nil, fmt.Errorf("section %q not found", act)
	}

	// Default execution mode
	if mode == "" {
		mode = "subshell"
	}

	return &ResolveResult{
		Script: script,
		Mode:   mode,
	}, nil
}
