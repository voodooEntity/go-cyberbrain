package registry

import (
	"errors"
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/go-cyberbrain-plugin-interface/src/interfaces"
	"github.com/voodooEntity/go-cyberbrain/src/system/action"
	"github.com/voodooEntity/go-cyberbrain/src/system/config"
	"github.com/voodooEntity/go-cyberbrain/src/system/mapper"
	"github.com/voodooEntity/go-cyberbrain/src/system/util"
	"strings"
)

var pluginsDir string

type Registry map[string]action.Action

var Data Registry

func New() Registry {
	return make(Registry)
}

func (self Registry) Index() Registry {
	// retrieve all available plugins
	pluginsDir = config.Get("PLUGIN_DIR")
	plugins := util.GetAvailablePlugins(pluginsDir)
	archivist.Info("Plugins found", plugins)
	//walk through our plugins,
	for _, plug := range plugins {
		// load the plugin
		instance, err := util.LoadPlugin(pluginsDir, plug)
		if nil != err {
			archivist.Error(err.Error())
			// skip current loop on error
			continue
		}
		archivist.Debug("Registration of plugin", plug, instance)
		// register the plugin to our ModuleRegistry
		self.registerModule(plug, instance)
	}
	archivist.DebugF("Register filled with %+v", self)
	return self
}

func (self Registry) registerModule(plug string, instance interfaces.PluginInterface) Registry {
	// store plugin config
	mapper.MapTransportDataWithContext(instance.GetConfig(), "System")

	// remove the.so from plug string
	name := strings.TrimSuffix(plug, ".so")

	// Get the mapped categories
	catQry := query.New().Read("Action").Match("Value", "==", name).To(query.New().Read("Category").TraverseOut(10))
	categories := gits.GetDefault().Query().Execute(catQry)

	// Get the mapped dependencies
	depQry := query.New().Read("Action").Match("Value", "==", name).To(query.New().Read("Dependency").TraverseOut(10))
	dependencies := gits.GetDefault().Query().Execute(depQry)

	// create a module struct instance satisfied with the just mapped config and dependency data & the plugin instance itself
	moduleInstance := *action.New().SetName(name).SetDependencies(dependencies.Entities[0].Children()).SetCategories(categories.Entities[0].Children()).SetPlugin(instance)

	// recursive filter all upcoming dependency types and map them onto lookup nodes for further faster processing
	for _, val := range moduleInstance.GetDependencies() {
		dependencyTypeList := getDependencyStructureTypes(val)
		archivist.Debug("Mapping dependency lookup ", dependencyTypeList, val.ID)
		mapDependencyEntityLookupNodes(dependencyTypeList, val.ID)
		mapDependencyRelationLookupNodes(val)
	}

	// finally we place the module instance inside our map
	self[name] = moduleInstance

	return self
}

func (self Registry) GetAction(name string) (action.Action, error) {
	if val, ok := self[name]; ok {
		return val, nil
	}
	return action.Action{}, errors.New("Action '" + name + "'not found in registry")
}

func (self Registry) GetInstance(name string) (interfaces.PluginInterface, error) {
	if val, ok := self[name]; ok {
		return val.GetPlugin().New(), nil
	}
	return nil, errors.New("Action '" + name + "'not found in registry")
}

func mapDependencyRelationLookupNodes(entity transport.TransportEntity) {
	var relationStructures []string
	relationStructures = rFindRelationStructures(entity, relationStructures)
	archivist.Info("Relation structures found in registry ", relationStructures)
	for _, val := range relationStructures {
		gits.GetDefault().MapData(transport.TransportEntity{
			Type:       "DependencyRelationLookup",
			ID:         0,
			Value:      val,
			Context:    "System",
			Properties: make(map[string]string),
			ChildRelations: []transport.TransportRelation{
				transport.TransportRelation{
					Context: "Structure",
					Target: transport.TransportEntity{
						Type: "Dependency",
						ID:   entity.ID,
					},
				},
			},
		})
	}
}

func rFindRelationStructures(entity transport.TransportEntity, relationStructures []string) []string {
	if 0 < len(entity.ChildRelations) {
		for _, childRelation := range entity.ChildRelations {
			tmpRelString := entity.Value + "-" + childRelation.Target.Value
			add := true
			for _, knownRelString := range relationStructures {
				if knownRelString == tmpRelString {
					add = false
				}
			}
			if add {
				relationStructures = append(relationStructures, tmpRelString)
			}
			relationStructures = rFindRelationStructures(childRelation.Target, relationStructures)
		}
	}
	return relationStructures
}

func mapDependencyEntityLookupNodes(dependencyTypes []string, dependencyId int) {
	for _, val := range dependencyTypes {
		gits.GetDefault().MapData(transport.TransportEntity{
			Type:       "DependencyEntityLookup",
			ID:         0,
			Value:      val,
			Context:    "System",
			Properties: make(map[string]string),
			ChildRelations: []transport.TransportRelation{
				transport.TransportRelation{
					Context: "Structure",
					Target: transport.TransportEntity{
						Type: "Dependency",
						ID:   dependencyId,
					},
				},
			},
		})
	}
}

func getDependencyStructureTypes(entity transport.TransportEntity) []string {
	var typeList []string
	rGetTypeList(&typeList, []transport.TransportEntity{entity})
	return typeList
}

func rGetTypeList(typeList *[]string, data []transport.TransportEntity) {
	for _, val := range data {
		// ### refactor if type changes , context should be structure so we gonne name this structure
		// but with context structurethis should be fiitting for all cases prolly i dunno. Also we check if type
		// is Primary, because only a Primary should trigger a new job , just having an state:open should not trigger
		// everything that might filter for a state. tho if its a Port with state open the Port should trigger. Thatfor
		// dependency structures should be marked as Primary or Secondary
		if !util.StringInArray(*typeList, val.Value) && "Structure" == val.Type && val.Properties["Type"] == "Primary" {
			*typeList = append(*typeList, val.Value)
		}
		if 0 < len(val.ChildRelations) {
			rGetTypeList(typeList, val.Children())
		}
	}
}
