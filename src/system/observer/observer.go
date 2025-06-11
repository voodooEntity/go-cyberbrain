package observer

import (
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/go-cyberbrain/src/system/cerebrum"
	"github.com/voodooEntity/go-cyberbrain/src/system/util"
	"time"
)

type Observer struct {
	InactiveIncrement int
	memory            *cerebrum.Memory
	runnerAmount      int
	callback          func(memoryInstance *cerebrum.Memory)
	Runners           []Tracker
	lethal            bool
}

type Tracker struct {
	ID      int
	Version int
}

func New(memoryInstance *cerebrum.Memory, runnerAmount int, cb func(memoryInstance *cerebrum.Memory), lethal bool) *Observer {
	archivist.Info("Creating observer")
	var runners []Tracker
	qry := query.New().Read("Neuron")
	res := memoryInstance.Gits.Query().Execute(qry)
	for _, val := range res.Entities {
		runners = append(runners, Tracker{
			ID:      val.ID,
			Version: val.Version,
		})
	}

	return &Observer{
		InactiveIncrement: 0,
		memory:            memoryInstance,
		Runners:           runners,
		callback:          cb,
		runnerAmount:      runnerAmount,
		lethal:            lethal,
	}
}

func (o *Observer) Loop() {
	for !o.ReachedEndgame() {
		archivist.Debug("Observer looping:")
		time.Sleep(100 * time.Millisecond)
	}
	o.Endgame()
	archivist.Info("Cyberbrain has been shutdown, neuron exiting")
}

func (o *Observer) ReachedEndgame() bool {
	runnerQry := query.New().Read("Neuron").Match("Properties.State", "==", "Searching")
	sysRunners := o.memory.Gits.Query().Execute(runnerQry)
	archivist.Debug("Observer: searching neurons", sysRunners.Amount)
	archivist.Debug("Observer: total amount created neurons", o.runnerAmount)
	openJobs := cerebrum.GetOpenJobs(o.memory.Gits)
	if openJobs.Amount == 0 && sysRunners.Amount == o.runnerAmount {
		changedVersion := false
		for _, sysRunner := range sysRunners.Entities {
			for tid, tracker := range o.Runners {
				if sysRunner.ID == tracker.ID {
					if sysRunner.Version != tracker.Version {
						changedVersion = true
						o.Runners[tid].Version = sysRunner.Version
					}
				}
			}
		}
		if changedVersion {
			o.InactiveIncrement = 0
			return false
		}
		if o.InactiveIncrement > 5 {
			return true
		}
		o.InactiveIncrement++
		return false
	}
	o.InactiveIncrement = 0
	return false
}

func (o *Observer) Endgame() {
	archivist.Info("executing endgame")
	// if we are lethal we gonne stop cyberbrain
	if o.lethal {
		util.Terminate(o.memory.Gits)
		for !o.AllNeuronDead() {
			time.Sleep(10 * time.Millisecond)
		}
	}
	// execute callback with memory instance provided
	o.callback(o.memory)
}

func (o *Observer) AllNeuronDead() bool {
	qry := query.New().Read("Neuron").Match("Properties.State", "==", "Dead")
	runners := o.memory.Gits.Query().Execute(qry)
	if runners.Amount == o.runnerAmount {
		return true
	}
	return false
}
