package cli

import (
	"flag"
	"fmt"
)

// Config holds the parsed command-line flags.
type Config struct {
	Project  string
	DryRun   bool
	Type     string
	Name     string
	Area     string
	Function string
	Verbose  bool // new verbose flag
}

// Parse parses and returns the command-line flags.
func Parse() *Config {
	project := flag.String("p", "", "Project Number (eg 301)")
	dryRun := flag.Bool("d", false, "Dry run")
	typ := flag.String("t", "", "Type (eg 'team, sig, wg')")
	name := flag.String("n", "", "Name of Team, SIG or WG (eg Rocket)")
	area := flag.String("a", "", "Area (eg KaaS)")
	functionF := flag.String("f", "", "Function (eg 'Product Strategy')")
	verboseF := flag.Bool("v", false, "Enable verbose output") // new flag

	flag.Usage = func() {
		fmt.Println("Usage:")
		flag.PrintDefaults()
	}

	flag.Parse()

	return &Config{
		Project:  *project,
		DryRun:   *dryRun,
		Type:     *typ,
		Name:     *name,
		Area:     *area,
		Function: *functionF,
		Verbose:  *verboseF,
	}
}
