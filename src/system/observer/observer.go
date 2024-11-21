package observer

import (
	"encoding/json"
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/go-cyberbrain/src/system/core"
	"github.com/voodooEntity/go-cyberbrain/src/system/job"
	"github.com/voodooEntity/go-cyberbrain/src/system/util"
	"strconv"
	"time"
)

type Observer struct {
	RootType          string
	RootID            int
	InactiveIncrement int
	gitsInstance      *gits.Gits
	Runners           []Tracker
}

type Tracker struct {
	ID      int
	Version int
}

func New(rootType string, rootID int) *Observer {
	archivist.Info("Creating observer")
	gi := gits.GetDefault()
	var runners []Tracker
	qry := query.New().Read("Runner")
	res := gi.Query().Execute(qry)
	for _, val := range res.Entities {
		runners = append(runners, Tracker{
			ID:      val.ID,
			Version: val.Version,
		})
	}

	return &Observer{
		RootType:          rootType,
		RootID:            rootID,
		InactiveIncrement: 0,
		gitsInstance:      gi,
		Runners:           runners,
	}
}

func (self *Observer) Loop() {
	for !self.ReachedEndgame() {
		archivist.Debug("Observer looping:")
		time.Sleep(100 * time.Millisecond)
	}
	self.Endgame()
	archivist.Info("Cyberbrain has been shutdown, runner exiting")
}

func (self *Observer) ReachedEndgame() bool {
	runnerQry := query.New().Read("Runner").Match("Properties.State", "==", "Searching")
	sysRunners := self.gitsInstance.Query().Execute(runnerQry)
	archivist.Debug("Observer: searching runners", sysRunners.Amount)
	archivist.Debug("Observer: total amount created runners", len(core.Runners))
	openJobs := job.GetOpenJobs()
	if openJobs.Amount == 0 && sysRunners.Amount == len(core.Runners) {
		changedVersion := false
		for _, sysRunner := range sysRunners.Entities {
			for tid, tracker := range self.Runners {
				if sysRunner.ID == tracker.ID {
					if sysRunner.Version != tracker.Version {
						changedVersion = true
						self.Runners[tid].Version = sysRunner.Version
					}
				}
			}
		}
		if changedVersion {
			self.InactiveIncrement = 0
			return false
		}
		if self.InactiveIncrement > 5 {
			return true
		}
		self.InactiveIncrement++
		return false
	}
	self.InactiveIncrement = 0
	return false
}

func (self *Observer) Endgame() {
	archivist.Info("executing endgame")
	util.Shutdown()
	for !self.AllRunnersDead() {
		time.Sleep(10 * time.Millisecond)
	}
	finalDataQuery := query.New().Read(self.RootType).Match("ID", "==", strconv.Itoa(self.RootID)).TraverseOut(100).TraverseIn(100)
	finalData := self.gitsInstance.Query().Execute(finalDataQuery)
	b, err := json.Marshal(finalData)
	if err != nil {
		archivist.Error("Error marshalling final data", err)
		return
	}
	archivist.Info("Result " + string(b))
}

func (self *Observer) AllRunnersDead() bool {
	qry := query.New().Read("Runner").Match("Properties.State", "==", "Dead")
	runners := self.gitsInstance.Query().Execute(qry)
	if runners.Amount == len(core.Runners) {
		return true
	}
	return false
}
