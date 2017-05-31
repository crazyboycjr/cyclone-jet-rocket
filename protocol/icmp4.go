package protocol

import (
	"encoding/binary"
)

type ICMPv4 struct {
	BaseLayer
	Type			uint8
	Code			uint8
	Checksum		uint16
	Id				uint16
	Seq				uint16
	Data			[]byte
}

func (p *ICMPv4) LayerType() LayerType {
	return LayerTypeICMPv4
}

func (p *ICMPv4) ToBytes(data *[]byte) error {
	header := make([]byte, len(p.Data) + 8)
	header[0] = p.Type
	header[1] = p.Code
	binary.BigEndian.PutUint16(header[2:4], uint16(0))
	binary.BigEndian.PutUint16(header[4:6], p.Id)
	binary.BigEndian.PutUint16(header[6:8], p.Seq)
	copy(header[8:], p.Data)

	p.Checksum = tcpipChecksum(header)
	binary.BigEndian.PutUint16(header[2:4], p.Checksum)
	*data = append(header, *data...)
	p.BaseLayer = BaseLayer{Header: (*data)[:8], Payload: (*data)[8:]}

	return nil
}
