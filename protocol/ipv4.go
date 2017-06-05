package protocol

import (
	"encoding/binary"
	"net"
)

type IPProtocol uint8

const (
	IPP_ICMP      IPProtocol = 0x01
	IPP_TCP       IPProtocol = 0x06
	IPP_UDP       IPProtocol = 0x11
	IPP_TLSP      IPProtocol = 0x38
	IPP_IPV6_ICMP IPProtocol = 0x3A
	IPP_SCTP      IPProtocol = 0x84
)

type IPv4Packet struct {
	BaseLayer
	Version			uint8
	IHL				uint8
	DSCP			uint8
	ECN				uint8
	TotalLength		uint16
	ID				uint16
	DF				bool
	MF				bool
	FragmentOffset	uint16
	TTL				uint8
	Protocol		IPProtocol
	Checksum		uint16
	SrcIP			net.IP
	DstIP			net.IP
	Options			[]byte
	Payload			TransportLayer
}

func (p *IPv4Packet) InternetPayload() TransportLayer {
	return p.Payload
}

func (p *IPv4Packet) LayerType() LayerType {
	return LayerTypeIPv4
}

func (p *IPv4Packet) ToBytes(data *[]byte) error {
	p.IHL = uint8((20 + len(p.Options)) / 4)
	header := make([]byte, p.IHL * 4)
	p.TotalLength = uint16(len(header) + len(*data))

	header[0] = ((p.Version & 0xf) << 4) | (p.IHL & 0xf)
	header[1] = ((p.DSCP & 0x3f) << 4) | (p.ECN & 0x3)
	binary.BigEndian.PutUint16(header[2:4], p.TotalLength)

	binary.BigEndian.PutUint16(header[4:6], p.ID)
	flag := 0
	if p.DF {
		flag |= 0x2
	}
	if p.MF {
		flag |= 0x1
	}
	binary.BigEndian.PutUint16(header[6:8], uint16((flag << 13) | (int(p.FragmentOffset) & 0x1fff)))

	header[8] = p.TTL
	header[9] = byte(p.Protocol)
	binary.BigEndian.PutUint16(header[10:12], uint16(0))

	copy(header[12:16], p.SrcIP.To4())
	copy(header[16:20], p.DstIP.To4())

	if len(p.Options) > 0 {
		copy(header[20:], p.Options)
	}

	*data = append(header, *data...)
	p.BaseLayer = BaseLayer{Header: (*data)[:len(header)], Payload: (*data)[len(header):]}

	p.Checksum = checksum(header)
	binary.BigEndian.PutUint16((*data)[10:12], p.Checksum)
	return nil
}

func checksum(data []byte) uint16 {
	sum := uint32(0)
	for i := 0; i < len(data); i += 2 {
		sum += uint32(data[i]) << 8
		sum += uint32(data[i + 1])
	}
	for {
		if sum <= 65535 {
			break
		}
		sum = (sum >> 16) + (sum & 0xffff)
	}
	return ^uint16(sum)
}

