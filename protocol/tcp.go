package protocol

import (
	"encoding/binary"
)

type TCP struct {
	BaseLayer
	PseudoHeader	[]byte
	SrcPort			uint16
	DstPort			uint16
	Seq				uint32
	Ack				uint32
	DataOffset		uint8
	Reserved		uint8
	Flags			TCPFlags
	Rwnd			uint16
	Checksum		uint16
	UrgPtr			uint16
	Options			[]byte
	Payload			ApplicationLayer
}

type TCPFlags struct {
	URG		bool
	ACK		bool
	PSH		bool
	RST		bool
	SYN		bool
	FIN		bool
}

func (t *TCPFlags) ToByte() byte {
	var b byte = 0
	if t.FIN {
		b |= 0x01
	}
	if t.SYN {
		b |= 0x02
	}
	if t.RST {
		b |= 0x04
	}
	if t.PSH {
		b |= 0x08
	}
	if t.ACK {
		b |= 0x10
	}
	if t.URG {
		b |= 0x20
	}
	return b
}

func (t *TCP) LayerType() LayerType {
	return LayerTypeTCP
}

func (t *TCP) TransportPayload() ApplicationLayer {
	return t.Payload
}

func (t *TCP) ToBytes(data *[]byte) error {
	length := 20 + (len(t.Options) + 3) / 4 * 4 
	t.DataOffset = uint8(length / 4)
	header := make([]byte, length, length)
	PutUint16 := binary.BigEndian.PutUint16
	PutUint32 := binary.BigEndian.PutUint32
	PutUint16(header[0:2], t.SrcPort)
	PutUint16(header[2:4], t.DstPort)
	PutUint32(header[4:8], t.Seq)
	PutUint32(header[8:12], t.Ack)
	header[12] = t.DataOffset << 4
	header[13] = t.Flags.ToByte()
	PutUint16(header[14:16], t.Rwnd)
	header[16] = 0
	header[17] = 0
	PutUint16(header[18:20], t.UrgPtr)
	if len(t.Options) > 0 {
		copy(header[20:], t.Options)
	}

	*data = append(header, *data...)
	t.BaseLayer= BaseLayer{Header: (*data)[:length], Payload: (*data)[length:]}

	t.PseudoHeader[8] = 0
	t.PseudoHeader[9] = byte(IPP_TCP)
	PutUint16(t.PseudoHeader[10:12], uint16(length))
	t.Checksum = tcpudpChecksum(t.PseudoHeader, *data)
	PutUint16((*data)[16:18], t.Checksum)
	return nil
}
