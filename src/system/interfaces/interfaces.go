package interfaces

import (
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/go-cyberbrain/src/system/cerebrum"
)

type ActionInterface interface {
	Execute(*cerebrum.Memory, transport.TransportEntity, string, string) ([]transport.TransportEntity, error)
	GetConfig() transport.TransportEntity
}
