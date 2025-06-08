package test

import (
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/transport"
	"net"
	"strconv"
	"time"
)

type Test struct {
}

// Execute method mandatory
func (self *Test) Execute(gitsInstance *gits.Gits, input transport.TransportEntity, requirement string, context string) ([]transport.TransportEntity, error) {
	archivist.DebugF("Plugin executed with input %+v", input)
	ips, err := net.LookupIP(input.Value)
	if nil != err {
		return []transport.TransportEntity{}, err
	}
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			archivist.Debug("ipv4 ", ipv4)
			// if we find an IPv4 we create a parental entity IP that links to our input domain dataset and return it for further processing
			input.ChildRelations = append(input.ChildRelations, transport.TransportRelation{
				Target: transport.TransportEntity{
					ID:         -2,
					Type:       "IP",
					Value:      ipv4.String(),
					Context:    context,
					Properties: map[string]string{"protocol": "V4", "created": strconv.FormatInt(time.Now().Unix(), 10)},
				}})
		}
	}
	input.Properties = make(map[string]string)
	return []transport.TransportEntity{input}, nil
}

func (self *Test) GetConfig() transport.TransportEntity {
	return transport.TransportEntity{
		ID:         -1,
		Type:       "Action",
		Value:      "resolveIPFromDomain",
		Context:    "System",
		Properties: make(map[string]string),
		ChildRelations: []transport.TransportRelation{
			{
				Target: transport.TransportEntity{
					ID:         -1,
					Type:       "Dependency",
					Value:      "alpha",
					Context:    "System",
					Properties: make(map[string]string),
					ChildRelations: []transport.TransportRelation{
						{
							Target: transport.TransportEntity{
								ID:         -1,
								Type:       "Structure",
								Value:      "Domain",
								Context:    "System",
								Properties: map[string]string{"Mode": "Set", "Type": "Primary"},
							},
						},
					},
				},
			},
			{
				Target: transport.TransportEntity{
					Type:       "Category",
					Value:      "Pentest",
					Properties: make(map[string]string),
					Context:    "System",
				},
			},
		},
	}
}

func New() *Test {
	tmp := &Test{}
	return tmp
}
