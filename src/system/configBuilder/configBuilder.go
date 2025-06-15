package configBuilder

import "github.com/voodooEntity/gits/src/transport"

type Priority string

const (
	PRIORITY_PRIMARY   Priority = "Primary"
	PRIORITY_SECONDARY Priority = "Secondary"
)

type Mode string

const (
	MODE_SET   Mode = "Set"
	MODE_MATCH Mode = "Match"
)

type ConfigBuilder struct {
	Dependencies map[string]*Structure
	Name         string
	Category     string
}

func NewConfig() *ConfigBuilder {
	return &ConfigBuilder{
		Dependencies: make(map[string]*Structure),
	}
}

func (builder *ConfigBuilder) SetName(name string) *ConfigBuilder {
	builder.Name = name
	return builder
}

func (builder *ConfigBuilder) SetCategory(category string) *ConfigBuilder {
	builder.Category = category
	return builder
}

func (builder *ConfigBuilder) AddDependency(name string, structure *Structure) *ConfigBuilder {
	builder.Dependencies[name] = structure
	return builder
}

func (builder *ConfigBuilder) Build() transport.TransportEntity {
	configStructure := transport.TransportEntity{
		ID:         -1,
		Type:       "Action",
		Value:      builder.Name,
		Context:    "System",
		Properties: make(map[string]string),
		ChildRelations: []transport.TransportRelation{
			{
				Target: transport.TransportEntity{
					Type:       "Category",
					Value:      builder.Category,
					Properties: make(map[string]string),
					Context:    "System",
				},
			},
		},
	}

	// nest the dependencies
	for name, structure := range builder.Dependencies {
		configStructure.ChildRelations = append(configStructure.ChildRelations, transport.TransportRelation{
			Target: transport.TransportEntity{
				ID:         -1,
				Type:       "Dependency",
				Value:      name,
				Context:    "System",
				Properties: make(map[string]string),
				ChildRelations: []transport.TransportRelation{
					{
						Target: structure.Transform(),
					},
				},
			},
		})
	}

	return configStructure
}

type Structure struct {
	Parents  []*Structure
	Children []*Structure
	Type     string
	Priority Priority
	Filter   map[string][3]string
	Mode     Mode
}

func NewStructure(nodeType string) *Structure {
	return &Structure{
		Parents:  make([]*Structure, 0),
		Children: make([]*Structure, 0),
		Filter:   make(map[string][3]string),
		Mode:     MODE_SET,
		Priority: PRIORITY_SECONDARY,
		Type:     nodeType,
	}
}

func (s *Structure) Transform() transport.TransportEntity {
	// create the base data for the current structure
	currEntity := transport.TransportEntity{
		Type:            "Structure",
		Value:           s.Type,
		ID:              -1,
		Context:         "System",
		Properties:      map[string]string{},
		ChildRelations:  make([]transport.TransportRelation, 0),
		ParentRelations: make([]transport.TransportRelation, 0),
	}

	// add the filters
	for key, value := range s.Filter {
		currEntity.Properties["Filter."+key+".Field"] = value[0]
		currEntity.Properties["Filter."+key+".Operator"] = value[1]
		currEntity.Properties["Filter."+key+".Value"] = value[2]
	}

	// set the mode & priority
	currEntity.Properties["Mode"] = string(s.Mode)
	currEntity.Properties["Type"] = string(s.Priority)

	// add the parents
	for _, parent := range s.Parents {
		currEntity.ParentRelations = append(currEntity.ParentRelations, transport.TransportRelation{
			Target: parent.Transform(),
		})
	}

	// add the children
	for _, child := range s.Children {
		currEntity.ChildRelations = append(currEntity.ChildRelations, transport.TransportRelation{
			Target: child.Transform(),
		})
	}

	return currEntity
}

func (s *Structure) AddParent(parent *Structure) *Structure {
	s.Parents = append(s.Parents, parent)
	return parent
}

func (s *Structure) AddChild(child *Structure) *Structure {
	s.Children = append(s.Children, child)
	return child
}

func (s *Structure) SetPriority(priority Priority) *Structure {
	s.Priority = priority
	return s
}

func (s *Structure) SetMode(mode Mode) *Structure {
	s.Mode = mode
	return s
}

func (s *Structure) AddFilter(name string, alpha string, operator string, beta string) *Structure {
	s.Filter[name] = [3]string{alpha, operator, beta}
	return s
}
