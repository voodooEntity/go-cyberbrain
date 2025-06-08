package cerebrum

import "github.com/voodooEntity/gits"

type Activity struct {
	Demultiplexer *Demultiplexer
	Scheduler     *Scheduler
}

type Consciousness struct {
	Memory   *Memory
	Cortex   *Cortex
	Activity *Activity
}

type Memory struct {
	Gits   *gits.Gits
	Mapper *Mapper
}
