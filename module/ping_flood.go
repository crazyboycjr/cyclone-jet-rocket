package module

import (
	"os"
	"log"
	"math/rand"
	_"encoding/binary"

	"cjr/protocol"
	flags "github.com/jessevdk/go-flags"
)

type PingFloodOpt struct {
	BaseOption
	Spoof string `short:"s" long:"spoof" description:"use spoof address" value-name:"address[/mask]" default:""`
	DestFunc func(string) `short:"d" long:"destination" description:"destination address" value-name:"address"`
	CountFunc func(int) `short:"c" long:"count" description:"stop after sending count packets" value-name:"count" default:"0"`
	RateFunc func(string) `short:"r" long:"rate" description:"send packets as a specific rate, such as 100/ms, 2/s, 100/min, the default is \"nolimit\"" value-name:"<speed>" default:"nolimit"`
	TTL uint `short:"t" long:"ttl" description:"set TTL of IP packet" value-name:"ttl" default:"64"`
}

func (p *PingFloodOpt) IsBroadcast() bool {
	return false
}

func pingFloodEntry(stopChan chan int, remainFlags []string) {
	var opts PingFloodOpt

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

	pingFloodStart(stopChan, &opts)
}

func pingFloodStart(stopChan chan int, opts *PingFloodOpt) {
	packetSend(stopChan, pingFloodBuild, opts)
}

func pingFloodBuild(opts_ CommonOption) []protocol.Layer {
	opts := opts_.(*PingFloodOpt)
	srcip := chooseIPv4(opts.Spoof)

	da := make([]byte, 1200, 1200)
	icmp4 := &protocol.ICMPv4 {
		Type: 8, // ICMP echo
		Code: 0,
		Id: 0x2237,
		Seq: 0x3a,
		//Data: []byte{0x1e,0x0a,0x05,0x00,0x00,0x00,0x00,0x00,0x10,0x11,0x12,0x13,0x14,0x15,0x16,0x17,0x18,0x19,0x1a,0x1b,0x1c,0x1d,0x1e,0x1f,0x20,0x21,0x22,0x23,0x24,0x25,0x26,0x27,0x28,0x29,0x2a,0x2b,0x2c,0x2d,0x2e,0x2f,0x30,0x31,0x32,0x33,0x34,0x35,0x36,0x37},
		Data: da,
	}
	ip4 := &protocol.IPv4Packet {
		Version: 4,
		IHL: 5,
		DSCP: 0,
		ECN: 0, // RFC 792 requires TOS = 0
		ID: uint16(rand.Intn(0xffff)),
		DF: true,
		MF: false,
		FragmentOffset: 0,
		TTL: uint8(opts.TTL),
		Protocol: protocol.IPP_ICMP,
		SrcIP: srcip,
		DstIP: opts.Dest(),
	}
	return []protocol.Layer{ip4, icmp4}
}
