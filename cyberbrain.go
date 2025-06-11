package cyberbrain

import (
	"errors"
	"fmt"
	"github.com/voodooEntity/archivist"
	"github.com/voodooEntity/gits"
	"github.com/voodooEntity/gits/src/transport"
	"github.com/voodooEntity/go-cyberbrain/src/system/cerebrum"
	"github.com/voodooEntity/go-cyberbrain/src/system/interfaces"
	"github.com/voodooEntity/go-cyberbrain/src/system/observer"
	"github.com/voodooEntity/go-cyberbrain/src/system/util"
	"runtime"
)

type Cyberbrain struct {
	ident        string
	neuronAmount int
	neurons      []int
	con          cerebrum.Consciousness
	initCfg      Settings
}

type Settings struct {
	Gits         *gits.Gits
	Ident        string
	NeuronAmount int
}

func New(cfg Settings) *Cyberbrain {
	// ident is required
	if cfg.Ident == "" {
		archivist.Error("no ident given")
		return nil
	}

	// setup the instance
	instance := &Cyberbrain{
		ident:        cfg.Ident,
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

//   - - - - - - - - - - - - - - - - - - - - - - - - - - - -
//     PUBLIC FUNCTIONS
//   - - - - - - - - - - - - - - - - - - - - - - - - - - - -

func (cb *Cyberbrain) GetGitsInstance() *gits.Gits {
	return cb.con.Memory.Gits
}

func (cb *Cyberbrain) Start() error {
	// make sure we dont start the same
	// cyberbrain instance twice
	if util.IsAlive(cb.con.Memory.Gits) {
		return errors.New("cyberbrain already running")
	}

	// set the "alife" dataset
	cb.bringToLife()

	// bootstrap our neurons
	cb.startNeurons()

	return nil
}

func (cb *Cyberbrain) Stop() error {
	if !util.IsAlive(cb.con.Memory.Gits) {
		return errors.New("can't stop not running cyberebrain")
	}

	util.Terminate(cb.con.Memory.Gits)

	return nil
}

func (cb *Cyberbrain) RegisterAction(actionName string, actionFactory func() interfaces.ActionInterface) error {
	if util.IsAlive(cb.con.Memory.Gits) {
		return errors.New("cyberbrain already running, can't register new actions")
	}
	cb.con.Cortex.RegisterAction(actionName, actionFactory)
	return nil
}

func (cb *Cyberbrain) LearnAndSchedule(data transport.TransportEntity) (transport.TransportEntity, error) {
	if !util.IsAlive(cb.con.Memory.Gits) {
		return transport.TransportEntity{}, errors.New("cyberbrain not running")
	}

	// than we learn and schedule
	learnedData, err := cb.Learn(data)
	if err != nil {
		return transport.TransportEntity{}, err
	}

	cb.Schedule(learnedData)

	// return the mapped data
	return learnedData, nil
}

func (cb *Cyberbrain) Learn(data transport.TransportEntity) (transport.TransportEntity, error) {
	if !util.IsAlive(cb.con.Memory.Gits) {
		return transport.TransportEntity{}, errors.New("cyberbrain not running")
	}

	// store the new data
	return cb.con.Memory.Mapper.MapTransportDataWithContext(data, "Data"), nil
}

func (cb *Cyberbrain) Schedule(data transport.TransportEntity) error {
	if !util.IsAlive(cb.con.Memory.Gits) {
		return errors.New("cyberbrain not running")
	}

	cb.con.Activity.Scheduler.Run(data, cb.con.Cortex)
	return nil
}

func (cb *Cyberbrain) GetObserverInstance(callback func(memoryInstance *cerebrum.Memory), lethal bool) *observer.Observer {
	return observer.New(cb.con.Memory, cb.neuronAmount, callback, lethal)
}

//   - - - - - - - - - - - - - - - - - - - - - - - - - - - -
//     INTERNAL FUNCTIONS
//   - - - - - - - - - - - - - - - - - - - - - - - - - - - -
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
	fmt.Print("yes we get here")
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
