package protocol

import (
	_"fmt"
)

func putUint16(data []byte, u uint16) {
	data[0], data[1] = uint8(u >> 8), uint8(u & 0xff)
}

func putUint32(data []byte, u uint32) {
	data[0], data[1], data[2], data[3] = uint8(u >> 24), uint8(u >> 16 & 0xff), uint8(u >> 8 & 0xff), uint8(u & 0xff)
}

func tcpipChecksum(data []byte) uint16 {
	sum := uint32(0)
	length := len(data) - 1
	for i := 0; i < length; i += 2 {
		sum += uint32(data[i]) << 8
		sum += uint32(data[i + 1])
	}
	if len(data) & 1 == 1 {
		sum += uint32(data[length]) << 8
	}
	for sum > 0xffff {
		sum = (sum >> 16) + (sum & 0xffff)
	}
	return ^uint16(sum)
}

func SerializeLayers(data *[]byte, layers ...Layer) error {
	for i := len(layers) - 1; i >= 0; i-- {
		err := layers[i].ToBytes(data)
		if err != nil {
			return err
		}
	}
	return nil
}
