package kerneltypes

import "sync"

type MutexWrapper struct {
	Mutex          sync.Mutex // El mutex original de la librería
	ID             int        // Identificador único del mutex
	AssignedTID    int        // ID del hilo que tiene el mutex asignado (0 si no está asignado)
	BlockedThreads []*TCB     // Lista de hilos bloqueados esperando este mutex
}
