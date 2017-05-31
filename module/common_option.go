package module

import (
	"log"
	"net"
	"time"
)

type CommonOption interface {
	Count() uint
	SetCount(uint)
	Rate() time.Duration
	SetRate(time.Duration)
	Dest() net.IP
	SetDest(net.IP)
	IsBroadcast() bool
}

func commonRateFunc(opts CommonOption, rate string) {
	var wait time.Duration
	if rate == "nolimit" {
		wait = time.Duration(0)
	} else {
		wait = parseRateOrDie(rate)
	}
	opts.SetRate(wait)
}

func commonDestFunc(opts CommonOption, dest string) {
	opts.SetDest(net.ParseIP(dest))
	if opts.Dest() == nil {
		log.Fatal("parse destination IP error")
	}
}

func commonCountFunc(opts CommonOption, count int) {
	if count <= 0 {
		opts.SetCount(^uint(0) >> 1)
	} else {
		opts.SetCount(uint(count))
	}
}

type BaseOption struct {
	dest net.IP
	count uint
	rate time.Duration
}

func (p *BaseOption) Dest() net.IP {
	return p.dest
}

func (p *BaseOption) Count() uint {
	return p.count
}

func (p *BaseOption) Rate() time.Duration {
	return p.rate
}

func (p *BaseOption) SetCount(count uint) {
	p.count = count
}

func (p *BaseOption) SetRate(wait time.Duration) {
	p.rate = wait
}

func (p *BaseOption) SetDest(dest net.IP) {
	p.dest = dest
}
