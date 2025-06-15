package configBuilder

import "github.com/voodooEntity/gits/src/transport"

type ConfigBuilder struct {
	Dependencies map[string]*transport.TransportEntity
	Name         string
	Categories   []transport.TransportEntity
}

func NewConfig() *ConfigBuilder {
	return &ConfigBuilder{}
}

type NewStructure struct {
}
