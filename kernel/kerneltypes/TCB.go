package kerneltypes

import "github.com/sisoputnfrba/tp-golang/types"

type TCB struct {
	// TID del hilo
	TID types.Tid

	// Prioridad del hilo
	Prioridad int

	// El PCB del proceso al que corresponde el hilo
	FatherPCB *PCB

	// Mutex que está lockeando el hilo
	Mutex []*MutexWrapper

	// El hilo joineado por este (pidió bloquearse hasta que <JoinedTCB> termine)
	JoinedTCB *TCB
}
