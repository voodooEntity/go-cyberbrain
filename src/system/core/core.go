package core

import (
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/gits/src/types"
	"github.com/voodooEntity/go-cyberbrain/src/system/config"
	"github.com/voodooEntity/go-cyberbrain/src/system/mapper"
	"github.com/voodooEntity/go-cyberbrain/src/system/registry"
	"github.com/voodooEntity/go-cyberbrain/src/system/runner"
	"runtime"
)

func Init(configs map[string]string) {
	// init the gits storage
	gits.Init(types.PersistenceConfig{
		PersistenceChannelBufferSize: 10000000,
		Active:                       false,
		RotationEntriesMax:           100000,
	})

	// then we populate the action registry
	registry.Data = registry.New().Index()

	// we gonne find a better place for this
	createNecessaryEntities()

	// bring it to life
	bringToLife()

	// now we start our runners
	startRunners()

	// keep the main thread alife
	//for {
	// Sleep for a short duration to avoid consuming excessive CPU
	//	time.Sleep(time.Second)
	//}
}

func startRunners() {
	cpuAmount := runtime.NumCPU()
	//cpuAmount = 1
	for i := 0; i < cpuAmount; i++ {
		instance := runner.New(i, registry.Data)
		go instance.Loop()
	}
}

func handleFlags(flags map[string]string) {
	if 0 < len(flags) {
		for key, val := range flags {
			config.Set(key, val) ///###todo think about maybe printing errors here.
		}
	}
}

func bringToLife() {
	properties := make(map[string]string)
	properties["State"] = "Alive"
	mapper.MapTransportData(transport.TransportEntity{
		Type:       "AI",
		Value:      "Bezel",
		Context:    "System",
		Properties: properties,
	})
}

func createNecessaryEntities() {
	// create Open state
	gits.MapTransportData(transport.TransportEntity{
		ID:         0,
		Type:       "State",
		Value:      "Open",
		Context:    "System",
		Properties: make(map[string]string),
	})
	// create Assigned state
	gits.MapTransportData(transport.TransportEntity{
		ID:         0,
		Type:       "State",
		Value:      "Assigned",
		Context:    "System",
		Properties: make(map[string]string),
	})

}
