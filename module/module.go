package module

import (
	"fmt"
	"log"
)

type moduleMeta struct {
	name string
	desc string
	entry func([]string)
}
var moduleMap map[string] moduleMeta

func ListModule() {
	for _, m := range moduleMap {
		fmt.Printf("%s: %s\n", m.name, m.desc)
	}
}

func LoadModule(name string, flags []string) {
	m, ok := moduleMap[name]
	if !ok {
		log.Fatalf("module %s do not exists\n", name)
	}
	m.entry(flags)
}

func init() {
	moduleMap = map[string] moduleMeta {
		"ping-flood": moduleMeta {
			name: "ping-flood",
			desc: `A ping flood is a simple DoS attack where attacker overwhelms the victim with ICMP "echo request"(ping) packets`,
			entry: pingFloodEntry,
		},
	}
}
