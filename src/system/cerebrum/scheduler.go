package cerebrum

import (
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/go-cyberbrain/src/system/archivist"
	"strconv"
	"strings"
)

type Scheduler struct {
	memory        *Memory
	demultiplexer *Demultiplexer
	log           *archivist.Archivist
}

func NewScheduler(memory *Memory, demultiplexerInstance *Demultiplexer, logger *archivist.Archivist) *Scheduler {
	return &Scheduler{
		memory:        memory,
		demultiplexer: demultiplexerInstance,
		log:           logger,
	}
}

func (s *Scheduler) Run(data transport.TransportEntity, cortex *Cortex) {
	// first we need to demultiplex the data we just gathered.
	// based on the results we can than identify and build new job payloads
	demultiplexedData := s.demultiplexer.Parse(data)
	s.log.Debug("Demultiplexed input", demultiplexedData)

	newRelationStructures := make(map[string][2]*transport.TransportEntity)
	newRelationStructures = s.rFilterRelationStructures(data, newRelationStructures)

	// build job inputs by each singleEntry of demultiplexed data
	for _, singleData := range demultiplexedData {
		s.createNewJobs(singleData, newRelationStructures, cortex)
	}
}

func (s *Scheduler) rFilterRelationStructures(entity transport.TransportEntity, relationStructures map[string][2]*transport.TransportEntity) map[string][2]*transport.TransportEntity {
	if 0 < len(entity.ChildRelations) {
		for _, childRelation := range entity.ChildRelations {
			if _, ok := childRelation.Properties["bMap"]; ok {
				tmpRelString := entity.Type + "-" + childRelation.Target.Type
				add := true
				for knownRelString, _ := range relationStructures {
					if knownRelString == tmpRelString {
						add = false
					}
				}
				if add {
					relationStructures[tmpRelString] = [2]*transport.TransportEntity{&entity, &childRelation.Target}
				}
			}
			relationStructures = s.rFilterRelationStructures(childRelation.Target, relationStructures)
		}
	}
	return relationStructures
}

func (s *Scheduler) createNewJobs(entity transport.TransportEntity, newRelationStructures map[string][2]*transport.TransportEntity, cortex *Cortex) []transport.TransportEntity {
	// first we will enrich some lookup variables we need later on
	// by recursively walking the given data
	lookup := make(map[string]int)
	var pointer [][]*transport.TransportEntity
	s.log.Debug("Enrich lookup by entity", entity)
	lookup, pointer = s.rEnrichLookupAndPointer(entity, lookup, pointer)
	s.log.Debug("Lookup data", lookup, pointer)
	// now we going to retrieve all action+dependency combos to that could potentially
	// be executed based on the new learned data which we just identified and stored
	// in our lookup/pointer variables
	var actionsAndDependencies [][2]string
	for entityType := range lookup {
		actionsAndDependencies = append(actionsAndDependencies, s.retrieveActionsByType(entityType)...)
	}
	s.log.Debug("Action and dependency found to input", actionsAndDependencies)

	// now also gonne lookup & enrich the actionsAndDependencies based on the newRelationStructures
	if 0 < len(newRelationStructures) {
		s.log.Debug("New relevant relation structures found in scheduler %+v", newRelationStructures)
		s.log.Debug("actionsAndDependencies before enrichin by relation structures", actionsAndDependencies)
		actionsAndDependencies = s.enrichActionsAndDependenciesByNewRelationStructures(newRelationStructures, actionsAndDependencies)
		s.log.Debug("actionsAndDependencies after enrichin by relation structures", actionsAndDependencies)
		s.log.Debug("lookupAndPointer before enrichment by relation structures", lookup, pointer)
		lookup, pointer = s.enrichLookupAndPointerByRelationStructures(newRelationStructures, lookup, pointer)
		s.log.Debug("lookupAndPointer after enrichment by relation structures", lookup, pointer)
		s.log.Debug("lookupAndPointer input structure", entity)
	}

	// at this point we go a single possible input structure and all potential actions/dependencies
	// that could be satisfied using it. Now we're going to try build actual input data by walking
	// through the dependencies and enrich an input datastructure using the given entity data and
	// the data that is in our storage
	for _, actionAndDependency := range actionsAndDependencies {
		act, _ := cortex.GetAction(actionAndDependency[0])
		requirement := act.GetDependencyByName(actionAndDependency[1])
		s.log.Debug("Trying to enrich data based on ", actionAndDependency)
		newJobInputs := s.buildInputData(requirement.Children()[0], lookup, pointer)
		//inputData, err := rBuildInputData(requirement.Children()[0], entity, pointer, lookup, false, "", -1, nil)
		if 0 < len(newJobInputs) {
			for _, inputData := range newJobInputs {
				s.log.Debug("Created a new job with payload", inputData)
				newJob := NewJob(s.memory, s.log)
				newJob.Create(act.GetName(), actionAndDependency[1], inputData)
			}
		} else {
			s.log.Debug("Requirement could not be satisfied", requirement)
		}

	}
	return []transport.TransportEntity{}
}

func (s *Scheduler) enrichLookupAndPointerByRelationStructures(newRelationStructures map[string][2]*transport.TransportEntity, lookup map[string]int, pointer [][]*transport.TransportEntity) (map[string]int, [][]*transport.TransportEntity) {
	for _, entityPair := range newRelationStructures {
		for _, entity := range entityPair {
			if _, ok := lookup[entity.Type]; !ok {
				pointer = append(pointer, []*transport.TransportEntity{entity})
				lookup[entity.Type] = len(pointer) - 1
			}
		}

	}
	return lookup, pointer
}

func (s *Scheduler) enrichActionsAndDependenciesByNewRelationStructures(newRelationStructures map[string][2]*transport.TransportEntity, actionsAndDependencies [][2]string) [][2]string {
	for relationStructure, _ := range newRelationStructures {
		actions := s.retrieveActionsByRelationStructure(relationStructure)
		s.log.Debug("Retrieved actions by relationStructure "+relationStructure, actions)
		for _, action := range actions {
			add := true
			for _, val := range actionsAndDependencies {
				if val[0] == action[0] && val[1] == action[1] {
					add = false
				}
			}
			if add {
				actionsAndDependencies = append(actionsAndDependencies, action)
			}
		}
	}
	return actionsAndDependencies
}

func (s *Scheduler) buildInputData(requirement transport.TransportEntity, lookup map[string]int, pointer [][]*transport.TransportEntity) []transport.TransportEntity {
	newJobs := []transport.TransportEntity{}
	qry := s.rBuildQuery(requirement, lookup, pointer)
	result := gits.GetDefault().Query().Execute(qry)

	if 0 < result.Amount {
		for _, enriched := range result.Entities {
			newJobs = append(newJobs, s.demultiplexer.Parse(enriched)...)
		}
	}
	return newJobs
}

func (s *Scheduler) rBuildQuery(requirement transport.TransportEntity, lookup map[string]int, pointer [][]*transport.TransportEntity) *query.Query {
	qry := query.New().Read(requirement.Value)
	// is requirement in index we add an exact ID matching filter
	if _, ok := lookup[requirement.Value]; ok {
		tmpEntity := pointer[lookup[requirement.Value]][0]
		qry.Match("ID", "==", strconv.Itoa(tmpEntity.ID))
	}
	// if its match mode we have to apply filters
	if requirement.Properties["Mode"] == "Match" {
		// we add match filters
		qry = s.enrichQueryFilters(qry, requirement)
	}
	// any child relations?
	if 0 < len(requirement.ChildRelations) {
		for _, childRelation := range requirement.ChildRelations {
			qry = qry.To(s.rBuildQuery(childRelation.Target, lookup, pointer))
		}
	}
	return qry
}

func (s *Scheduler) enrichQueryFilters(query *query.Query, requirement transport.TransportEntity) *query.Query {
	filters := make(map[string][]string)
	for name, val := range requirement.Properties {
		if len(name) > 6 && name[:6] == "Filter" {
			splitName := strings.Split(name, ".")
			// invalid structure
			if len(splitName) != 3 {
				s.log.Error("invalid filter format name: %s : skipping filter", name)
				continue // ### maybe should be handled different
			}
			key := splitName[1]
			typ := splitName[2]
			if _, ok := filters[key]; !ok {
				filters[key] = []string{"", "", ""}
			}
			switch typ {
			case "Field":
				filters[key][0] = val
			case "Operator":
				filters[key][1] = val
			case "Value":
				filters[key][2] = val
			}
		}
	}
	for _, val := range filters {
		query = query.Match(val[0], val[1], val[2])
	}
	return query
}

func (s *Scheduler) retrieveActionsByType(entityType string) [][2]string {
	var ret [][2]string
	qry := query.New().Read("DependencyEntityLookup").Match("Value", "==", entityType).To(
		query.New().Read("Dependency").From(
			query.New().Read("Action"),
		),
	)
	result := gits.GetDefault().Query().Execute(qry)
	s.log.Debug("DependencyEntityLookup ", entityType, result)
	if 0 < len(result.Entities) {
		for _, dependencyEntity := range result.Entities[0].Children() {
			for _, actionEntity := range dependencyEntity.Parents() { // ### todo : this is a very wierd behaviour, it works for us here but one would expect to also find the DependencyEntityLookup when checking the parents. but due to the way we build the return json tree its not
				ret = append(ret, [2]string{actionEntity.Value, dependencyEntity.Value})
			}
		}
	}
	return ret
}

func (s *Scheduler) retrieveActionsByRelationStructure(relationStructure string) [][2]string {
	var ret [][2]string
	qry := query.New().Read("DependencyRelationLookup").Match("Value", "==", relationStructure).To(
		query.New().Read("Dependency").From(
			query.New().Read("Action"),
		),
	)
	result := gits.GetDefault().Query().Execute(qry)
	s.log.Debug("DependencyRelationLookup ", relationStructure, result)
	if 0 < len(result.Entities) {
		for _, dependencyEntity := range result.Entities[0].Children() {
			for _, actionEntity := range dependencyEntity.Parents() { // ### todo : this is a very wierd behaviour, it works for us here but one would expect to also find the DependencyEntityLookup when checking the parents. but due to the way we build the return json tree its not
				ret = append(ret, [2]string{actionEntity.Value, dependencyEntity.Value})
			}
		}
	}
	return ret
}

func (s *Scheduler) rEnrichLookupAndPointer(entity transport.TransportEntity, lookup map[string]int, pointer [][]*transport.TransportEntity) (map[string]int, [][]*transport.TransportEntity) {
	s.log.Debug("Enrichting step", entity)
	// lets see if this is newly learned data
	if _, ok := entity.Properties["bMap"]; ok {
		// do we already know about this entity type?
		if _, well := lookup[entity.Type]; !well {
			// it's not known, so we create wa whole new first level entry on pointer and
			// also add it to our lookup map for later use
			s.log.Debug("Adding entity to pointer", entity)
			pointer = append(pointer, []*transport.TransportEntity{&entity})
			lookup[entity.Type] = len(pointer) - 1
		} else {
			// ### for now we gonne assume we only need the first upcome, later ones we skip. We might need to overthink
			// this since it hard impacts the scheduler (we would ne to multiplex on feeding into our dependency structure
			// in case there are same types on different levels. We keep this following line and else just in cae
			// we need to reactivate it. comments dont hurt
			//pointer[val] = append(pointer[val], &entity)
		}
	}
	for _, childRelation := range entity.ChildRelations {
		lookup, pointer = s.rEnrichLookupAndPointer(childRelation.Target, lookup, pointer)
	}
	//for _, parentRelation := range entity.ParentRelations { // ### enrichment towards parents is disabled for now
	//	lookup, pointer = rEnrichLookupAndPointer(parentRelation.Target, lookup, pointer)
	//}
	return lookup, pointer
}
