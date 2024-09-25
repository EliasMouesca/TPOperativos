package kerneltypes

type Mutex struct {
	// El nombre por el que se lo conoce al mutex desde el pseudocódigo (léase CPU MUTEX_CREATE RECURSO_1)
	Name string

	// ID del hilo que tiene el mutex asignado
	AssignedTCB *TCB

	// Lista de hilos bloqueados esperando este mutex
	BlockedTCBs []*TCB
}

func (a *Mutex) Equal(b *Mutex) bool {
	return a.Name == b.Name
}
