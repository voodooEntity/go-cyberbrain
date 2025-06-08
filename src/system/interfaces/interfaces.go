package interfaces

import (
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/transport"
)

type ActionInterface interface {
	Execute(*gits.Gits, transport.TransportEntity, string, string) ([]transport.TransportEntity, error)
	GetConfig() transport.TransportEntity
}
