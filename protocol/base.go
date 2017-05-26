package protocol

type BaseLayer struct {
	Header []byte
	Payload []byte
}

func (b *BaseLayer) LayerHeader() []byte {
	return b.Header
}

func (b *BaseLayer) LayerPayload() []byte {
	return b.Payload
}
