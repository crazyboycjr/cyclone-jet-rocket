package protocol

type LayerType int

const (
	LayerTypeIPv4			LayerType = 1
	LayerTypeICMPv4			LayerType = 2
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
