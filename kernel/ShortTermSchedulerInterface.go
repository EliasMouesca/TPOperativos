package main

type ShortTermSchedulerInterface interface {
	Planificar() (TCB, error)
	AddToReady(TCB) error
}

var AlgorithmsMap = map[string]ShortTermSchedulerInterface{
	"FIFO": &Fifo{},
	"P":    &Prioridades{},
	//"CMM":  CortoPlazo.CMM{},
}
