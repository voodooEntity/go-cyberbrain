package runner

import (
	"encoding/json"
	"errors"
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/go-cyberbrain/src/system/job"
	"github.com/voodooEntity/go-cyberbrain/src/system/mapper"
	"github.com/voodooEntity/go-cyberbrain/src/system/registry"
	"github.com/voodooEntity/go-cyberbrain/src/system/scheduler"
	"github.com/voodooEntity/go-cyberbrain/src/system/util"
	"strconv"
	"time"
)

const (
	INTERCOM_BUFF_SIZE   int = 100000
	INTERCOM_INPUT_CHAN  int = 0
	INTERCOM_OUTPUT_CHAN int = 1
)

type Runner struct {
	id             int
	uid            string
	intercom       [2]chan string
	actionRegistry registry.Registry
	job            job.Job
}

func New(id int, taskRegistry registry.Registry) *Runner {
	archivist.Info("Creating runner", id)
	properties := make(map[string]string)
	properties["State"] = "Searching"
	mapper.MapTransportData(transport.TransportEntity{
		ID:         -1,
		Type:       "Runner",
		Value:      strconv.Itoa(id),
		Context:    "Bezel",
		Properties: properties,
	})

	return &Runner{
		id:             id,
		intercom:       [2]chan string{make(chan string, INTERCOM_BUFF_SIZE), make(chan string, INTERCOM_BUFF_SIZE)},
		actionRegistry: taskRegistry,
	}
}

func (self *Runner) Loop() {
	for util.IsActive() {
		archivist.Debug("Runner looping id: ", self.id)
		// lets try to assign a job
		if self.FindJob() {
			// now we gonne try execute the just assigned job
			results, err := self.ExecuteJob()
			// did it work out?
			if nil != err {
				self.FinishJobError(err)
				// ### todo think about how to handle errors
			} else {
				// now we map the data into our storage and run the result of the mapper
				// through our scheduler to create new Jobs based on what we just learned
				self.FinishJobSuccess(results)
			}
		} else {
			//			time.Sleep(1000000000)
			time.Sleep(100 * time.Millisecond)
		}
		//time.Sleep(time.Second * 4)
	}
	archivist.Info("Bezel has been shutdown, runner exiting")
}

func (self *Runner) FindJob() bool {
	// query can be optimized by joining ###todo
	jobList := job.GetOpenJobs()
	archivist.Debug("Open Jobs found", jobList)
	// if there are any jobs
	if 0 < jobList.Amount {
		// iterate through them
		for _, jobEntity := range jobList.Entities[0].Parents() {
			// load the full job data as instance of job struct
			newJob := job.Load(jobEntity.ID)
			if nil != newJob {
				// finally assign the job
				ok := self.AssignJob(newJob)
				if ok {
					return true
				}
			}
		}
	}
	// could not assign a job
	return false
}

func (self *Runner) ExecuteJob() ([]transport.TransportEntity, error) {
	qry := query.New().Read("Runner").Match("Value", "==", strconv.Itoa(self.id)).To(query.New().Read("Job").To(query.New().Read("Input")))
	ret := query.Execute(qry)

	if 0 == ret.Amount {
		return []transport.TransportEntity{}, errors.New("Runner could not find any assigned job. This should be rather impossible")
	}

	// convert json input to actual struct instance
	inputJson := ret.Entities[0].Children()[0].Children()[0].Properties["Data"]
	var inputEntity transport.TransportEntity
	err := json.Unmarshal([]byte(inputJson), &inputEntity)
	if nil != err {
		archivist.Error("Job: "+ret.Entities[0].Children()[0].Value+" - could not convert job input json back to struct data", inputJson)
		return []transport.TransportEntity{}, err
	}

	// retrieve the action from taskRegistry and apply it
	jobAction := self.actionRegistry[ret.Entities[0].Children()[0].Properties["Action"]]
	archivist.Info("Runner " + strconv.Itoa(self.id) + " executing action " + jobAction.GetName() + " with Job " + ret.Entities[0].Children()[0].Value)
	// clear bMap properties from inputEntity, so we don't endless run
	rRemovebMap(inputEntity)
	// finally we execute it
	results, err := jobAction.GetPlugin().Execute(inputEntity, ret.Entities[0].Children()[0].Properties["Requirement"], "Runner")
	if nil != err {
		return []transport.TransportEntity{}, errors.New("Job: " + ret.Entities[0].Children()[0].Value + " execution failed with error " + err.Error())
	}
	archivist.Info("Job: " + ret.Entities[0].Children()[0].Value + " finished successfully")
	//archivist.Debug("Job result", results)
	return results, nil
}

func (self *Runner) AssignJob(newJob *job.Job) bool {
	// update runners status to Assigning...
	self.ChangeState("Assigning")

	// letes see if we can assign that job
	ok := newJob.AssignToRunner(self.id)
	if !ok {
		// job could not be assigned, lets think about what reasons this could have
		// for different error handlings. for now we just gonne log it
		archivist.Debug("Runner couldnt assign job: ", newJob.Data)

		// update runners status...
		self.ChangeState("Searching")

		return false
	}

	// update runners status...
	self.ChangeState("Working")

	// ... link the job to the runner
	qry := query.New().Find("Runner").Match(
		"Value",
		"==",
		strconv.Itoa(self.id),
	).Link("Job").Match(
		"ID",
		"==",
		strconv.Itoa(newJob.GetID()),
	)
	query.Execute(qry)

	self.job = *newJob

	return true
}

func (self *Runner) CheckChannel() {

}

func (self *Runner) GetInputIntercom() chan string {
	return self.intercom[INTERCOM_INPUT_CHAN]
}

func (self *Runner) GetOutputIntercom() chan string {
	return self.intercom[INTERCOM_OUTPUT_CHAN]
}

func (self *Runner) ChangeState(state string) {
	qry := query.New().Update("Runner").Match(
		"Value",
		"==",
		strconv.Itoa(self.id),
	).Set(
		"Properties.State",
		state,
	)
	query.Execute(qry)
}

func (self *Runner) FinishJobSuccess(results []transport.TransportEntity) {
	// going through the results
	for _, result := range results {
		archivist.Debug("Mapping result from job", result)
		mappedResult := mapper.MapTransportData(result)
		archivist.Debug("Running freshly mapped job return with scheduler", mappedResult)
		scheduler.Run(result, self.actionRegistry)
	}

	qry := query.New().Read("Runner").Match(
		"Value",
		"==",
		strconv.Itoa(self.id),
	).To(query.New().Read("Job"))

	runnerWithJob := query.Execute(qry)
	jobId := runnerWithJob.Entities[0].Children()[0].ID
	archivist.Debug("Detaching job from runner", runnerWithJob)
	qry = query.New().Unlink("Runner").Match("Value", "==", strconv.Itoa(self.id)).To(
		query.New().Find("Job").Match("ID", "==", strconv.Itoa(jobId)),
	)
	query.Execute(qry)
}

func (self *Runner) FinishJobError(err error) {
	archivist.Info("Ended job with error: ", err.Error())
	qry := query.New().Read("Runner").Match(
		"Value",
		"==",
		strconv.Itoa(self.id),
	).To(query.New().Read("Job"))

	runnerWithJob := query.Execute(qry)
	jobId := runnerWithJob.Entities[0].Children()[0].ID
	archivist.Debug("Detaching job from runner", runnerWithJob)
	qry = query.New().Unlink("Runner").Match("Value", "==", strconv.Itoa(self.id)).To(
		query.New().Find("Job").Match("ID", "==", strconv.Itoa(jobId)),
	)
	query.Execute(qry)
}

func rRemovebMap(entity transport.TransportEntity) {
	if _, ok := entity.Properties["bMap"]; ok {
		delete(entity.Properties, "bMap")
	}
	for _, val := range entity.ChildRelations {
		rRemovebMap(val.Target)
	}
	for _, val := range entity.ParentRelations {
		rRemovebMap(val.Target)
	}
}
