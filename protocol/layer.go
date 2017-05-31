package protocol

type LayerType int

const (
	LayerTypeIPv4			LayerType = 1
	LayerTypeICMPv4			LayerType = 2
	LayerTypeUDP			LayerType = 3
	LayerTypeTCP			LayerType = 4
	LayerTypeUnknownApp		LayerType = 999
)

type Layer interface {
	LayerType() LayerType
	LayerHeader() []byte
	LayerPayload() []byte
	//FromBytes(data []byte) error
	// not necessary every Layer can ToBytes
	ToBytes(data *[]byte) error
}

type LinkLayer interface {
	Layer
	LinkPayload() NetworkLayer
}

type NetworkLayer interface {
	Layer
	NetworkPayload() TransportLayer
}

type TransportLayer interface {
	Layer
	TransportPayload() ApplicationLayer
}

type ApplicationLayer interface {
	Layer
	ApplicationPayload() []byte
}

type UnknownApplicationLayer struct {
	BaseLayer
	Data []byte
}

func (u *UnknownApplicationLayer) LayerType() LayerType {
	return LayerTypeUnknownApp
}

func (u *UnknownApplicationLayer) ApplicationPayload() []byte {
	return u.Data
}


func (u *UnknownApplicationLayer) ToBytes(data *[]byte) error {
	*data = append(u.Data, *data...)
	u.BaseLayer = BaseLayer{Header: *data, Payload: *data}
	return nil
}
