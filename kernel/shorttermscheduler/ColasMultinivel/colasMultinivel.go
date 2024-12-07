package ColasMultinivel

import (
	"errors"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

type ColasMultiNivel struct {
	ReadyQueue  []*types.Queue[*kerneltypes.TCB]
	isRRRunning bool
}

func (cmm *ColasMultiNivel) Init() {
	go func() {
		err := EsperarYAvisarFinDeQuantum()
		logger.Error("Error: %v", err)
	}()
}

func (cmm *ColasMultiNivel) ThreadExists(tid types.Tid, pid types.Pid) (bool, error) {
	for _, queue := range cmm.ReadyQueue {
		for _, tcb := range queue.GetElements() {
			if tcb.TID == tid && tcb.FatherPCB.PID == pid {
				return true, nil
			}
		}
	}
	return false, errors.New("hilo no encontrado o no pertenece al PCB con PID especificado")
}

func (cmm *ColasMultiNivel) ThreadRemove(tid types.Tid, pid types.Pid) error {
	existe, _ := cmm.ThreadExists(tid, pid)
	if !existe {
		return errors.New("no se pudo eliminar el hilo con TID especificado o no pertenece al PCB con PID especificado")
	}

	for _, queue := range cmm.ReadyQueue {
		queueSize := queue.Size()
		for i := 0; i < queueSize; i++ {
			r, err := queue.GetAndRemoveNext()
			if err != nil {
				return err
			}

			if r.TID != tid || r.FatherPCB.PID != pid {
				queue.Add(r)
			} else {
				kernelglobals.ExitStateQueue.Add(r)
				go func() {
					<-kernelsync.PendingThreadsChannel
				}()
			}
		}
	}

	return errors.New("no se pudo eliminar el hilo con TID especificado o no pertenece al PCB con PID especificado")
}

func (cmm *ColasMultiNivel) Planificar() (*kerneltypes.TCB, error) {
	logger.Debug("Planificando en CMN")

	var nextTcb *kerneltypes.TCB
	var err error
	for _, cola := range cmm.ReadyQueue {
		if !cola.IsEmpty() {
			nextTcb, err = cola.GetAndRemoveNext()
			if err != nil {
				return nil, err
			}

			return nextTcb, nil
		}
	}
	return nil, errors.New("no hay ningun hilo en ready")
}

func (cmm *ColasMultiNivel) AddToReady(tcb *kerneltypes.TCB) error {
	// Inicializo la cola si es la primera vez que se llama
	if cmm.ReadyQueue == nil {
		cmm.ReadyQueue = make([]*types.Queue[*kerneltypes.TCB], 0)
	}

	inserted := false
	for i := range cmm.ReadyQueue {
		// Verifico si ya existe una cola de la prioridad del hilo
		if cmm.ReadyQueue[i].Priority == tcb.Prioridad {
			// Si existe lo agrego de forma FIFO a la cola y salgo
			cmm.ReadyQueue[i].Add(tcb)
			inserted = true
			break
		}
	}

	// Si no existe una lista de esa prioridad
	if !inserted {
		err := cmm.addNewQueue(tcb)
		if err != nil {
			return err
		}
	}

	logger.Info("Se agrego a Ready de CMN el TID: %v", tcb.TID)
	go func() {
		logger.Debug("Se manda que hay hilos pendientes por el PendingThreadsChannel")
		kernelsync.PendingThreadsChannel <- true
	}()

	// Desalojo la cpu si es necesario
	if kernelglobals.ExecStateThread != nil {
		if tcb.Prioridad < kernelglobals.ExecStateThread.Prioridad {
			logger.Info("Desalojando el hilo con TID: %v y Prioridad: %v por el hilo con TID: %v y Prioridad mayor: %v",
				kernelglobals.ExecStateThread.TID, kernelglobals.ExecStateThread.Prioridad, tcb.TID, tcb.Prioridad)
			err := shorttermscheduler.CpuInterrupt(
				types.Interruption{
					Type:        types.InterruptionEviction,
					Description: "Interrupcion por desalojo",
				})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (cmm *ColasMultiNivel) addNewQueue(tcb *kerneltypes.TCB) error {
	// Creo la cola y la agrego al slice de colas
	newQueue := new(types.Queue[*kerneltypes.TCB])
	newQueue.Priority = tcb.Prioridad
	newQueue.Add(tcb)
	logger.Info("Se crea nueva cola de prioridad: %v en CMN", tcb.Prioridad)
	// Buscar la posición correcta para insertar la nueva cola
	insertedAt := false
	for i := range cmm.ReadyQueue {
		if newQueue.Priority < cmm.ReadyQueue[i].Priority {
			// Insertar la nueva cola en la posición `i` sin remover otros elementos
			cmm.ReadyQueue = append(cmm.ReadyQueue[:i], append([]*types.Queue[*kerneltypes.TCB]{newQueue}, cmm.ReadyQueue[i:]...)...)
			insertedAt = true
			break
		}
	}
	// Si la prioridad es la menor (número más alto), se agrega al final
	if !insertedAt {
		cmm.ReadyQueue = append(cmm.ReadyQueue, newQueue)
	}

	for i, v := range cmm.ReadyQueue {
		logger.Trace("Cola %v, prioridad: %v", i, v.Priority)
	}

	return nil
}
