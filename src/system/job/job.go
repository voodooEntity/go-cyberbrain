package job

import (
	"encoding/json"
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/gits/src/types"
	"github.com/voodooEntity/go-cyberbrain/src/system/mapper"
	"github.com/voodooEntity/go-cyberbrain/src/system/util"
	"strconv"
)

type Job struct {
	Data transport.TransportEntity
}

func Create(action string, requirement string, input transport.TransportEntity) *Job {
	jobProperties := make(map[string]string)
	jobProperties["Action"] = action
	jobProperties["Requirement"] = requirement
	inputProperties := make(map[string]string)
	inputJson, err := json.Marshal(input)
	if nil != err {
		archivist.Error("Could not create json input payload from input string", input)
		return &Job{}
	}
	inputProperties["Data"] = string(inputJson)

	mapped := mapper.MapTransportDataWithContextForceCreate(transport.TransportEntity{
		ID:         -1,
		Type:       "Job",
		Context:    "Bezel",
		Value:      util.UniqueID(),
		Properties: jobProperties,
		ChildRelations: []transport.TransportRelation{
			{
				Target: transport.TransportEntity{
					Type:       "Input",
					ID:         -1,
					Context:    "Bezel",
					Value:      util.UniqueID(),
					Properties: inputProperties,
				},
			},
			{
				Target: transport.TransportEntity{
					Type:       "State",
					Value:      "Open",
					Context:    "Bezel",
					ID:         0,
					Properties: make(map[string]string),
				},
			},
		},
	}, "System")
	archivist.Debug("Mapped new job", mapped)
	return &Job{}
}

func Load(id int) *Job {
	// make sure that job actually exists
	ret := query.Execute(query.New().Read("Job").Match("ID", "==", strconv.Itoa(id)).TraverseOut(30)) // ### add max depth config for job
	if 0 < ret.Amount {
		return &Job{
			Data: ret.Entities[0],
		}
	}
	return nil
}

func (self *Job) AssignToRunner(runnerID int) bool {
	// since we have to make sure we dont run into race conditions we gonne do some direct
	// api calls into gits here. we may change this at some point to query logics or something else
	// its not bad by design it just could be done smoother ###todo recheck if this cant be solved by qry where conditions
	gits.EntityStorageMutex.Lock()
	gits.RelationStorageMutex.Lock()
	jobTypeID, err := gits.GetTypeIdByStringUnsafe("Job")
	if nil != err {
		// this shouldn't be happening but rather handle every error than assume its impossible
		archivist.Debug("The impossible occured. Run you fools")
		gits.EntityStorageMutex.Unlock()
		gits.RelationStorageMutex.Unlock()
		return false
	}

	e, err := gits.GetEntityByPathUnsafe(jobTypeID, self.Data.ID, "")
	if nil != err {
		gits.EntityStorageMutex.Unlock()
		gits.RelationStorageMutex.Unlock()
		archivist.Debug("Runner tries to assign nonexisting job", self.Data.ID)
		return false
	}
	self.Data = transport.TransportEntity{
		Type:       "Job",
		ID:         e.ID,
		Context:    e.Context,
		Properties: e.Properties,
		Version:    e.Version,
	}

	childRelations, _ := gits.GetChildRelationsBySourceTypeAndSourceIdUnsafe(jobTypeID, e.ID, "")
	stateTypeID, _ := gits.GetTypeIdByStringUnsafe("State")
	for _, childRelation := range childRelations {
		if stateTypeID == childRelation.TargetType {
			openState, _ := gits.GetEntityByPathUnsafe(stateTypeID, childRelation.TargetID, "")
			archivist.Debug("state retrieved from job", openState)
			if "Open" != openState.Value {
				gits.EntityStorageMutex.Unlock()
				gits.RelationStorageMutex.Unlock()
				archivist.Debug("Runner tries to assign job that state is not Open", self.Data.ID)
				return false
			}
			// detach open state from job
			gits.DeleteRelationUnsafe(e.Type, e.ID, stateTypeID, openState.ID)
			// get assigned state entity
			assignedState, _ := gits.GetEntitiesByTypeAndValueUnsafe("State", "Assigned", "match", "System")
			archivist.Debug("assigned state entity", assignedState)
			// now we map the job to the assigned entity
			gits.CreateRelationUnsafe(e.Type, e.ID, stateTypeID, assignedState[0].ID, types.StorageRelation{
				SourceType: jobTypeID,
				SourceID:   e.ID,
				TargetType: stateTypeID,
				TargetID:   assignedState[0].ID,
				Context:    "Bezel",
				Properties: make(map[string]string),
			})
		}
	}

	// finally we assign the job to the runner
	//runnerTypeID, _ := gits.GetTypeIdByStringUnsafe("Runner")
	runnerEntity, _ := gits.GetEntitiesByTypeAndValueUnsafe("Runner", strconv.Itoa(runnerID), "match", "Bezel")
	archivist.Debug("Map runner to job", runnerEntity[0].Type, runnerEntity[0].ID, jobTypeID, e.ID)
	gits.CreateRelationUnsafe(runnerEntity[0].Type, runnerEntity[0].ID, jobTypeID, e.ID, types.StorageRelation{
		SourceType: runnerEntity[0].Type,
		SourceID:   runnerEntity[0].ID,
		TargetType: jobTypeID,
		TargetID:   e.ID,
		Context:    "Bezel",
		Properties: make(map[string]string),
	})

	gits.EntityStorageMutex.Unlock()
	gits.RelationStorageMutex.Unlock()
	return true
}

func (self *Job) GetState() string {
	ret := query.Execute(query.New().Read("Job").Match("ID", "==", strconv.Itoa(self.Data.ID)).To(query.New().Read("State")))
	if 0 < ret.Amount {
		return ret.Entities[0].Children()[0].Value
	}
	archivist.Error("Retrieving state of an non existing job , should actually not happen, jobid is ", self.GetID())
	return ""
}

func (self *Job) ChangeState(newState string) {
	// ###todo implement consider timing issues
}

func (self *Job) GetID() int {
	return self.Data.ID
}

func GetOpenJobs() transport.Transport {
	qry := query.New().Read("State").Match("Value", "==", "Open").Match("Context", "==", "System").From(
		query.New().Read("Job").Match("Context", "==", "System"))
	return query.Execute(qry)
}
