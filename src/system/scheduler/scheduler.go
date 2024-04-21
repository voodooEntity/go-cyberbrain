package scheduler

import (
	"errors"
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/go-cyberbrain/src/system/demultiplexer"
	"github.com/voodooEntity/go-cyberbrain/src/system/job"
	"github.com/voodooEntity/go-cyberbrain/src/system/registry"
	"github.com/voodooEntity/go-cyberbrain/src/system/util"
	"strconv"
	"strings"
)

func Run(data transport.TransportEntity, registry registry.Registry) {
	// first we need to demultiplex the data we just gathered.
	// based on the results we can than identify and build new job payloads
	demultiplexedData := demultiplexer.Parse(data)
	archivist.Debug("Demultiplexed input", demultiplexedData)

	// build job inputs by each singleEntry of demultiplexed data
	for _, singleData := range demultiplexedData {
		createNewJobs(singleData, registry)
	}
}

func createNewJobs(entity transport.TransportEntity, registry registry.Registry) []transport.TransportEntity {
	// first we will enrich some lookup variables we need later on
	// by recursively walking the given data
	lookup := make(map[string]int)
	var pointer [][]*transport.TransportEntity
	archivist.Debug("Enrich lookup by entity", entity)
	lookup, pointer = rEnrichLookupAndPointer(entity, lookup, pointer)
	archivist.Debug("Lookup data", lookup, pointer)
	// now we going to retrieve all action+dependency combos to that could potentially
	// be executed based on the new learned data which we just identified and stored
	// in our lookup/pointer variables
	var actionsAndDependencies [][2]string
	for entityType := range lookup {
		actionsAndDependencies = append(actionsAndDependencies, retrieveActionsByType(entityType)...)
	}
	archivist.Debug("Action and dependency found to input", actionsAndDependencies)

	// at this point we go a single possible input structure and all potential actions/dependencies
	// that could be satisfied using it. Now we're going to try build actual input data by walking
	// through the dependencies and enrich an input datastructure using the given entity data and
	// the data that is in our storage ### should be revisited , the way this is going to be implemented
	// could possibly satisfy multiple dependency structures of the same action at once. im not sure
	// if that is actually something we want but its easy to purge later on and for now I'm going to do it
	// this way.
	for _, actionAndDependency := range actionsAndDependencies {
		act, _ := registry.GetAction(actionAndDependency[0])
		requirement := act.GetDependencyByName(actionAndDependency[1])
		inputData, err := rBuildInputData(requirement.Children()[0], entity, pointer, lookup, false, "", -1, nil)
		// if we got no err the inputData should be complete to create a new job
		if nil != err {
			// ### must review, is this an actual error case? i think this only occurs on dependencies which could not get fully satisfied which
			// us a legitimate case to happen (we check the data for types but not structure) !important ###
			archivist.Debug(err.Error(), requirement.Children()[0], entity)
		} else {
			archivist.Debug("Created a new job with payload", inputData)
			job.Create(act.GetName(), actionAndDependency[1], inputData)
		}
	}
	return []transport.TransportEntity{}
}

func rBuildInputData(
	requirement transport.TransportEntity,
	pool transport.TransportEntity,
	lookupPointer [][]*transport.TransportEntity,
	lookupIndex map[string]int,
	inPool bool,
	parentType string,
	parentId int,
	nextEntity *transport.TransportEntity,
) (transport.TransportEntity, error) {
	archivist.Debug("buildinput step ", requirement, pool, lookupPointer, lookupIndex, inPool, parentType, parentId, nextEntity)
	var ret transport.TransportEntity
	if inPool {
		archivist.Debug("buildInput inPool")
		// apply matching filters
		if !applyMatchingFilter(*nextEntity, requirement) {
			return transport.TransportEntity{}, errors.New("Unsatisfiable requirement by filter")
		}
		// we are inPool so the pool variable will be the next entity we are searchin for
		ret = transport.TransportEntity{
			ID:         nextEntity.ID,
			Type:       nextEntity.Type,
			Value:      nextEntity.Value,
			Context:    nextEntity.Context,
			Properties: util.CopyStringStringMap(nextEntity.Properties),
		}
		var tmpChildren []transport.TransportRelation
		for _, childRequirement := range requirement.ChildRelations {
			foundNext := false
			var nextSub transport.TransportEntity
			for _, nextSubEntity := range nextEntity.Children() {
				if childRequirement.Target.Type == nextSubEntity.Type {
					foundNext = true
					nextSub = nextSubEntity
				}
			}
			// if the next entity is a sub of the entity we just took from pool we again pass it as next entity and
			// inPool true
			if foundNext {
				archivist.Debug("buildInput inPool foundnext")
				retEntity, err := rBuildInputData(childRequirement.Target, pool, lookupPointer, lookupIndex, true, nextEntity.Type, nextEntity.ID, &nextSub)
				tmpChildren = append(tmpChildren, transport.TransportRelation{
					Target: retEntity,
				})
				// handles errors
				if nil != err {
					return transport.TransportEntity{}, err
				}
			} else {
				archivist.Debug("buildInput inPool no found next")
				// the next sub is not a child of the from pool retrieved entity so we are not in pool any more
				// thatfor we pass with inPool false to retrieve data from storage
				retEntity, err := rBuildInputData(childRequirement.Target, pool, lookupPointer, lookupIndex, false, nextEntity.Type, nextEntity.ID, nil)
				tmpChildren = append(tmpChildren, transport.TransportRelation{
					Target: retEntity,
				})
				// handles errors
				if nil != err {
					return transport.TransportEntity{}, err
				}
			}
		}
		// store the retrieved children
		ret.ChildRelations = tmpChildren
	} else {
		archivist.Debug("buildInput not inPool")
		// do we have new data to the requested type?
		if typeId, ok := lookupIndex[requirement.Value]; ok {
			archivist.Debug("buildInput not inPool found in lookup")
			// we retrieve the first entry since we right now only store the first upcome
			// we tho keep rn the subarray structure in case this was stupid and i need to adjust it
			// tho keeping this comment so we know it and can refactor it easy in case ###
			retrievedEntity := *lookupPointer[typeId][0]
			// apply matching filters
			if !applyMatchingFilter(retrievedEntity, requirement) {
				return transport.TransportEntity{}, errors.New("Unsatisfiable requirement by filter")
			}
			ret = transport.TransportEntity{
				Type:       retrievedEntity.Type,
				ID:         retrievedEntity.ID,
				Value:      retrievedEntity.Value,
				Properties: util.CopyStringStringMap(retrievedEntity.Properties),
				Context:    retrievedEntity.Context,
			}
			var tmpChildren []transport.TransportRelation
			for _, childRequirement := range requirement.Children() {
				foundNext := false
				var nextPoolEntity transport.TransportEntity
				for _, child := range retrievedEntity.Children() {
					if childRequirement.Value == child.Type {
						nextPoolEntity = child
						foundNext = true
					}
				}
				// in case we found the child in the children of our pool
				// retrieved entities children we gonne go on mapping using it
				// saying inPool = true
				if foundNext {
					retEntity, err := rBuildInputData(childRequirement, pool, lookupPointer, lookupIndex, true, retrievedEntity.Type, retrievedEntity.ID, &nextPoolEntity)
					tmpChildren = append(tmpChildren, transport.TransportRelation{
						Target: retEntity,
					})
					// handles errors
					if nil != err {
						return transport.TransportEntity{}, err
					}
				} else {
					retEntity, err := rBuildInputData(childRequirement, pool, lookupPointer, lookupIndex, false, retrievedEntity.Type, retrievedEntity.ID, nil)
					tmpChildren = append(tmpChildren, transport.TransportRelation{
						Target: retEntity,
					})
					// handles errors
					if nil != err {
						return transport.TransportEntity{}, err
					}
				}
			}
			ret.ChildRelations = tmpChildren
		} else {
			archivist.Debug("buildInput not inPool not found in lookup")
			// so we are not inPool and we dont have the current dependency in our lookup
			// thatfor it must be resolved from storage. In this case we have to differ if we are below
			// an already identified entity or above. If we are below there are coordinates given so we check
			// if parentid is != -1
			if parentId != -1 {
				archivist.Debug("buildInput not inPool not found in lookup parent given")
				// seems like we identified the parental entity so we need to find the current one linked to the parental one
				// thatfor we query
				qry := query.New().Read(requirement.Value).From(query.New().Read(parentType).Match("ID", "==", strconv.Itoa(parentId)))
				match := query.Execute(qry)
				// did we get any hits?
				if match.Amount == 0 {
					// no hits so we couldnt satisfy our requirement thatfor we gonne return error to stop this
					// lookup process
					return transport.TransportEntity{}, errors.New("Unsatisfiable requirement by type")
				} else {
					// apply matching filters
					if !applyMatchingFilter(match.Entities[0], requirement) {
						return transport.TransportEntity{}, errors.New("Unsatisfiable requirement by filter")
					}
					// seems like we get a hit, thatfor we gonne take it and go on processing
					// for now we gonne assume we only get 1 hit, thos this might be wrong and we
					// have to adjust this part. im not sure rn ### refactor
					var tmpChildren []transport.TransportRelation
					for _, childRequirement := range requirement.Children() {
						childEntity, err := rBuildInputData(childRequirement, pool, lookupPointer, lookupIndex, false, match.Entities[0].Type, match.Entities[0].ID, nil)
						// first we check if we could satisfy the following calls
						if nil != err {
							return transport.TransportEntity{}, err
						}
						// if we could we gonne add the childEntity to our tmpChildren
						tmpChildren = append(tmpChildren, transport.TransportRelation{
							Target: childEntity,
						})
					}
					// finally update our return
					ret = transport.TransportEntity{
						ID:             match.Entities[0].ID,
						Type:           match.Entities[0].Type,
						Value:          match.Entities[0].Value,
						Context:        match.Entities[0].Context,
						ChildRelations: tmpChildren,
					}
				}
			} else {
				archivist.Debug("buildInput not inPool not found in lookup no parent given")
				// if we got no parent IDs given meaning we are at the first levels before hitting the first pool entities
				// ### IMPLEMENT ME
				qry := query.New().Read(requirement.Value)
				var tmpChildren []transport.TransportRelation
				for _, childRequirement := range requirement.Children() {
					childEntity, err := rBuildInputData(childRequirement, pool, lookupPointer, lookupIndex, inPool, "", -1, nil)
					// handle error
					if nil != err {
						return transport.TransportEntity{}, err
					}
					// add the data to our tmpChildren
					tmpChildren = append(tmpChildren, transport.TransportRelation{
						Target: childEntity,
					})
					// append our query to filter correctly
					qry = qry.To(query.New().Reduce(childEntity.Type).Match("ID", "==", strconv.Itoa(childEntity.ID)))
				}
				archivist.Debug("buildInput not inPool not found in lookup no parent given dynamic query", *qry)
				// now we search for the entity that is mapped to all the given children
				result := query.Execute(qry)
				// if we found none its an error
				if 0 == result.Amount {
					return transport.TransportEntity{}, errors.New("Unsatisfiable requirement by type")
				}
				// apply filters
				if !applyMatchingFilter(result.Entities[0], requirement) {
					return transport.TransportEntity{}, errors.New("Unsatisfiable requirement by filter")
				}
				// seems like we found the parent, for now we gonne assume there only is one parent. in a even more
				// dynamic datastructure there might be multiple but for the current implementation we gonne assume
				// entities just have 1 parent thatfor we pick the first result.
				ret = transport.TransportEntity{
					Type:           result.Entities[0].Type,
					ID:             result.Entities[0].ID,
					Value:          result.Entities[0].Value,
					Context:        result.Entities[0].Context,
					Properties:     util.CopyStringStringMap(result.Entities[0].Properties),
					ChildRelations: tmpChildren,
				}
			}
		}
	}
	return ret, nil
}

func applyMatchingFilter(entity transport.TransportEntity, requirement transport.TransportEntity) bool {
	filterType := requirement.Properties["Mode"]
	if "Set" == filterType {
		return true
	} else if "Match" == filterType {
		return matchFields(requirement.Properties["FilterValue"], requirement.Properties["FilterOperator"], entity.Value)
	}
	archivist.Debug("No known matching type given in dependency, prolly a plugin bug", requirement)
	return false
}

func matchFields(alpha string, operator string, beta string) bool {
	switch operator {
	case "==":
		if alpha == beta {
			return true
		}
	case "prefix":
		// starts with
		if strings.HasPrefix(alpha, beta) {
			return true
		}
	case "suffix":
		// ends with
		if strings.HasSuffix(alpha, beta) {
			return true
		}
	case "contain":
		// string contains string
		if strings.Contains(alpha, beta) {
			return true
		}
	case ">":
		alphaInt, err := strconv.Atoi(alpha)
		if nil != err {
			return false
		}
		betaInt, err := strconv.Atoi(beta)
		if nil != err {
			return false
		}
		if alphaInt > betaInt {
			return true
		}
	case ">=":
		alphaInt, err := strconv.Atoi(alpha)
		if nil != err {
			return false
		}
		betaInt, err := strconv.Atoi(beta)
		if nil != err {
			return false
		}
		if alphaInt >= betaInt {
			return true
		}
	case "<":
		alphaInt, err := strconv.Atoi(alpha)
		if nil != err {
			return false
		}
		betaInt, err := strconv.Atoi(beta)
		if nil != err {
			return false
		}
		if alphaInt < betaInt {
			return true
		}
	case "<=":
		alphaInt, err := strconv.Atoi(alpha)
		if nil != err {
			return false
		}
		betaInt, err := strconv.Atoi(beta)
		if nil != err {
			return false
		}
		if alphaInt <= betaInt {
			return true
		}
	case "in":
		list := strings.Split(beta, ",")
		for _, value := range list {
			if alpha == value {
				return true
			}
		}
	}
	return false
}

func retrieveActionsByType(entityType string) [][2]string {
	var ret [][2]string
	qry := query.New().Read("DependencyLookup").Match("Value", "==", entityType).To(
		query.New().Read("Dependency").From(
			query.New().Read("Action"),
		),
	)
	result := query.Execute(qry)
	archivist.Debug("DependencyLookup ", entityType, result)
	if 0 < len(result.Entities) {
		for _, dependencyEntity := range result.Entities[0].Children() {
			for _, actionEntity := range dependencyEntity.Parents() {
				ret = append(ret, [2]string{actionEntity.Value, dependencyEntity.Value})
			}
		}
	}
	return ret
}

func rEnrichLookupAndPointer(entity transport.TransportEntity, lookup map[string]int, pointer [][]*transport.TransportEntity) (map[string]int, [][]*transport.TransportEntity) {
	archivist.Debug("Enrichting step", entity)
	// lets see if this is newly learned data
	if _, ok := entity.Properties["bMap"]; ok {
		// do we already know about this entity type?
		if _, well := lookup[entity.Type]; !well {
			// it's not known, so we create wa whole new first level entry on pointer and
			// also add it to our lookup map for later use
			archivist.Debug("Adding entity to pointer", entity)
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
		lookup, pointer = rEnrichLookupAndPointer(childRelation.Target, lookup, pointer)
	}
	//for _, parentRelation := range entity.ParentRelations {
	//	lookup, pointer = rEnrichLookupAndPointer(parentRelation.Target, lookup, pointer)
	//}
	return lookup, pointer
}
