package kerneltypes

type ShortTermSchedulerInterface interface {
	Planificar() (TCB, error)
	AddToReady(TCB) error
	//ThreadExists(int) bool  // TODO: LARGO PLAZO NECESITA PARA LAS SYSCALLS!!!!!
	//ThreadRemove(int) error // ESTE TMB
}
