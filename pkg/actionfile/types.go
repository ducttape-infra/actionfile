package actionfile

// Section represents a parsed action section from the Actionfile.
type Section struct {
	Key  string // section key (e.g. "cross-make", "install")
	Body string // shell script body
	Mode string // execution mode hint
}

// ParseOptions controls parsing behavior.
type ParseOptions struct {
	File   string // explicit file path
	SearchDir string // directory to search for Actionfile
}

// ResolveResult is the matched action ready for execution.
type ResolveResult struct {
	Script string // the script to execute
	Mode   string // execution mode
	Shared string // shared script block
	Vars   string // variable assignments
	Config string // config exports
}

// Candidate filenames searched in order.
var CandidateFiles = []string{"Actionfile.md", "Actfile.md", "README.md"}
