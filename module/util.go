package module

import (
	"syscall"
	"log"
	"errors"

	"cjr/protocol"
)

// The order must be bottom up
func send(fd int, l ...protocol.Layer) error {
	var data []byte
	err := protocol.SerializeLayers(&data, l...)
	if err != nil {
		return err
	}
	//
	layerType := l[0].LayerType()
	if layerType == protocol.LayerTypeIPv4 {
		ip := l[0].(*protocol.IPv4Packet)
		// this addr is used to choose the out interface
		tmpip := ip.DstIP.To4()
		dstip := [4]byte{tmpip[0], tmpip[1], tmpip[2], tmpip[3]}
		//log.Println("dstip = ", dstip)

		addr := &syscall.SockaddrInet4 {
			Port: 0,
			Addr: dstip,
		}
		err = syscall.Sendto(fd, data, 0, addr)
		if err != nil {
			if err == errors.New("Bad file descriptor") {
				log.Fatal("You may need to execute setcap cap_net_raw+ep `which cjr`")
			} else {
				log.Fatal("Sendto: ", err, " ", addr)
			}
		}
	} /*else if layerType == protocol.LayerTypeIPv6 {
		ip := l[0].(protocol.IPv6Packet)
		fd, err := syscall.Socket(syscall.AF_INET6, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
		addr := &syscall.SockaddrInet6 {
			Port: 0,
			Addr: ip.DstIP,
		}
		err = syscall.Sendto(fd, data, 0, addr)
		if err != nil {
			log.Fatal("Sendto:", err)
		}
	}*/
	return nil
}
