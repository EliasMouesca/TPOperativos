package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler/Fifo"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"sync"
	"testing"
)

// TODO: ---------------------------------- TEST PARA THREAD ----------------------------------

func TestThreadCreate(t *testing.T) {
	// Inicializar variables globales
	kernelglobals.EveryPCBInTheKernel = []kerneltypes.PCB{}
	kernelglobals.EveryTCBInTheKernel = []kerneltypes.TCB{}
	kernelglobals.ExecStateThread = nil // Empezamos sin un thread ejecutándose

	kernelglobals.ShortTermScheduler = &Fifo.Fifo{
		Ready: types.Queue[*kerneltypes.TCB]{}, // Inicializa la cola FIFO
	}

	// Crear un PCB y agregarlo a EveryPCBInTheKernel
	newPID := types.Pid(1)
	newPCB := kerneltypes.PCB{
		PID:            newPID,
		TIDs:           []types.Tid{},
		CreatedMutexes: []kerneltypes.Mutex{},
	}
	kernelglobals.EveryPCBInTheKernel = append(kernelglobals.EveryPCBInTheKernel, newPCB)

	// Asignar la referencia correcta del PCB guardado en EveryPCBInTheKernel
	fatherPCB := buscarPCBPorPID(newPID)

	// Crear un TCB para ese PCB
	execTCB := kerneltypes.TCB{
		TID:           5,         // Primer hilo
		Prioridad:     1,         // Prioridad inicial
		FatherPCB:     fatherPCB, // Usar el puntero al PCB que está en EveryPCBInTheKernel
		LockedMutexes: []*kerneltypes.Mutex{},
		JoinedTCB:     nil,
	}

	// Añadir el TCB a EveryTCBInTheKernel y obtener la referencia
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, execTCB)
	kernelglobals.ExecStateThread = buscarTCBPorTID(5, fatherPCB.PID)
	kernelglobals.ExecStateThread.FatherPCB.TIDs = append(kernelglobals.ExecStateThread.FatherPCB.TIDs, kernelglobals.ExecStateThread.TID)

	logCurrentState("Estado inicial")

	// Argumentos de entrada para ThreadCreate
	args := []string{"file.psc", "5"} // Archivo y prioridad del nuevo hilo

	var wg sync.WaitGroup
	wg.Add(2)

	// Llamar a ThreadCreate para crear un nuevo hilo (TCB)
	go func() {
		defer wg.Done()
		ThreadCreate(args)
	}()

	// Consumir el canal para evitar el bloqueo
	go func() {
		defer wg.Done()
		<-kernelsync.ChannelThreadCreate
	}()

	wg.Wait()

	// Verificar que se creó un nuevo TCB
	if len(kernelglobals.EveryTCBInTheKernel) != 2 {
		t.Errorf("Debería haber 2 TCBs en EveryTCBInTheKernel, pero hay %d", len(kernelglobals.EveryTCBInTheKernel))
	}

	// Verificar que el nuevo TCB fue añadido a EveryTCBInTheKernel
	newTCB := kernelglobals.EveryTCBInTheKernel[1] // El segundo TCB es el nuevo
	if newTCB.TID != 1 {
		t.Errorf("El nuevo TCB debería tener TID 1, pero tiene TID %d", newTCB.TID)
	}
	if newTCB.FatherPCB.PID != newPID {
		t.Errorf("El nuevo TCB debería pertenecer al PCB con PID %d, pero tiene PID %d", newPID, newTCB.FatherPCB.PID)
	}

	// Verificar que el nuevo TCB está en la cola de new usando ThreadExists
	foundInNewStateQueue := false
	kernelglobals.NewStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == newTCB.TID && tcb.FatherPCB.PID == newTCB.FatherPCB.PID {
			foundInNewStateQueue = true
		}
	})

	if !foundInNewStateQueue {
		t.Errorf("El nuevo TCB con TID %d y PID %d no fue añadido a la cola NewStateQueue", newTCB.TID, newTCB.FatherPCB.PID)
	}

	logCurrentState("Estado Final")
}

func TestThreadJoin(t *testing.T) {
	// Inicializar variables globales
	kernelglobals.EveryPCBInTheKernel = []kerneltypes.PCB{}
	kernelglobals.EveryTCBInTheKernel = []kerneltypes.TCB{}
	kernelglobals.ExecStateThread = nil
	kernelglobals.ExitStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.BlockedStateQueue = types.Queue[*kerneltypes.TCB]{}

	// Crear un PCB y agregarlo a EveryPCBInTheKernel
	newPID := types.Pid(1)
	newPCB := kerneltypes.PCB{
		PID:            newPID,
		TIDs:           []types.Tid{},
		CreatedMutexes: []kerneltypes.Mutex{},
	}
	kernelglobals.EveryPCBInTheKernel = append(kernelglobals.EveryPCBInTheKernel, newPCB)

	// Asignar la referencia correcta del PCB guardado en EveryPCBInTheKernel
	fatherPCB := buscarPCBPorPID(newPID)

	// Crear dos TCBs, uno para el hilo actual y otro para el TID a joinear
	execTCB := kerneltypes.TCB{
		TID:           0,         // Hilo actual
		Prioridad:     1,         // Prioridad inicial
		FatherPCB:     fatherPCB, // Asignar el PCB
		LockedMutexes: []*kerneltypes.Mutex{},
		JoinedTCB:     nil,
	}
	joinedTCB := kerneltypes.TCB{
		TID:           2, // Hilo a joinear
		Prioridad:     1,
		FatherPCB:     fatherPCB, // Mismo PCB
		LockedMutexes: []*kerneltypes.Mutex{},
		JoinedTCB:     nil,
	}

	// Añadir el TCB del hilo actual a EveryTCBInTheKernel
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, execTCB)

	// Inicializar el hilo actual en ejecución
	kernelglobals.ExecStateThread = buscarTCBPorTID(0, fatherPCB.PID)

	// Añadir el TCB del hilo a joinear a EveryTCBInTheKernel
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, joinedTCB)
	kernelglobals.ShortTermScheduler.AddToReady(buscarTCBPorTID(2, fatherPCB.PID))

	fatherPCB.TIDs = append(fatherPCB.TIDs, execTCB.TID)
	fatherPCB.TIDs = append(fatherPCB.TIDs, joinedTCB.TID)

	logCurrentState("Estado Inicial.")

	// Argumentos de entrada para ThreadJoin (TID del hilo a joinear)
	args := []string{"2"} // El TID del hilo a joinear

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ThreadJoin(args)
	}()

	wg.Wait()

	logger.Info("everytcbinthekernel:  %v", kernelglobals.EveryTCBInTheKernel)

	logCurrentState("Estado Final")

	// Verificar que el hilo actual está bloqueado en la cola de BlockedStateQueue
	if kernelglobals.ExecStateThread != nil {
		t.Errorf("ExecStateThread debería ser nil, pero no lo es")
	}

	tcb := buscarTCBPorTID(0, fatherPCB.PID)
	if tcb.JoinedTCB != buscarTCBPorTID(2, fatherPCB.PID) {
		t.Errorf("El joinedTCB del TCB que ejecuto la syscall quedo en nil, deberia estar apuntando al que llamo.")
	}

	// Verificar que el hilo actual está en la cola de BlockedStateQueue
	logger.Info("Recorriendo BlockedStateQueue para ver si se agrego correctamente el TCB: %v. ", execTCB.TID)
	blocked := false
	queueSize := kernelglobals.BlockedStateQueue.Size()
	for i := 0; i < queueSize; i++ {
		tcb, _ := kernelglobals.BlockedStateQueue.GetAndRemoveNext()
		if tcb.TID == execTCB.TID {
			blocked = true
			logger.Info("Se encontro el TCB con TID %v en la cola BlockedStateQueue. ", execTCB.TID)
		}
		kernelglobals.BlockedStateQueue.Add(tcb) // Volver a agregar a la cola
	}
	if !blocked {
		t.Errorf("El hilo actual no fue añadido a la BlockedStateQueue correctamente")
	}
}

func TestThreadCancel(t *testing.T) {
	// Inicializar variables globales
	kernelglobals.EveryPCBInTheKernel = []kerneltypes.PCB{}
	kernelglobals.EveryTCBInTheKernel = []kerneltypes.TCB{}
	kernelglobals.ExecStateThread = nil
	kernelglobals.ExitStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.BlockedStateQueue = types.Queue[*kerneltypes.TCB]{}

	// Inicializar el planificador con FIFO para facilitar la prueba
	kernelglobals.ShortTermScheduler = &Fifo.Fifo{
		Ready: types.Queue[*kerneltypes.TCB]{}, // Inicializa la cola FIFO
	}

	// Crear un PCB y agregarlo a EveryPCBInTheKernel
	newPID := types.Pid(1)
	newPCB := kerneltypes.PCB{
		PID:            newPID,
		TIDs:           []types.Tid{},
		CreatedMutexes: []kerneltypes.Mutex{},
	}
	kernelglobals.EveryPCBInTheKernel = append(kernelglobals.EveryPCBInTheKernel, newPCB)

	// Asignar la referencia correcta del PCB guardado en EveryPCBInTheKernel
	fatherPCB := buscarPCBPorPID(newPID)

	// Crear dos TCBs, uno para el hilo actual y otro para el TID a cancelar
	execTCB := kerneltypes.TCB{
		TID:           0,         // Hilo actual
		Prioridad:     1,         // Prioridad inicial
		FatherPCB:     fatherPCB, // Asignar el PCB
		LockedMutexes: []*kerneltypes.Mutex{},
		JoinedTCB:     nil,
	}
	cancelTCB := kerneltypes.TCB{
		TID:           2, // Hilo a cancelar
		Prioridad:     1,
		FatherPCB:     fatherPCB, // Mismo PCB
		LockedMutexes: []*kerneltypes.Mutex{},
		JoinedTCB:     nil,
	}

	// Añadir el TCB del hilo actual a EveryTCBInTheKernel
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, execTCB)

	// Inicializar el hilo actual en ejecución
	kernelglobals.ExecStateThread = buscarTCBPorTID(0, fatherPCB.PID)

	// Añadir el TCB del hilo a cancelar a la cola de Ready (simulando que está listo para ejecutarse)
	kernelglobals.ShortTermScheduler.AddToReady(&cancelTCB)

	// Añadir el TID del hilo actual y el hilo a cancelar al PCB
	fatherPCB.TIDs = append(fatherPCB.TIDs, execTCB.TID)
	fatherPCB.TIDs = append(fatherPCB.TIDs, cancelTCB.TID)

	logCurrentState("Estado Inicial.")

	// Argumentos de entrada para ThreadCancel (TID del hilo a cancelar)
	args := []string{"2"} // El TID del hilo a cancelar

	// Llamar a ThreadCancel
	err := ThreadCancel(args)
	if err != nil {
		t.Errorf("Error inesperado en ThreadCancel: %v", err)
	}

	logCurrentState("Estado luego de llamar a ThreadCancel")

	// Verificar que el hilo cancelado fue movido a la cola de ExitStateQueue
	logger.Info("Recorriendo ExitStateQueue para ver si se agregó correctamente el TCB: %v.", cancelTCB.TID)
	cancelled := false
	queueSize := kernelglobals.ExitStateQueue.Size()
	for i := 0; i < queueSize; i++ {
		tcb, _ := kernelglobals.ExitStateQueue.GetAndRemoveNext()
		if tcb.TID == cancelTCB.TID {
			cancelled = true
			logger.Info("Se encontró el TCB con TID %v en la cola ExitStateQueue.", cancelTCB.TID)
		}
		kernelglobals.ExitStateQueue.Add(tcb) // Volver a agregar a la cola
	}
	if !cancelled {
		t.Errorf("El hilo cancelado no fue añadido a la ExitStateQueue correctamente")
	}

	// Verificar que el hilo ya no está en la cola de ready
	exists, _ := kernelglobals.ShortTermScheduler.ThreadExists(cancelTCB.TID, fatherPCB.PID)
	if exists {
		t.Errorf("El hilo cancelado no fue removido de la cola de Ready")
	}
}

func TestThreadExit(t *testing.T) {
	// Inicializar variables globales
	kernelglobals.EveryPCBInTheKernel = []kerneltypes.PCB{}
	kernelglobals.EveryTCBInTheKernel = []kerneltypes.TCB{}
	kernelglobals.ExecStateThread = nil
	kernelglobals.ExitStateQueue = types.Queue[*kerneltypes.TCB]{}

	// Crear un PCB y agregarlo a EveryPCBInTheKernel
	newPID := types.Pid(1)
	newPCB := kerneltypes.PCB{
		PID:            newPID,
		TIDs:           []types.Tid{},
		CreatedMutexes: []kerneltypes.Mutex{},
	}
	kernelglobals.EveryPCBInTheKernel = append(kernelglobals.EveryPCBInTheKernel, newPCB)

	// Asignar la referencia correcta del PCB guardado en EveryPCBInTheKernel
	fatherPCB := buscarPCBPorPID(newPID)

	// Crear un TCB para el hilo actual
	execTCB := kerneltypes.TCB{
		TID:           0,         // Hilo actual
		Prioridad:     1,         // Prioridad inicial
		FatherPCB:     fatherPCB, // Asignar el PCB
		LockedMutexes: []*kerneltypes.Mutex{},
		JoinedTCB:     nil,
	}

	// Añadir el TCB del hilo actual a EveryTCBInTheKernel
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, execTCB)

	// Inicializar el hilo actual en ejecución
	kernelglobals.ExecStateThread = buscarTCBPorTID(0, fatherPCB.PID)

	// Añadir el TID del hilo actual al PCB
	fatherPCB.TIDs = append(fatherPCB.TIDs, execTCB.TID)

	logCurrentState("Estado Inicial.")

	//Envio al canal com que le dio signal, ya que como estamos testeando solo la syscall y no el planificador, se quedaria en deadlock...
	go func() {
		kernelsync.ThreadExitComplete <- struct{}{}
	}()

	// Llamar a ThreadExit (el hilo actual se auto-finaliza)
	err := ThreadExit([]string{})
	if err != nil {
		t.Errorf("Error inesperado en ThreadExit: %v", err)
	}

	logCurrentState("Estado luego de llamar a ThreadExit")

	// Verificar que el hilo actual fue movido a la cola de ExitStateQueue
	logger.Info("Recorriendo ExitStateQueue para ver si se agregó correctamente el TCB: %v.", execTCB.TID)
	exited := false
	queueSize := kernelglobals.ExitStateQueue.Size()
	for i := 0; i < queueSize; i++ {
		tcb, _ := kernelglobals.ExitStateQueue.GetAndRemoveNext()
		if tcb.TID == execTCB.TID {
			exited = true
			logger.Info("Se encontró el TCB con TID %v en la cola ExitStateQueue.", execTCB.TID)
		}
		kernelglobals.ExitStateQueue.Add(tcb) // Volver a agregar a la cola
	}
	if !exited {
		t.Errorf("El hilo actual no fue añadido a la ExitStateQueue correctamente")
	}

	// Verificar que ExecStateThread ahora es nil
	if kernelglobals.ExecStateThread != nil {
		t.Errorf("ExecStateThread debería ser nil, pero no lo es")
	}
}
