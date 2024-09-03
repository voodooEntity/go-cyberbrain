package mapper

import (
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/gits/src/types"
	"github.com/voodooEntity/go-cyberbrain/src/system/util"
)

// custom implementation of gits MapTransportData. Since we need to identify
// every new dataset for later dispatching of new jobs its important to store
// all those in some way. This code is very close to the one from gits tho
// for the moment the given functionality is not neccesary for gits so its not
// implemented in storage itself
func MapTransportData(data transport.TransportEntity) transport.TransportEntity {
	// first we lock all the storages
	gits.EntityTypeMutex.Lock()
	gits.EntityStorageMutex.Lock()
	gits.RelationStorageMutex.Lock()

	// lets start recursive mapping of the data
	archivist.DebugF("MapTransportData ", data)
	ret := mapRecursive(data, -1, -1, gits.DIRECTION_NONE, "", false, nil)

	// now we unlock all the mutexes again
	gits.RelationStorageMutex.Unlock()
	gits.EntityStorageMutex.Unlock()
	gits.EntityTypeMutex.Unlock()

	return ret
}

func MapTransportDataWithContext(data transport.TransportEntity, context string) transport.TransportEntity {
	// first we lock all the storages
	gits.EntityTypeMutex.Lock()
	gits.EntityStorageMutex.Lock()
	gits.RelationStorageMutex.Lock()

	// lets start recursive mapping of the data
	archivist.DebugF("MapTransportDataWithContext ", data)
	ret := mapRecursive(data, -1, -1, gits.DIRECTION_NONE, context, false, nil)

	// now we unlock all the mutexes again
	gits.RelationStorageMutex.Unlock()
	gits.EntityStorageMutex.Unlock()
	gits.EntityTypeMutex.Unlock()

	return ret
}

func MapTransportDataWithContextForceCreate(data transport.TransportEntity, context string) transport.TransportEntity {
	// first we lock all the storages
	gits.EntityTypeMutex.Lock()
	gits.EntityStorageMutex.Lock()
	gits.RelationStorageMutex.Lock()

	// lets start recursive mapping of the data
	archivist.DebugF("MapTransportDataWithContextForceCreate ", data)
	ret := mapRecursive(data, -1, -1, gits.DIRECTION_NONE, context, true, nil)

	// now we unlock all the mutexes again
	gits.RelationStorageMutex.Unlock()
	gits.EntityStorageMutex.Unlock()
	gits.EntityTypeMutex.Unlock()

	return ret
}

func MapTransportDataForceCreate(data transport.TransportEntity) transport.TransportEntity {
	// first we lock all the storages
	gits.EntityTypeMutex.Lock()
	gits.EntityStorageMutex.Lock()
	gits.RelationStorageMutex.Lock()

	// lets start recursive mapping of the data
	archivist.DebugF("MapTransportDataForceCreate ", data)
	ret := mapRecursive(data, -1, -1, gits.DIRECTION_NONE, "", true, nil)

	// now we unlock all the mutexes again

	gits.RelationStorageMutex.Unlock()
	gits.EntityStorageMutex.Unlock()
	gits.EntityTypeMutex.Unlock()

	return ret
}

type RecursiveMapCtx struct {
	SourceEntity   *transport.TransportEntity
	SourceRelation *transport.TransportRelation
	Direction      int
}

func mapRecursive(entity transport.TransportEntity, relatedType int, relatedID int, direction int, overwriteContext string, forceCreate bool, ctx *RecursiveMapCtx) transport.TransportEntity {
	// first we get the right TypeID
	var TypeID int
	var err error
	var newEntity types.StorageEntity
	createEntity := false
	TypeID, err = gits.GetTypeIdByStringUnsafe(entity.Type)
	if nil != err {
		TypeID, _ = gits.CreateEntityTypeUnsafe(entity.Type)
	}

	// now we check if its a forceCreate. If yes we gonne overwrite
	// the entity.ID with -1
	if forceCreate {
		entity.ID = -1
	}

	var mapID int
	// lets see if an ID was given, if not its a new entity to be created
	//if !reflect.ValueOf(entity).FieldByName("ID").IsValid() { // ### revalidate why this doesnt work should hit on non-constructed struct field - maybe 1.9 vs 2.0 issue?
	if -1 == entity.ID {
		// now we create the fitting entity
		newEntity = types.StorageEntity{
			ID:         -1,
			Type:       TypeID,
			Value:      entity.Value,
			Context:    entity.Context,
			Version:    1,
			Properties: util.CopyStringStringMap(entity.Properties),
		}
		// now we create the entity
		createEntity = true
		// ##### mapID, _ = gits.CreateEntityUnsafe(newEntity)
		entity.Properties["bMap"] = ""
	} else if -2 == entity.ID {
		tmp, _ := gits.GetEntitiesByTypeAndValueUnsafe(entity.Type, entity.Value, "match", "")
		if 0 < len(tmp) {
			mapID = tmp[0].ID
		} else {
			// now we create the fitting entity
			newEntity = types.StorageEntity{
				ID:         -1,
				Type:       TypeID,
				Value:      entity.Value,
				Context:    entity.Context,
				Version:    1,
				Properties: util.CopyStringStringMap(entity.Properties),
			}
			// now we create the entity
			createEntity = true
			entity.Properties["bMap"] = ""
		}
	} else if 0 == entity.ID {
		// checking if a source related entity exists with the same value, if yes map onto it, if not create it.
		// this could be extended by using structural uniqueness defintions of we see those are neccesary
		rEntity, hit, _ := getRelatedEntityWithTypeAndValue(entity, TypeID, relatedType, relatedID, direction) // ### skipping error value as _ for now, recheck later
		// we found a related entity with the same value so we use it for further processing
		if hit {
			mapID = rEntity.ID
			TypeID, _ = gits.GetTypeIdByStringUnsafe(rEntity.Type) // ### error supressed because we get the type of an entity we just retrieved while in a locked state errors should be impossible
		} else {
			// there is no related entity existing with the given value so we create one
			newEntity = types.StorageEntity{
				ID:         -1,
				Type:       TypeID,
				Value:      entity.Value,
				Context:    entity.Context,
				Version:    1,
				Properties: util.CopyStringStringMap(entity.Properties),
			}
			// now we create the entity and store the mapID
			createEntity = true
			// ##### temporary disabled mapID, _ = gits.CreateEntityUnsafe(newEntity)
			// also we add a property to the entity that allows us to identify this as a new mapped entity. this is
			// necessary for the scheduler later on. maybe solved different later ### (thought about running a diff
			// between input data and data returned from the mapping tho this seems a lot more computation heavy
			// than adding some kind of flag that the scheduler later can use to identify the diff - that way it only
			// has to walk the return data and not deep-compare two n dimension object nesting)
			// we dont really need a value for now just a property existing that we can check for
			entity.Properties["bMap"] = ""
		}
	} else {
		// it seems we got an already existing entity given, so we use this id to map
		mapID = entity.ID
	}

	// is there a new entity that has to be created?
	if createEntity {
		// do we need to adjust the Context?
		if "" != overwriteContext {
			newEntity.Context = overwriteContext
		}
		mapID, _ = gits.CreateEntityUnsafe(newEntity)
	}

	// lets map the child elements
	if len(entity.ChildRelations) != 0 {
		// there are children lets iteater over
		// the map
		for key, childRelation := range entity.ChildRelations {
			// pas the child entity and the parent coords to
			// create the relation after inserting the entity
			entity.ChildRelations[key].Target = mapRecursive(childRelation.Target, TypeID, mapID, gits.DIRECTION_CHILD, overwriteContext, forceCreate, &RecursiveMapCtx{
				SourceEntity:   &entity,
				SourceRelation: &entity.ChildRelations[key],
				Direction:      gits.DIRECTION_CHILD,
			})
		}
	}
	// than map the parent elements
	if len(entity.ParentRelations) != 0 {
		// there are children lets iteater over
		// the map
		for key, parentRelation := range entity.ParentRelations {
			// pas the child entity and the parent coords to
			// create the relation after inserting the entity
			entity.ParentRelations[key].Target = mapRecursive(parentRelation.Target, TypeID, mapID, gits.DIRECTION_PARENT, overwriteContext, forceCreate, &RecursiveMapCtx{
				SourceEntity:   &entity,
				SourceRelation: &entity.ParentRelations[key],
				Direction:      gits.DIRECTION_PARENT,
			})
		}
	}
	// now lets check if our parent Type and id
	// are not -1 , if so we need to create
	// a relation
	createdRelation := false
	if relatedType != -1 && relatedID != -1 {
		// lets create the relation to our parent
		if gits.DIRECTION_CHILD == direction {
			// first we make sure the relation doesnt already exist (because we allow mapped existing data inside a to map json)
			if !gits.RelationExistsUnsafe(relatedType, relatedID, TypeID, mapID) {
				tmpRelation := types.StorageRelation{
					SourceType: relatedType,
					SourceID:   relatedID,
					TargetType: TypeID,
					TargetID:   mapID,
					Version:    1,
				}
				gits.CreateRelationUnsafe(relatedType, relatedID, TypeID, mapID, tmpRelation)
				createdRelation = true
			}
		} else if gits.DIRECTION_PARENT == direction {
			// first we make sure the relation doesnt already exist (because we allow mapped existing data inside a to map json)
			if !gits.RelationExistsUnsafe(TypeID, mapID, relatedType, relatedID) {
				// or relation towards the child
				tmpRelation := types.StorageRelation{
					SourceType: TypeID,
					SourceID:   mapID,
					TargetType: relatedType,
					TargetID:   relatedID,
					Version:    1,
				}
				gits.CreateRelationUnsafe(TypeID, mapID, relatedType, relatedID, tmpRelation)
				createdRelation = true
			}
		}
	}

	// if we created a relation we gonne mark the relation with bmap so we can identify it late ron in the scheduler in terms of structural mapping
	if createdRelation {
		// add  only add such structures that should trigger a follow up job. but marking all new entities would be more correct
		if _, ok := ctx.SourceEntity.Properties["bMap"]; !ok && !createEntity {
			ctx.SourceRelation.Properties = map[string]string{"bMap": ""}
			archivist.Debug("Created relation from " + ctx.SourceEntity.Type + " to " + entity.Type + " without surrounding new entities")
		}
	}
	// only the first return is interesting since it
	// returns the most parent id
	entity.ID = mapID
	return entity
}

func getRelatedEntityWithTypeAndValue(entity transport.TransportEntity, entityTypeID int, relatedType int, relatedID int, direction int) (transport.TransportEntity, bool, error) {
	archivist.Debug("Trying to find related entity by type, value and related addr", entity, entityTypeID, relatedType, relatedID, direction)
	entities, err := gits.GetEntitiesByTypeAndValueUnsafe(entity.Type, entity.Value, "match", "")
	archivist.Debug("Followin entities retrieved by type and value", entities)

	// we skipping on error here. this errors can only appear if given entity.Type doesnt exist.
	// maybe recheck later ### | also hitting this case if we got 0 entities in return
	if nil != err || 0 == len(entities) {
		return transport.TransportEntity{}, false, err
	}

	// now we check if we even have a related entity or if this is the first in line
	// if there is no related we have to assume a correct one - this we do by just taking
	// the first available result. may be reviewd later on ###
	if -1 == relatedType && -1 == relatedID {
		return transport.TransportEntity{
			Type:       entity.Type,
			ID:         entities[0].ID,
			Value:      entities[0].Value,
			Context:    entities[0].Context,
			Properties: entities[0].Properties,
		}, true, nil
	}

	// ok we actually have to check if its related to the given related type/id
	// so we iterate over the results and check if one of them is linked to the related
	var alphaType, alphaID, betaType, betaID int
	for _, resultEntity := range entities {
		if gits.DIRECTION_PARENT == direction {
			alphaType = entityTypeID
			alphaID = resultEntity.ID
			betaType = relatedType
			betaID = relatedID
		} else {
			alphaType = relatedType
			alphaID = relatedID
			betaType = entityTypeID
			betaID = resultEntity.ID
		}
		if gits.RelationExistsUnsafe(alphaType, alphaID, betaType, betaID) {
			return transport.TransportEntity{
				Type:       entity.Type,
				ID:         resultEntity.ID,
				Value:      resultEntity.Value,
				Context:    resultEntity.Context,
				Properties: resultEntity.Properties,
			}, true, nil
		}
	}
	return transport.TransportEntity{}, false, nil
}
