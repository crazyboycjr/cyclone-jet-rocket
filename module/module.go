package module

import (
	"fmt"
	"time"
	"math/rand"
)

type moduleMeta struct {
	name string
	desc string
	entry func(chan int, []string) error
}
var moduleMap map[string] moduleMeta

func ListModule() {
	for _, m := range moduleMap {
		fmt.Printf("%s: %s\n", m.name, m.desc)
	}
}

func LoadModule(stopChan chan int, name string, flags []string) error {
	m, ok := moduleMap[name]
	if !ok {
		return fmt.Errorf("module %s do not exists", name)
	}
	return m.entry(stopChan, flags)
}

func init() {
	rand.Seed(time.Now().UnixNano())
	moduleMap = map[string] moduleMeta {
		"ping-flood": moduleMeta {
			name: "ping-flood",
			desc: `A ping flood is a simple DoS attack where attacker overwhelms the victim with ICMP "echo request"(ping) packets`,
			entry: pingFloodEntry,
		},
		"smurf": moduleMeta {
			name: "smurf",
			desc: `Braodcast victim's spoofed source IP address`,
			entry: smurfEntry,
		},
		"udp-flood": moduleMeta {
			name: "udp-flood",
			desc: `A UDP flood attack is a DoS attack using User Datagram Protocol`,
			entry: udpFloodEntry,
		},
		"syn-flood": moduleMeta {
			name: "syn-flood",
			desc: `A SYN flood is a form of DoS attack in which sending a succession of SYN requests to a target's system to consume server resources`,
			entry: synFloodEntry,
		},
		"slowloris": moduleMeta {
			name: "slowloris",
			desc: `aka slow HTTP attack`,
			entry: slowlorisEntry,
		},
		"rdns": moduleMeta {
			name: "rdns",
			desc: `DNS amplification is a type of reflection attack`,
			entry: rdnsEntry,
		},
		"land": moduleMeta {
			name: "land",
			desc: `LAND (Local Area Network Denial) attack use the destination ip addresss and port as the source address and port`,
			entry: landEntry,
		},
		"http-flood": moduleMeta {
			name: "http-flood",
			desc: `HTTP request flood`,
			entry: httpFloodEntry,
		},
	}
}
