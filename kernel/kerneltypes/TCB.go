package kerneltypes

import (
	"github.com/sisoputnfrba/tp-golang/types"
	"time"
)

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

	// Instante en el que el thread entró a la CPU, de la ultima vez
	ExecInstant time.Time

	// Instante en el que el thread salió de la CPU, de la ultima vez
	ExitInstant time.Time

	// Quantum restante en nanosegundos (eeem duration de go, revisar la docu)
	// mucho cuidado! no cambiar
	QuantumRestante time.Duration

	// True si el proceso se fue de la CPU antes de que se acabe su quantum
	HayQuantumRestante bool
}

func (a *TCB) Null() *TCB {
	return nil
}

func (a *TCB) Equal(b *TCB) bool {
	return a.TID == b.TID && a.FatherPCB.Equal(b.FatherPCB)
}
