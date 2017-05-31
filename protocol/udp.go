package protocol

import (
	"encoding/binary"
	"errors"
)

type UDP struct {
	BaseLayer
	PseudoHeader	[]byte
	SrcPort			uint16
	DstPort			uint16
	Length			uint16
	Checksum		uint16
	Payload			ApplicationLayer
}

func (u *UDP) LayerType() LayerType {
	return LayerTypeUDP
}

func (u *UDP) TransportPayload() ApplicationLayer {
	return u.Payload
}

func (u *UDP) ToBytes(data *[]byte) error {
	header := make([]byte, 8)
	if len(*data) + 8 > 0xffff {
		return errors.New("udp packet too large")
	}
	u.Length = uint16(len(*data) + 8)
	PutUint16 := binary.BigEndian.PutUint16
	PutUint16(header[0:2], u.SrcPort)
	PutUint16(header[2:4], u.DstPort)
	PutUint16(header[4:6], u.Length)
	PutUint16(header[6:8], uint16(0))
	*data = append(header, *data...)
	u.BaseLayer = BaseLayer{Header: (*data)[:8], Payload: (*data)[8:]}
	// rfc 768 udp pseudo header, but how about ipv6?
	u.PseudoHeader[8] = 0
	u.PseudoHeader[9] = byte(IPP_UDP)
	binary.BigEndian.PutUint16(u.PseudoHeader[10:12], u.Length)
	u.Checksum = udpChecksum(u.PseudoHeader, *data)
	PutUint16((*data)[6:8], u.Checksum)
	return nil
}

func udpChecksum(pseudoHeader, data []byte) uint16 {
	data = append(pseudoHeader, data...)
	return tcpipChecksum(data)
}
