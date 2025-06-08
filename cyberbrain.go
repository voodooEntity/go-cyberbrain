package cyberbrain

import (
	"errors"
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/go-cyberbrain/src/system/cerebrum"
	"github.com/voodooEntity/go-cyberbrain/src/system/interfaces"
	"github.com/voodooEntity/go-cyberbrain/src/system/observer"
	"runtime"
)

type Cyberbrain struct {
	ident        string
	neuronAmount int
	isRunning    bool
	neurons      []int
	con          cerebrum.Consciousness
	initCfg      Settings
	//observer *observer.Observer
}

type Settings struct {
	Gits         *gits.Gits
	Ident        string
	NeuronAmount int
}

func New(cfg Settings) *Cyberbrain {
	// setup the instance
	instance := &Cyberbrain{
		ident:        cfg.Ident,
		isRunning:    false,
		con:          cerebrum.Consciousness{},
		neuronAmount: runtime.NumCPU(), // default neuron amount is num logical cpus
		neurons:      make([]int, 0),
	}

	// if the given neuronAmount is
	// a positive >0 int
	if cfg.NeuronAmount > 0 {
		instance.neuronAmount = cfg.NeuronAmount
	}

	// first we prepare the memory
	instance.setupMemory(cfg.Gits)

	// with memory setup we can
	// setup the cortex
	instance.con.Cortex = cerebrum.NewCortex(instance.con.Memory)

	// now the "activity"
	instance.setupActivity()

	// finally create some necessary
	// datasets
	instance.createNecessaryEntities()

	return instance
}

func (cb *Cyberbrain) GetGitsInstance() *gits.Gits {
	return cb.con.Memory.Gits
}

func (cb *Cyberbrain) Start() error {
	// make sure we dont start the same
	// cyberbrain instance twice
	if cb.isRunning {
		return errors.New("cyberbrain already running")
	}
	cb.isRunning = true

	// set the "alife" dataset
	cb.bringToLife()

	// bootstrap our neurons
	cb.startNeurons()

	return nil
}

func (cb *Cyberbrain) RegisterAction(actionName string, actionFactory func() interfaces.ActionInterface) error {
	if cb.isRunning == true {
		return errors.New("cyberbrain already running, cant create new actions")
	}
	cb.con.Cortex.RegisterAction(actionName, actionFactory)
	return nil
}

func (cb *Cyberbrain) LearnAndSchedule(data transport.TransportEntity) transport.TransportEntity {
	learnedData := cb.Learn(data)
	cb.Schedule(learnedData)
	return learnedData
}

func (cb *Cyberbrain) Learn(data transport.TransportEntity) transport.TransportEntity {
	// store the new data
	return cb.con.Memory.Mapper.MapTransportDataWithContext(data, "Data")
}

func (cb *Cyberbrain) Schedule(data transport.TransportEntity) {
	cb.con.Activity.Scheduler.Run(data, cb.con.Cortex)
}

func (cb *Cyberbrain) setupActivity() {
	activities := cerebrum.Activity{}

	// first demultiplexer
	activities.Demultiplexer = cerebrum.NewDemultiplexer()

	// than scheduler
	activities.Scheduler = cerebrum.NewScheduler(cb.con.Memory, activities.Demultiplexer)

	// finally store it
	cb.con.Activity = &activities
}

func (cb *Cyberbrain) setupMemory(customGits *gits.Gits) {
	// dispatch gits instance and autocreate
	// if not given
	gitsInstance := customGits
	if nil == gitsInstance {
		gitsInstance = gits.NewInstance(cb.ident)
	}

	// based on the gits instance we gonne
	// bootstrap the mapper
	mapperInstance := cerebrum.NewMapper(gitsInstance)

	// combine to memory and store
	cb.con.Memory = &cerebrum.Memory{
		Mapper: mapperInstance,
		Gits:   gitsInstance,
	}
}

func (cb *Cyberbrain) startNeurons() {
	for i := 0; i < cb.neuronAmount; i++ {
		instance := cerebrum.NewNeuron(i, cb.con.Cortex, cb.con.Memory, cb.con.Activity)
		go instance.Loop()
		cb.neurons = append(cb.neurons, i)
	}
}

func (cb *Cyberbrain) bringToLife() {
	properties := make(map[string]string)
	properties["State"] = "Alive"
	cb.con.Memory.Mapper.MapTransportData(transport.TransportEntity{
		Type:       "AI",
		Value:      "Bezel",
		Context:    "System",
		Properties: properties,
	})
}

func (cb *Cyberbrain) createNecessaryEntities() {
	// create Open state
	cb.con.Memory.Gits.MapData(transport.TransportEntity{
		ID:         0,
		Type:       "State",
		Value:      "Open",
		Context:    "System",
		Properties: make(map[string]string),
	})
	// create Assigned state
	cb.con.Memory.Gits.MapData(transport.TransportEntity{
		ID:         0,
		Type:       "State",
		Value:      "Assigned",
		Context:    "System",
		Properties: make(map[string]string),
	})
}

func (cb *Cyberbrain) GetObserverInstance(callback func(memoryInstance *cerebrum.Memory), lethal bool) *observer.Observer {
	return observer.New(cb.con.Memory, cb.neuronAmount, callback, lethal)
}

//func (cb *Cyberbrain) StartContinouus() {
//	gitsapiConfig.Init(map[string]string{})
//
//	// init the archivist logger ### maybe will access a different config later on
//	// prolly should access the one of bezel not the gitsapi one. for now, we gonne stick with it ###
//	archivist.Init(gitsapiConfig.GetValue("LOG_LEVEL"), gitsapiConfig.GetValue("LOG_TARGET"), gitsapiConfig.GetValue("LOG_PATH"))
//
//	// initing some additional application specific endpoints
//	api.Extend()
//
//	// start the actual gitsapi
//	gitsapi.Start()
//}
