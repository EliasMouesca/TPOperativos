package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
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

/*
func TestMutexUnlock(t *testing.T) {
	setup()

	// Crear un PCB y TCB de prueba
	pcb := kerneltypes.PCB{
		PID:            0,
		TIDs:           []int{0, 1, 2}, // Hilos con TID 0, 1 y 2
		CreatedMutexes: []int{},
	}

	// Crear tres TCBs asociados al mismo PCB
	tcb1 := kerneltypes.TCB{
		TID:       0,
		Prioridad: 0,
		FatherPCB: &pcb,
	}
	tcb2 := kerneltypes.TCB{
		TID:       1,
		Prioridad: 0,
		FatherPCB: &pcb,
	}
	tcb3 := kerneltypes.TCB{
		TID:       2,
		Prioridad: 0,
		FatherPCB: &pcb,
	}

	// Simulamos que el primer hilo (tcb1) es el que está en ejecución
	kernelglobals.ExecStateThread = tcb1

	// Log inicial del estado
	logCurrentState("Estado inicial antes de cualquier syscall")

	// Primero, creamos un mutex para usarlo en la prueba
	argsCreate := []string{}
	syscallCreate := syscalls.Syscall{
		Type:      syscalls.MutexCreate,
		Arguments: argsCreate,
	}

	err := ExecuteSyscall(syscallCreate)
	if err != nil {
		t.Fatalf("Error al ejecutar syscall MUTEX_CREATE: %v", err)
	}
	logCurrentState("Después de crear el mutex")

	// Obtener el ID del mutex recién creado
	mutexID := len(pcb.CreatedMutexes)

	// Ahora intentamos que el primer hilo (tcb1) tome el mutex
	argsLock1 := []string{fmt.Sprintf("%d", mutexID)}
	syscallLock1 := syscalls.Syscall{
		Type:      syscalls.MutexLock,
		Arguments: argsLock1,
	}

	err = ExecuteSyscall(syscallLock1)
	if err != nil {
		t.Fatalf("Error al ejecutar syscall MUTEX_LOCK (primer hilo): %v", err)
	}
	logCurrentState("Después de que TID 0 toma el mutex")

	// Ahora intentamos que el segundo hilo (tcb2) tome el mismo mutex y se bloquee
	kernelglobals.ExecStateThread = tcb2

	argsLock2 := []string{fmt.Sprintf("%d", mutexID)}
	syscallLock2 := syscalls.Syscall{
		Type:      syscalls.MutexLock,
		Arguments: argsLock2,
	}

	err = ExecuteSyscall(syscallLock2)
	if err != nil {
		t.Fatalf("Error al ejecutar syscall MUTEX_LOCK (segundo hilo): %v", err)
	}
	logCurrentState("Después de que TID 1 intenta tomar el mutex y se bloquea")

	// Verificar que el segundo hilo (tcb2) está bloqueado
	mutexWrapper, exists := kernelglobals.GlobalMutexRegistry[mutexID]
	if !exists {
		t.Fatalf("No se encontró el mutex con ID <%d> en el registro global", mutexID)
	}
	if len(mutexWrapper.BlockedTCBs) != 1 {
		t.Fatalf("Se esperaba 1 hilo bloqueado en el mutex con ID <%d>, pero se encontraron %d", mutexID, len(mutexWrapper.BlockedTCBs))
	}

	// Ahora intentamos desbloquear el mutex con el primer hilo (tcb1)
	kernelglobals.ExecStateThread = tcb1

	argsUnlock := []string{fmt.Sprintf("%d", mutexID)}
	syscallUnlock := syscalls.Syscall{
		Type:      syscalls.MutexUnlock,
		Arguments: argsUnlock,
	}

	err = ExecuteSyscall(syscallUnlock)
	if err != nil {
		t.Fatalf("Error al ejecutar syscall MUTEX_UNLOCK: %v", err)
	}
	logCurrentState("Después de que TID 0 libera el mutex")

	// Verificar que el mutex ha sido reasignado al segundo hilo (tcb2)
	if mutexWrapper.AssignedTID != tcb2.TID {
		t.Fatalf("El mutex con ID <%d> no fue reasignado al TID <%d> como se esperaba", mutexID, tcb2.TID)
	}

	// Ahora intentamos que el tercer hilo (tcb3) tome el mismo mutex y se bloquee
	kernelglobals.ExecStateThread = tcb3

	argsLock3 := []string{fmt.Sprintf("%d", mutexID)}
	syscallLock3 := syscalls.Syscall{
		Type:      syscalls.MutexLock,
		Arguments: argsLock3,
	}

	err = ExecuteSyscall(syscallLock3)
	if err != nil {
		t.Fatalf("Error al ejecutar syscall MUTEX_LOCK (tercer hilo): %v", err)
	}
	logCurrentState("Después de que TID 2 intenta tomar el mutex y se bloquea")

	// Verificar que el tercer hilo (tcb3) está bloqueado
	if len(mutexWrapper.BlockedTCBs) != 1 {
		t.Fatalf("Se esperaba 1 hilo bloqueado en el mutex con ID <%d>, pero se encontraron %d", mutexID, len(mutexWrapper.BlockedTCBs))
	}

	// Verificar el estado final de ExecStateThread
	logCurrentState("Estado final después de las pruebas de MutexUnlock")

	t.Logf("El tercer hilo con TID <%d> se bloqueó correctamente al intentar tomar el mutex con ID <%d>", tcb3.TID, mutexID)
}
*/
