package kerneltypes

import "github.com/sisoputnfrba/tp-golang/types"

type PCB struct {
	// PID del proceso
	PID types.Pid

	// Lista de los TIDS asociados a este proceso
	TIDs []types.Tid

	// Lista de los mutex creados para el proceso
	CreatedMutexes []*MutexWrapper
}
