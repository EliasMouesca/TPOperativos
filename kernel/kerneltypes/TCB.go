package kerneltypes

type TCB struct {
	TID       int
	Prioridad int
	ConectPCB *PCB
}
