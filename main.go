package main

import (
	_"fmt"
	_"flag"
	"os"
	"runtime"

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

var cmd *flags.Parser = flags.NewParser(&opts, flags.PassAfterNonOption | flags.PrintErrors | flags.IgnoreUnknown)

func init() {
}

var Usage = func() {
	cmd.WriteHelp(os.Stderr)
}

func main() {
	runtime.GOMAXPROCS(1)

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
