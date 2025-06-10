package main

import (
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits/src/storage"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/go-cyberbrain"
	"github.com/voodooEntity/go-cyberbrain/src/example"
	"github.com/voodooEntity/go-cyberbrain/src/system/cerebrum"
	"github.com/voodooEntity/go-cyberbrain/src/system/interfaces"
)

func main() {
	// create base instance. ident is required.
	// NeuronAmount will default back to
	// runtime.NumCPU == num logical cpu's
	cb := cyberbrain.New(cyberbrain.Settings{
		NeuronAmount: 1,
		Ident:        "GreatName",
	})

	// register actions
	cb.RegisterAction("resolveIPFromDomain", func() interfaces.ActionInterface {
		return example.New()
	})

	// start the neurons
	cb.Start()

	// Learn data and schedule based on it
	cb.LearnAndSchedule(transport.TransportEntity{
		ID:         storage.MAP_FORCE_CREATE,
		Type:       "Domain",
		Value:      "laughingman.dev",
		Context:    "example code",
		Properties: map[string]string{},
	})

	// get an observer instance. provide a callback
	// to be executed at the end and lethal=true
	// which stops the cyberbrain at the end
	obsi := cb.GetObserverInstance(func(mi *cerebrum.Memory) {
		qry := mi.Gits.Query().New().Read("IP")
		ret := mi.Gits.Query().Execute(qry)
		archivist.Info("Result:", ret)
	}, true)

	// blocking while neurons are
	// working & non-finished jobs exist
	obsi.Loop()
}
