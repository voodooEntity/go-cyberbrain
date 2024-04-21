package observer

import (
	"encoding/json"
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits/src/query"
	"github.com/voodooEntity/go-cyberbrain/src/system/core"
	"github.com/voodooEntity/go-cyberbrain/src/system/util"
	"strconv"
	"time"
)

type Observer struct {
	RootType          string
	RootID            int
	InactiveIncrement int
}

func New(rootType string, rootID int) *Observer {
	archivist.Info("Creating observer")
	return &Observer{
		RootType:          rootType,
		RootID:            rootID,
		InactiveIncrement: 0,
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
	runners := query.Execute(runnerQry)
	archivist.Debug("Observer: searching runners", runners.Amount)
	archivist.Debug("Observer: total amount created runners", len(core.Runners))
	if runners.Amount == len(core.Runners) {
		archivist.Debug("Observer: current inactiveIncrement", self.InactiveIncrement)
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
	finalData := query.Execute(finalDataQuery)
	b, err := json.Marshal(finalData)
	if err != nil {
		archivist.Error("Error marshalling final data", err)
		return
	}
	archivist.Info("Result " + string(b))
}

func (self *Observer) AllRunnersDead() bool {
	qry := query.New().Read("Runner").Match("Properties.State", "==", "Dead")
	runners := query.Execute(qry)
	if runners.Amount == len(core.Runners) {
		return true
	}
	return false
}
