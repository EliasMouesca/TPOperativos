package kerneltypes

import "sync"

type MutexWrapper struct {
	// El nombre por el que se lo conoce al mutex desde el pseudocódigo (léase CPU MUTEX_CREATE RECURSO_1)
	Name string

	// El mutex original de la librería
	Mutex sync.Mutex

	// ID del hilo que tiene el mutex asignado
	AssignedTCB *TCB

	// Lista de hilos bloqueados esperando este mutex
	BlockedTCBs []*TCB
}
