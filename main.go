package main

import (
	_"fmt"
	_"flag"
	"os"

	flags "github.com/jessevdk/go-flags"
	modules "cjr/module"
)

var (
	module string
	listModule bool
	help bool
)

var opts struct {
	Module string `short:"m" long:"module" description:"specify a module to use" value-name:"module" default:""`
	ListModule bool `short:"l" long:"list-module" description:"list available modules"`
	Help bool `short:"h" long:"help" description:"print this help"`
}

/*
var (
	module *string = flag.String("module", "", "specify a module to use")
	listModule *bool = flag.Bool("list-module", false, "list available modules")
	help *bool = flag.Bool("help", false, "print this help")
)
*/

//var cmd *flag.FlagSet = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
var cmd *flags.Parser = flags.NewParser(&opts, flags.PassAfterNonOption | flags.PrintErrors | flags.IgnoreUnknown)

func init() {
	/*
	cmd.StringVar(&module, "module", "", "specify a `module` to use")
	cmd.StringVar(&module, "m", "", "specify a `module` to use")
	cmd.BoolVar(&listModule, "list-module", false, "list available modules")
	cmd.BoolVar(&listModule, "lm", false, "list available modules")
	cmd.BoolVar(&help, "help", false, "print this help")
	cmd.BoolVar(&help, "h", false, "print this help")
	*/
}

var Usage = func() {
	cmd.WriteHelp(os.Stderr)
}

func main() {

	args, err := cmd.ParseArgs(os.Args[1:])

	if err != nil {
		panic(err)
	}
	if opts.Help {
		Usage()
		os.Exit(0)
	}

	if opts.ListModule {
		modules.ListModule()
		os.Exit(0)
	}

	if len(opts.Module) > 0 {
		modules.LoadModule(opts.Module, args)
		os.Exit(0)
	}

	//Usage()
}
