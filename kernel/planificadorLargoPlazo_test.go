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
	"time"
)

func TestNewProcessToReady(t *testing.T) {
	// Inicializar variables globales
	kernelglobals.EveryPCBInTheKernel = []kerneltypes.PCB{}
	kernelglobals.EveryTCBInTheKernel = []kerneltypes.TCB{}
	kernelglobals.NewPCBStateQueue = types.Queue[*kerneltypes.PCB]{}
	kernelglobals.ShortTermScheduler = &Fifo.Fifo{
		Ready: types.Queue[*kerneltypes.TCB]{}, // Inicializa la cola Ready
	}

	// Crear el PCB y agregarlo a la cola NewPCBStateQueue
	processCreate := kerneltypes.PCB{
		PID:  types.Pid(5),
		TIDs: []types.Tid{0},
	}
	kernelglobals.EveryPCBInTheKernel = append(kernelglobals.EveryPCBInTheKernel, processCreate)
	fatherPCB := buscarPCBPorPID(5)

	// Añadir el PCB a NewPCBStateQueue
	kernelglobals.NewPCBStateQueue.Add(fatherPCB)

	// Inicializar los canales
	kernelsync.ChannelProcessArguments = make(chan []string, 1)

	// Crear los argumentos que se enviarán a través del canal
	args := []string{"test_file", "500", "0"}

	// Ejecutar la syscall ProcessCreate en un goroutine

	logCurrentState("Estado Inicial")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := ProcessCreate(args)
		if err != nil {
			t.Errorf("Error inesperado en ProcessCreate: %v", err)
		}
	}()

	go func() {
		NewProcessToReady()
	}()

	wg.Wait()

	// Verificar que el hilo principal fue movido a la cola Ready
	existsInReady, _ := kernelglobals.ShortTermScheduler.ThreadExists(0, 2)
	if !existsInReady {
		t.Errorf("El (<%d:0>) no fue movido correctamente a ReadyStateQueue", processCreate.PID)
	} else {
		logger.Info("## (<%v:%v>) fue movido a la cola de READY.", processCreate.PID, processCreate.TIDs[0])
	}

	// Verificar que el TCB fue agregado a EveryTCBInTheKernel
	if len(kernelglobals.EveryTCBInTheKernel) != 1 {
		t.Errorf("No se agregó correctamente el TCB a EveryTCBInTheKernel")
	}

	logCurrentState("Estado Final")
}

func TestProcessToExit(t *testing.T) {
	// Inicializar variables globales
	kernelglobals.EveryPCBInTheKernel = []kerneltypes.PCB{}
	kernelglobals.EveryTCBInTheKernel = []kerneltypes.TCB{}
	kernelglobals.NewStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.BlockedStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.ExitStateQueue = types.Queue[*kerneltypes.TCB]{}

	// Configurar el scheduler FIFO para Ready
	kernelglobals.ShortTermScheduler = &Fifo.Fifo{
		Ready: types.Queue[*kerneltypes.TCB]{}, // Inicializa la cola Ready
	}

	// Crear un PCB y agregarlo a EveryPCBInTheKernel
	newPID := types.Pid(1)
	pcb := kerneltypes.PCB{
		PID:  newPID,
		TIDs: []types.Tid{0, 1, 2, 3},
	}
	kernelglobals.EveryPCBInTheKernel = append(kernelglobals.EveryPCBInTheKernel, pcb)

	// Obtener la referencia correcta del PCB desde EveryPCBInTheKernel
	fatherPCB := buscarPCBPorPID(newPID)

	// Crear 4 hilos asociados a este PCB
	mainThread := kerneltypes.TCB{TID: 0, Prioridad: 1, FatherPCB: fatherPCB}
	readyThread := kerneltypes.TCB{TID: 1, Prioridad: 1, FatherPCB: fatherPCB}
	blockedThread := kerneltypes.TCB{TID: 2, Prioridad: 1, FatherPCB: fatherPCB}
	newThread := kerneltypes.TCB{TID: 3, Prioridad: 1, FatherPCB: fatherPCB}

	// Agregar los hilos a EveryTCBInTheKernel
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, mainThread, readyThread, blockedThread, newThread)

	// Asignar el hilo principal como el hilo ejecutándose
	kernelglobals.ExecStateThread = &mainThread

	// Agregar hilos a las colas correspondientes
	kernelglobals.ShortTermScheduler.AddToReady(&readyThread)
	kernelglobals.BlockedStateQueue.Add(&blockedThread)
	kernelglobals.NewStateQueue.Add(&newThread)

	// Simular la señalización para finalizar el proceso
	kernelsync.ChannelFinishprocess = make(chan types.Pid, 1)

	logCurrentState("Estado Inicial")

	// Ejecutar ProcessExit en un goroutine para que no bloquee
	go func() {
		err := ProcessExit([]string{})
		if err != nil {
			t.Errorf("Error inesperado en ProcessExit: %v", err)
		}
	}()

	// Ejecutar ProcessToExit en otro goroutine para manejar la comunicación con memoria
	go ProcessToExit()

	// Enviar la señal de finalización de proceso
	kernelsync.ChannelFinishprocess <- pcb.PID

	// Esperar para asegurar que los hilos hayan sido procesados
	time.Sleep(100 * time.Millisecond)

	// Verificar que todos los hilos fueron movidos a la cola ExitStateQueue
	for _, tcb := range []*kerneltypes.TCB{&mainThread, &readyThread, &blockedThread, &newThread} {
		found := false
		kernelglobals.ExitStateQueue.Do(func(exitTCB *kerneltypes.TCB) {
			if exitTCB.TID == tcb.TID && exitTCB.FatherPCB.PID == pcb.PID {
				found = true
			}
		})
		if !found {
			t.Errorf("El TID <%d> del PCB con PID <%d> no fue movido correctamente a ExitStateQueue", tcb.TID, pcb.PID)
		}
	}

	logCurrentState("Estado final después de ProcessToExit")
}

func TestNewThreadToReady(t *testing.T) {
	// Inicializar variables globales
	kernelglobals.EveryPCBInTheKernel = []kerneltypes.PCB{}
	kernelglobals.EveryTCBInTheKernel = []kerneltypes.TCB{}
	kernelglobals.NewStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.BlockedStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.ExitStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.ShortTermScheduler = &Fifo.Fifo{
		Ready: types.Queue[*kerneltypes.TCB]{}, // Inicializa la cola Ready
	}

	// Crear un PCB y agregarlo a EveryPCBInTheKernel
	newPID := types.Pid(1)
	pcb := kerneltypes.PCB{
		PID:  newPID,
		TIDs: []types.Tid{0}, // Iniciar con TID 0
	}
	kernelglobals.EveryPCBInTheKernel = append(kernelglobals.EveryPCBInTheKernel, pcb)
	fatherPCB := buscarPCBPorPID(newPID)

	// Crear el TCB principal y asignarlo como hilo en ejecución
	mainThread := kerneltypes.TCB{TID: 0, Prioridad: 1, FatherPCB: fatherPCB}
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, mainThread)
	mainThreadPtr := &kernelglobals.EveryTCBInTheKernel[0]
	kernelglobals.ExecStateThread = mainThreadPtr

	// Inicializar el canal de creación de hilos y sincronización
	kernelsync.ChannelThreadCreate = make(chan []string, 1)

	logCurrentState("Estado Inicial")

	// Crear un nuevo hilo utilizando la syscall ThreadCreate
	args := []string{"test_file_thread", "1"} // Nombre de archivo y prioridad
	err := ThreadCreate(args)
	if err != nil {
		t.Errorf("Error inesperado en ThreadCreate: %v", err)
	}

	logCurrentState("Estado antes de Planificar")

	// Ejecutar la función del planificador en un goroutine (simulando comportamiento concurrente)
	go NewThreadToReady()

	// Esperar para asegurar que el hilo haya sido procesado
	time.Sleep(100 * time.Millisecond)

	// Verificar que el TCB creado fue movido a la cola Ready
	existsInReady, _ := kernelglobals.ShortTermScheduler.ThreadExists(1, pcb.PID)
	if !existsInReady {
		t.Errorf("El TID <1> del PCB con PID <%d> no fue movido correctamente a ReadyStateQueue", pcb.PID)
	} else {
		logger.Info("## (<%v:%v>) fue movido a la cola de READY.", pcb.PID, 1)
	}

	// Verificar que el TCB fue agregado a EveryTCBInTheKernel
	if len(kernelglobals.EveryTCBInTheKernel) != 2 {
		t.Errorf("No se agregó correctamente el TCB a EveryTCBInTheKernel")
	}

	// Verificar que el TCB ya no esté en la cola NewStateQueue
	if !kernelglobals.NewStateQueue.IsEmpty() {
		t.Errorf("El TID <1> del PCB con PID <%d> aún se encuentra en la cola NewStateQueue", pcb.PID)
	}

	logCurrentState("Estado Final luego de mover el hilo a Ready")
}

func TestThreadExitAndToExit(t *testing.T) {
	// Inicializar variables globales
	kernelglobals.EveryPCBInTheKernel = []kerneltypes.PCB{}
	kernelglobals.EveryTCBInTheKernel = []kerneltypes.TCB{}
	kernelglobals.NewStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.BlockedStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.ExitStateQueue = types.Queue[*kerneltypes.TCB]{}
	kernelglobals.ShortTermScheduler = &Fifo.Fifo{
		Ready: types.Queue[*kerneltypes.TCB]{}, // Inicializa la cola de Ready
	}

	// Inicializar canal
	kernelsync.ChannelFinishThread = make(chan []string)
	kernelsync.ThreadExitComplete = make(chan struct{})

	// Crear un PCB
	newPID := types.Pid(1)
	pcb := kerneltypes.PCB{
		PID:  newPID,
		TIDs: []types.Tid{0, 1, 2}, // El TCB principal y dos threads adicionales
	}
	kernelglobals.EveryPCBInTheKernel = append(kernelglobals.EveryPCBInTheKernel, pcb)
	fatherPCB := buscarPCBPorPID(newPID)

	// Crear TCB principal (execTCB)
	execTCB := kerneltypes.TCB{TID: 0, Prioridad: 1, FatherPCB: fatherPCB}
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, execTCB)
	execTCBPtr := buscarTCBPorTID(0, fatherPCB.PID)
	kernelglobals.ExecStateThread = execTCBPtr

	// Crear mutex y asignarlo al execTCB
	mutex := kerneltypes.Mutex{Name: "testMutex", AssignedTCB: execTCBPtr}
	fatherPCB.CreatedMutexes = append(fatherPCB.CreatedMutexes, mutex)
	execTCBPtr.LockedMutexes = append(execTCBPtr.LockedMutexes, &fatherPCB.CreatedMutexes[0])

	// Crear TCB bloqueado por THREAD_JOIN
	joinTCB := kerneltypes.TCB{
		TID:       1,
		Prioridad: 1,
		FatherPCB: fatherPCB,
		JoinedTCB: execTCBPtr, // Bloqueado esperando que execTCB termine
	}
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, joinTCB)
	joinTCBPtr := buscarTCBPorTID(1, fatherPCB.PID)
	kernelglobals.BlockedStateQueue.Add(joinTCBPtr)

	// Crear TCB bloqueado por mutex
	mutexBlockedTCB := kerneltypes.TCB{
		TID:       2,
		Prioridad: 1,
		FatherPCB: fatherPCB,
	}
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, mutexBlockedTCB)
	mutexBlockedTCBPtr := buscarTCBPorTID(2, fatherPCB.PID)
	mutex.BlockedTCBs = append(mutex.BlockedTCBs, mutexBlockedTCBPtr)

	// Asignar mutex al PCB
	for i := range kernelglobals.EveryPCBInTheKernel {
		if kernelglobals.EveryPCBInTheKernel[i].PID == fatherPCB.PID {
			kernelglobals.EveryPCBInTheKernel[i].CreatedMutexes[0] = mutex
		}
	}

	kernelglobals.BlockedStateQueue.Add(mutexBlockedTCBPtr)

	logCurrentState("Estado Inicial")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ThreadExit([]string{})
	}()

	// Ejecutar ThreadToExit en paralelo
	go func() {
		ThreadToExit()
	}()

	wg.Wait()

	foundExecTCBInExit := false
	foundJoinTCBInReady := false
	foundMutexBlockedTCBInReady := false

	// Verificar el TCB finalizado en ExitStateQueue
	kernelglobals.ExitStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == execTCBPtr.TID {
			foundExecTCBInExit = true
		}
	})

	// Verificar que el hilo bloqueado por THREAD_JOIN fue movido a Ready
	kernelglobals.ShortTermScheduler.(*Fifo.Fifo).Ready.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == joinTCBPtr.TID {
			foundJoinTCBInReady = true
		}
	})

	// Verificar que el hilo bloqueado por mutex fue movido a Ready
	kernelglobals.ShortTermScheduler.(*Fifo.Fifo).Ready.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == mutexBlockedTCBPtr.TID {
			foundMutexBlockedTCBInReady = true
		}
	})

	// Verificaciones finales
	if !foundExecTCBInExit {
		t.Errorf("El (<%d:%d>) no fue movido correctamente a ExitStateQueue", fatherPCB.PID, execTCBPtr.TID)
	}
	if !foundJoinTCBInReady {
		t.Errorf("El TID <%d> del PCB <%d> bloqueado por THREAD_JOIN no fue movido correctamente a ReadyStateQueue", joinTCBPtr.TID, fatherPCB.PID)
	}
	if !foundMutexBlockedTCBInReady {
		t.Errorf("El TID <%d> del PCB <%d> bloqueado por mutex no fue movido correctamente a ReadyStateQueue", mutexBlockedTCBPtr.TID, fatherPCB.PID)
	}

	logCurrentState("Estado Final")
}

func setup2() {
	logger.ConfigureLogger("test.log", "INFO")
}

func TestIO(t *testing.T) {
	setup2()

	pcbs := make([]kerneltypes.PCB, 0)
	tcbChannels := make(chan *kerneltypes.TCB, 10) // Canal para almacenar TCBs

	// Inicializa varios PCBs y TCBs
	for i := 0; i < 10; i++ {
		pcb := kerneltypes.PCB{PID: types.Pid(i)}
		pcbs = append(pcbs, pcb)
		tcb := kerneltypes.TCB{TID: types.Tid(i), FatherPCB: &pcb}
		tcbChannels <- &tcb
	}

	close(tcbChannels)

	go UnlockIO()

	for tcb := range tcbChannels {
		kernelsync.WaitPlanificadorLP.Add(1) // Aumentar el contador del WaitGroup
		go func(tcb *kerneltypes.TCB) {
			defer kernelsync.WaitPlanificadorLP.Done() // Decrementar al final
			kernelglobals.ExecStateThread = tcb
			args := []string{"100"} // Tiempo de bloqueo
			err := IO(args)
			if err != nil {
				t.Errorf("Error en TCB %v: %v", tcb.TID, err)
			}
		}(tcb)
	}

	kernelsync.WaitPlanificadorLP.Wait()
}
