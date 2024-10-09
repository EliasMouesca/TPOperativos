package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/kernel/shorttermscheduler/Fifo"
	"github.com/sisoputnfrba/tp-golang/types"
	"testing"
)

// TODO: ---------------------------------- TEST PARA MUTEX ----------------------------------

func TestMutexCreate(t *testing.T) {
	// Inicializar variables globales
	kernelglobals.EveryPCBInTheKernel = []kerneltypes.PCB{}
	kernelglobals.EveryTCBInTheKernel = []kerneltypes.TCB{}
	kernelglobals.ExecStateThread = nil

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
	kernelglobals.ExecStateThread = &kernelglobals.EveryTCBInTheKernel[len(kernelglobals.EveryTCBInTheKernel)-1]
	fatherPCB.TIDs = append(fatherPCB.TIDs, execTCB.TID)

	// Argumentos de entrada para MutexCreate (nombre del mutex)
	args := []string{"mutex_1"}

	// Llamar a MutexCreate
	err := MutexCreate(args)
	if err != nil {
		t.Errorf("Error inesperado en MutexCreate: %v", err)
	}

	logCurrentState("Estado luego de llamar a MutexCreate")

	// Verificar que se creó un mutex y se añadió a la lista CreatedMutexes del PCB
	if len(fatherPCB.CreatedMutexes) != 1 {
		t.Errorf("Debería haber 1 mutex en CreatedMutexes, pero hay %d", len(fatherPCB.CreatedMutexes))
	}

	// Verificar que el nombre del mutex es el correcto
	createdMutex := fatherPCB.CreatedMutexes[0]
	if createdMutex.Name != "mutex_1" {
		t.Errorf("El mutex debería tener el nombre 'mutex_1', pero tiene '%s'", createdMutex.Name)
	}

	// Verificar que el mutex no está asignado a ningún hilo
	if createdMutex.AssignedTCB != nil {
		t.Errorf("El mutex no debería estar asignado a ningún hilo, pero AssignedTCB no es nil")
	}

	// Verificar que la lista de BlockedTCBs está vacía
	if len(createdMutex.BlockedTCBs) != 0 {
		t.Errorf("La lista de BlockedTCBs del mutex debería estar vacía, pero tiene %d elementos", len(createdMutex.BlockedTCBs))
	}
}

func TestMutexLock(t *testing.T) {
	// Inicializar variables globales
	kernelglobals.EveryPCBInTheKernel = []kerneltypes.PCB{}
	kernelglobals.EveryTCBInTheKernel = []kerneltypes.TCB{}
	kernelglobals.ExecStateThread = nil

	// Crear un PCB y agregarlo a EveryPCBInTheKernel
	newPID := types.Pid(1)
	newPCB := kerneltypes.PCB{
		PID:            newPID,
		TIDs:           []types.Tid{},
		CreatedMutexes: []kerneltypes.Mutex{},
	}
	kernelglobals.EveryPCBInTheKernel = append(kernelglobals.EveryPCBInTheKernel, newPCB)

	// Asignar la referencia correcta del PCB guardado en EveryPCBInTheKernel
	fatherPCB := &kernelglobals.EveryPCBInTheKernel[len(kernelglobals.EveryPCBInTheKernel)-1]

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
	kernelglobals.ExecStateThread = &kernelglobals.EveryTCBInTheKernel[len(kernelglobals.EveryTCBInTheKernel)-1]
	fatherPCB.TIDs = append(fatherPCB.TIDs, execTCB.TID)

	// Crear un mutex y añadirlo al PCB
	mutex := kerneltypes.Mutex{
		Name:        "mutex_1",
		AssignedTCB: nil,
		BlockedTCBs: []*kerneltypes.TCB{},
	}
	fatherPCB.CreatedMutexes = append(fatherPCB.CreatedMutexes, mutex)

	logCurrentState("Estado Inicial")

	// Argumentos de entrada para MutexLock (nombre del mutex)
	args := []string{"mutex_1"}

	// Llamar a MutexLock con el mutex disponible
	err := MutexLock(args)
	if err != nil {
		t.Errorf("Error inesperado en MutexLock: %v", err)
	}

	logCurrentState("Estado luego de llamar a MutexLock")

	// Verificar que el mutex fue asignado al hilo actual
	if fatherPCB.CreatedMutexes[0].AssignedTCB.TID != kernelglobals.ExecStateThread.TID {
		t.Errorf("El mutex no fue asignado al hilo actual correctamente")
	}

	// Verificar que el hilo actual tiene el mutex bloqueado
	if len(kernelglobals.ExecStateThread.LockedMutexes) != 1 || kernelglobals.ExecStateThread.LockedMutexes[0].Name != "mutex_1" {
		t.Errorf("El hilo actual no tiene el mutex bloqueado correctamente")
	}

	// Crear un segundo hilo (TCB) que intente tomar el mutex
	blockedTCB := kerneltypes.TCB{
		TID:           1,         // Segundo hilo
		Prioridad:     1,         // Prioridad inicial
		FatherPCB:     fatherPCB, // Mismo PCB
		LockedMutexes: []*kerneltypes.Mutex{},
		JoinedTCB:     nil,
	}
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, blockedTCB)
	fatherPCB.TIDs = append(fatherPCB.TIDs, blockedTCB.TID)

	// Inicializar el nuevo hilo en ejecución
	kernelglobals.ExecStateThread = &kernelglobals.EveryTCBInTheKernel[len(kernelglobals.EveryTCBInTheKernel)-1]

	// Llamar a MutexLock con el mutex ya tomado
	err = MutexLock(args)
	if err != nil {
		t.Errorf("Error inesperado en MutexLock: %v", err)
	}

	// Verificar que el segundo hilo fue bloqueado
	if len(fatherPCB.CreatedMutexes[0].BlockedTCBs) != 1 || fatherPCB.CreatedMutexes[0].BlockedTCBs[0].TID != blockedTCB.TID {
		t.Errorf("El segundo hilo no fue bloqueado correctamente")
	}

	// Verificar que el mutex sigue asignado al primer hilo
	if fatherPCB.CreatedMutexes[0].AssignedTCB.TID != execTCB.TID {
		t.Errorf("El mutex debería seguir asignado al primer hilo, pero no lo está")
	}

	// Verificar el caso en el que el mutex no existe
	argsInvalid := []string{"mutex_inexistente"}
	err = MutexLock(argsInvalid)
	if err == nil || err.Error() != "No se encontró el mutex <mutex_inexistente>" {
		t.Errorf("Debería haberse producido un error al no encontrar el mutex, pero no ocurrió")
	}

	logCurrentState("Estado Final")
}
func TestMutexUnlock(t *testing.T) {
	// Inicializar variables globales
	kernelglobals.EveryPCBInTheKernel = []kerneltypes.PCB{}
	kernelglobals.EveryTCBInTheKernel = []kerneltypes.TCB{}
	kernelglobals.ExecStateThread = nil
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
	fatherPCB := &kernelglobals.EveryPCBInTheKernel[len(kernelglobals.EveryPCBInTheKernel)-1]

	// Crear un TCB para el hilo actual que bloqueará el mutex
	execTCB := kerneltypes.TCB{
		TID:           0,                      // Hilo actual
		Prioridad:     1,                      // Prioridad inicial
		FatherPCB:     fatherPCB,              // Asignar el PCB
		LockedMutexes: []*kerneltypes.Mutex{}, // Inicializar lista de mutexes
		JoinedTCB:     nil,
	}
	// Añadir el TCB del hilo actual a EveryTCBInTheKernel
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, execTCB)
	// Inicializar el hilo actual en ejecución
	kernelglobals.ExecStateThread = &kernelglobals.EveryTCBInTheKernel[len(kernelglobals.EveryTCBInTheKernel)-1]
	fatherPCB.TIDs = append(fatherPCB.TIDs, execTCB.TID)

	// Crear un mutex y bloquearlo por el hilo actual
	mutex := kerneltypes.Mutex{
		Name:        "mutex_1",
		AssignedTCB: kernelglobals.ExecStateThread,
		BlockedTCBs: []*kerneltypes.TCB{},
	}
	// Añadir el mutex a la lista de CreatedMutexes del PCB
	fatherPCB.CreatedMutexes = append(fatherPCB.CreatedMutexes, mutex)

	// Obtener el puntero al mutex real en CreatedMutexes
	mutexPtr := &fatherPCB.CreatedMutexes[0]

	// Añadir el mutex al LockedMutexes del hilo
	kernelglobals.ExecStateThread.LockedMutexes = append(kernelglobals.ExecStateThread.LockedMutexes, mutexPtr)

	// Crear un segundo TCB que estará bloqueado esperando el mutex
	blockedTCB := kerneltypes.TCB{
		TID:           1,                      // Segundo hilo
		Prioridad:     1,                      // Prioridad inicial
		FatherPCB:     fatherPCB,              // Mismo PCB
		LockedMutexes: []*kerneltypes.Mutex{}, // Inicializar lista de mutexes
		JoinedTCB:     nil,
	}
	// Añadir el segundo TCB a la lista de BlockedTCBs del mutex
	mutexPtr.BlockedTCBs = append(mutexPtr.BlockedTCBs, &blockedTCB)

	// Añadir el segundo TCB a EveryTCBInTheKernel y actualizar la lista de TIDs del PCB
	kernelglobals.EveryTCBInTheKernel = append(kernelglobals.EveryTCBInTheKernel, blockedTCB)
	fatherPCB.TIDs = append(fatherPCB.TIDs, blockedTCB.TID)

	logCurrentState("Estado Inicial")

	// Argumentos de entrada para MutexUnlock (nombre del mutex)
	args := []string{"mutex_1"}

	// Llamar a MutexUnlock con el mutex bloqueado por el hilo actual
	err := MutexUnlock(args)
	if err != nil {
		t.Errorf("Error inesperado en MutexUnlock: %v", err)
	}

	// Verificar que el mutex ha sido asignado al segundo hilo (blockedTCB)
	if mutexPtr.AssignedTCB == nil || mutexPtr.AssignedTCB.TID != blockedTCB.TID {
		t.Errorf("El mutex no fue reasignado correctamente al segundo hilo bloqueado")
	}

	// Verificar que el segundo hilo tiene el mutex en su lista de LockedMutexes
	if len(blockedTCB.LockedMutexes) != 1 || blockedTCB.LockedMutexes[0].Name != "mutex_1" {
		t.Errorf("El segundo hilo no tiene el mutex bloqueado correctamente")
	}

	// Verificar que el hilo actual ya no tiene el mutex en su lista de LockedMutexes
	if len(kernelglobals.ExecStateThread.LockedMutexes) != 0 {
		t.Errorf("El hilo actual no liberó correctamente el mutex")
	}

	logCurrentState("Estado luego de llamar a MutexUnlock")

	// Verificar que el segundo hilo fue movido a la cola de Ready
	exists, err := kernelglobals.ShortTermScheduler.ThreadExists(blockedTCB.TID, fatherPCB.PID)
	if err != nil {
		t.Errorf("Error al verificar la existencia del TCB en la cola de Ready: %v", err)
	}
	if !exists {
		t.Errorf("El segundo hilo no fue añadido a la cola de Ready correctamente")
	}
}
