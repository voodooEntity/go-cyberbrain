package action

import (
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/go-cyberbrain-plugin-interface/src/interfaces"
)

type Action struct {
	name         string
	categories   []transport.TransportEntity
	dependencies []transport.TransportEntity
	plugin       interfaces.PluginInterface
}

func New() *Action {
	return &Action{}
}

func (self *Action) SetName(name string) *Action {
	self.name = name
	return self
}

func (self *Action) GetName() string {
	return self.name
}

func (self *Action) SetDependencies(dependencies []transport.TransportEntity) *Action {
	self.dependencies = dependencies
	return self
}

func (self *Action) GetDependencies() []transport.TransportEntity {
	return self.dependencies
}

func (self *Action) GetDependencyByName(name string) transport.TransportEntity {
	for _, dependency := range self.GetDependencies() {
		if name == dependency.Value {
			return dependency
		}
	}
	return transport.TransportEntity{}
}

func (self *Action) SetPlugin(plugin interfaces.PluginInterface) *Action {
	self.plugin = plugin
	return self
}
func (self *Action) GetPlugin() interfaces.PluginInterface {
	return self.plugin
}

func (self *Action) GetInstance() interfaces.PluginInterface {
	return self.GetPlugin().New()
}

func (self *Action) SetCategories(categories []transport.TransportEntity) *Action {
	self.categories = categories
	return self
}
func (self *Action) GetCategories() []transport.TransportEntity {
	return self.categories
}
