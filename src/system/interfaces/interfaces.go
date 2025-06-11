package interfaces

import (
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/transport"
)

type ActionInterface interface {
	Execute(transport.TransportEntity, string, string) ([]transport.TransportEntity, error)
	SetGits(*gits.Gits) ActionInterface
	GetConfig() transport.TransportEntity
}
