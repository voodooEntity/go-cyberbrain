package main

import (
	"fmt"
	"github.com/voodooEntity/gits/src/storage"
	"github.com/voodooEntity/gits/src/transport"
	cyberbrain "github.com/voodooEntity/go-cyberbrain"
	"github.com/voodooEntity/go-cyberbrain/src/system/cerebrum"
	"github.com/voodooEntity/go-cyberbrain/src/system/interfaces"
	"github.com/voodooEntity/go-cyberbrain/src/test"
)

func main() {
	cb := cyberbrain.New(cyberbrain.Settings{
		NeuronAmount: 1,
		Ident:        "GreatName",
	})
	cb.RegisterAction("resolveIPFromDomain", func() interfaces.ActionInterface {
		return test.New()
	})
	cb.Start()

	// mappedData :=
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
		fmt.Println("Result:" + fmt.Sprintf("%+v", ret))
	}, true)

	// blocking while neurons are
	// working & non-finished jobs exist
	obsi.Loop()

	fmt.Println("finished")
}
