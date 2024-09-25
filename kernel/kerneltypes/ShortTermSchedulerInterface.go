package kerneltypes

type ShortTermSchedulerInterface interface {
	Planificar() (*TCB, error)
	AddToReady(*TCB) error
	ThreadExists(int, int) (bool, error) // Si existe -> true, sino -> false 		// RECIBE (TID, PID)
	ThreadRemove(int, int) error         // SACA DE READY (si existe, sino, error)	// RECIBE (TID, PID)
}
