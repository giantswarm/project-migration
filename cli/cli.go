package cli

import (
	"flag"
	"fmt"
	"os"
)

// Options holds the command-line flag values.
type Options struct {
	Project  string
	DryRun   bool
	Type     string
	Name     string
	Area     string
	Function string
}

// ParseFlags parses and validates command-line flags.
func ParseFlags() Options {
	projectFlag := flag.String("p", "", "Project Number (eg 301)")
	dryRun := flag.Bool("d", false, "Dry run mode")
	typeFlag := flag.String("t", "", "Type (team, sig, wg)")
	nameFlag := flag.String("n", "", "Name of Team, SIG or WG")
	areaFlag := flag.String("a", "", "Area (eg KaaS)")
	functionFlag := flag.String("f", "", "Function (eg 'Product Strategy')")
	flag.Parse()

	if *projectFlag == "" {
		fmt.Println("Project number is missing. Exiting")
		os.Exit(1)
	}
	if *typeFlag == "" {
		fmt.Println("Type is missing. Exiting")
		os.Exit(1)
	}
	if *nameFlag == "" {
		fmt.Println("Name is missing. Exiting")
		os.Exit(1)
	}
	if *typeFlag != "team" && *typeFlag != "sig" && *typeFlag != "wg" {
		fmt.Println("Type must be either team, sig or wg. Exiting")
		os.Exit(1)
	}

	return Options{
		Project:  *projectFlag,
		DryRun:   *dryRun,
		Type:     *typeFlag,
		Name:     *nameFlag,
		Area:     *areaFlag,
		Function: *functionFlag,
	}
}
