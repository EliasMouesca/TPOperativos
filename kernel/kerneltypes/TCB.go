package kerneltypes

import "github.com/sisoputnfrba/tp-golang/types"

type TCB struct {
	// TID del hilo
	TID types.Tid

	// Prioridad del hilo
	Prioridad int

	// El PCB del proceso al que corresponde el hilo
	FatherPCB *PCB

	// LockedMutexes mutexes que está lockeando el hilo
	LockedMutexes []*Mutex

	// El hilo joineado por este (pidió bloquearse hasta que <JoinedTCB> termine)
	JoinedTCB *TCB
}

func (a *TCB) New() *TCB {
	return nil
}

func (a *TCB) Equal(b *TCB) bool {
	return a.TID == b.TID && a.FatherPCB.Equal(b.FatherPCB)
}
