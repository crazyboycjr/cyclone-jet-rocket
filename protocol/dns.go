package protocol

import (
	"encoding/binary"
)
// rfc 1035

type DNS struct {
	BaseLayer
	ID			uint16
	QR			uint8
	Opcode		uint8
	AA			uint8
	TC			uint8
	RD			uint8
	RA			uint8
	Z			uint8
	RCODE		uint8
	QDCOUNT		uint16
	ANCOUNT		uint16
	NSCOUNT		uint16
	ARCOUNT		uint16
	Question	DNSQuestion
}

type DNSQuestion struct {
	QNAME			[]byte
	QTYPE			uint16
	QCLASS			uint16
}

func (d *DNSQuestion) ToBytes() []byte {
	length := len(d.QNAME) + 2 + 4
	data := make([]byte, length, length)
	curlen := 0
	for i, c := range d.QNAME {
		if c == '.' {
			data[i - curlen] = byte(curlen)
			curlen = 0
		} else {
			data[i + 1] = c
			curlen++
		}
	}
	data[len(d.QNAME) + 1] = 0
	data[len(d.QNAME) - curlen] = byte(curlen)
	binary.BigEndian.PutUint16(data[len(d.QNAME) + 2:], d.QTYPE)
	binary.BigEndian.PutUint16(data[len(d.QNAME) + 4:], d.QCLASS)
	return data
}

func (d *DNS) LayerType() LayerType {
	return LayerTypeDNS
}

func (d *DNS) ToBytes(data *[]byte) error {
	question := d.Question.ToBytes()
	header := make([]byte, 12 + len(question), 12 + len(question))
	PutUint16 := binary.BigEndian.PutUint16
	PutUint16(header[0:2], d.ID)
	header[2] = uint8(d.QR << 7 | d.Opcode << 3 | d.AA | d.TC | d.RD)
	header[3] = uint8(d.RA << 7 | d.Z << 4 | d.RCODE)
	PutUint16(header[4:6], d.QDCOUNT)
	PutUint16(header[6:8], d.ANCOUNT)
	PutUint16(header[8:10], d.NSCOUNT)
	PutUint16(header[10:12], d.ARCOUNT)
	// currently we do not consider ANCOUNT NSCOUNT and ARCOUNT
	copy(header[12:], question)
	*data = append(header, *data...)
	d.BaseLayer = BaseLayer{Header: (*data)[:len(header)], Payload: (*data)[len(header):]}
	return nil
}
