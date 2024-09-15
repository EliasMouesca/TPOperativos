package kerneltypes

type ShortTermSchedulerInterface interface {
	Planificar() (TCB, error)
	AddToReady(TCB) error
}
