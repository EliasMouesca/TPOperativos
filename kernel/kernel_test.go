package main

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"testing"
)

type SyscallRequest struct {
	Type        int      `json:"type"`
	Arguments   []string `json:"arguments"`
	Description string   `json:"description"`
}

func setup() {
	logger.ConfigureLogger("test.log", "INFO")
}

func TestProcessToReady(t *testing.T) {
	setup()
	pcb := kerneltypes.PCB{
		PID:   0,
		TIDs:  []int{0},
		Mutex: nil,
	}
	tcb := kerneltypes.TCB{
		TID:       0,
		Prioridad: 0,
		ConectPCB: &pcb,
	}
	kernelglobals.ExecStateThread = tcb

	args := []string{"testfile", "1024", "1"}
	syscall := syscalls.Syscall{
		Type:      2,
		Arguments: args,
	}

	kernelsync.WaitPlanificadorLP.Add(1)
	go func() {
		defer kernelsync.WaitPlanificadorLP.Done()
		err := ExecuteSyscall(syscall)
		if err != nil {
			logger.Error("%v", err)
		}
	}()

	kernelsync.WaitPlanificadorLP.Add(1)
	go func() {
		defer kernelsync.WaitPlanificadorLP.Done()
		processToReady()
	}()
	//kernelsync.SemCreateprocess <- 0
	// no se va a conectar con memoria pero ya le
	// estoy dando el visto para que se conecte

	// Esperamos a que finalicen todas las rutinas
	kernelsync.WaitPlanificadorLP.Wait()
}

func TestProcessToExit(t *testing.T) {
	setup()
	pcb := kerneltypes.PCB{
		PID:   0,
		TIDs:  []int{0, 1},
		Mutex: nil,
	}
	tcb := kerneltypes.TCB{
		TID:       0,
		Prioridad: 0,
		ConectPCB: &pcb,
	}
	tcb1 := kerneltypes.TCB{
		TID:       1,
		Prioridad: 0,
		ConectPCB: &pcb,
	}
	kernelglobals.ReadyStateQueue.Add(&tcb1)
	kernelglobals.ExecStateThread = tcb
	args := []string{}
	syscall := syscalls.Syscall{
		Type:      10,
		Arguments: args,
	}

	kernelsync.WaitPlanificadorLP.Add(1)
	go func() {
		defer kernelsync.WaitPlanificadorLP.Done()
		err := ExecuteSyscall(syscall)
		if err != nil {
			logger.Error("%v", err)
		}
	}()

	kernelsync.WaitPlanificadorLP.Add(1)
	go func() {
		defer kernelsync.WaitPlanificadorLP.Done()
		processToExit()
	}()
	kernelsync.SemFinishprocess <- 0

	// Esperamos a que finalicen todas las rutinas
	kernelsync.WaitPlanificadorLP.Wait()
}

func TestProcessExit(t *testing.T) {
	setup()
	pcb := kerneltypes.PCB{
		PID:   0,
		TIDs:  []int{0, 1},
		Mutex: nil,
	}
	tcb := kerneltypes.TCB{
		TID:       0,
		Prioridad: 0,
		ConectPCB: &pcb,
	}
	tcb1 := kerneltypes.TCB{
		TID:       1,
		Prioridad: 0,
		ConectPCB: &pcb,
	}
	kernelglobals.ExecStateThread = tcb
	kernelglobals.ReadyStateQueue.Add(&tcb1)
	args := []string{}
	syscall := syscalls.Syscall{
		Type:      10,
		Arguments: args,
	}

	err := ExecuteSyscall(syscall)
	if err != nil {
		logger.Error("%v", err)
	}
}

func TestMutexCreate(t *testing.T) {
	setup()

	// Crear un PCB y TCB de prueba
	pcb := kerneltypes.PCB{
		PID:   0,
		TIDs:  []int{0},
		Mutex: []int{},
	}
	tcb := kerneltypes.TCB{
		TID:       0,
		Prioridad: 0,
		ConectPCB: &pcb,
	}
	kernelglobals.ExecStateThread = tcb

	// Preparar los argumentos para la syscall MUTEX_CREATE
	args := []string{}
	syscall := syscalls.Syscall{
		Type:      syscalls.MutexCreate,
		Arguments: args,
	}

	// Ejecutar la syscall
	err := ExecuteSyscall(syscall)
	if err != nil {
		t.Fatalf("Error al ejecutar syscall MUTEX_CREATE: %v", err)
	}

	// Verificar que el mutex se ha creado en el registro global
	mutexID := len(pcb.Mutex)
	if _, exists := kernelglobals.GlobalMutexRegistry[mutexID]; !exists {
		t.Fatalf("No se encontró el mutex con ID <%d> en el registro global", mutexID)
	}

	// Verificar que el mutex ha sido añadido a la lista de mutexes del PCB
	found := false
	for _, id := range pcb.Mutex {
		if id == mutexID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("El mutex con ID <%d> no se añadió a la lista de mutexes del PCB con PID <%d>", mutexID, pcb.PID)
	}

	t.Logf("Se creó correctamente el mutex con ID <%d> para el proceso con PID <%d>", mutexID, pcb.PID)
}

// Función auxiliar para registrar el estado actual de ExecStateThread y de los hilos bloqueados
func logCurrentState(context string) {
	logger.Info("### %s ###", context)
	logger.Info("Estado actual de ExecStateThread: PID <%d>, TID <%d>, Mutex: %v",
		kernelglobals.ExecStateThread.ConectPCB.PID,
		kernelglobals.ExecStateThread.TID,
		kernelglobals.ExecStateThread.Mutex,
	)

	for mutexID, mutexWrapper := range kernelglobals.GlobalMutexRegistry {
		logger.Info("Estado del Mutex ID <%d>: AssignedTID <%d>, BlockedThreads: [", mutexID, mutexWrapper.AssignedTID)
		for _, blockedTCB := range mutexWrapper.BlockedThreads {
			logger.Info("  TID <%d> del PCB con PID <%d>", blockedTCB.TID, blockedTCB.ConectPCB.PID)
		}
		logger.Info("]")
	}
}

func TestMutexLock(t *testing.T) {
	setup()

	// Crear un PCB y TCB de prueba
	pcb := kerneltypes.PCB{
		PID:   0,
		TIDs:  []int{0, 1}, // Hilos con TID 0 y 1
		Mutex: []int{},
	}

	// Crear dos TCBs asociados al mismo PCB
	tcb1 := kerneltypes.TCB{
		TID:       0,
		Prioridad: 0,
		ConectPCB: &pcb,
	}
	tcb2 := kerneltypes.TCB{
		TID:       1,
		Prioridad: 0,
		ConectPCB: &pcb,
	}

	// Simulamos que el primer hilo (tcb1) es el que está en ejecución
	kernelglobals.ExecStateThread = tcb1

	// Mostrar estado inicial
	logCurrentState("Estado inicial")

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

	// Obtener el ID del mutex recién creado
	mutexID := len(pcb.Mutex)

	// Mostrar estado después de crear el mutex
	logCurrentState(fmt.Sprintf("Después de crear el mutex con ID <%d>", mutexID))

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

	// Mostrar estado después de que el primer hilo toma el mutex
	logCurrentState("Después de que el primer hilo (tcb1) toma el mutex")

	// Verificar que el mutex ha sido asignado al primer hilo
	mutexWrapper, exists := kernelglobals.GlobalMutexRegistry[mutexID]
	if !exists {
		t.Fatalf("No se encontró el mutex con ID <%d> en el registro global", mutexID)
	}
	if mutexWrapper.AssignedTID != tcb1.TID {
		t.Fatalf("El mutex con ID <%d> no fue asignado al TID <%d> como se esperaba", mutexID, tcb1.TID)
	}

	// Ahora intentamos que el segundo hilo (tcb2) tome el mismo mutex
	kernelglobals.ExecStateThread = tcb2

	// Mostrar estado después de cambiar el hilo en ejecución a tcb2
	logCurrentState("Después de cambiar el hilo en ejecución a tcb2")

	argsLock2 := []string{fmt.Sprintf("%d", mutexID)}
	syscallLock2 := syscalls.Syscall{
		Type:      syscalls.MutexLock,
		Arguments: argsLock2,
	}

	err = ExecuteSyscall(syscallLock2)
	if err != nil {
		t.Fatalf("Error al ejecutar syscall MUTEX_LOCK (segundo hilo): %v", err)
	}

	// Mostrar estado después de que el segundo hilo intenta tomar el mutex
	logCurrentState("Después de que el segundo hilo (tcb2) intenta tomar el mutex")

	// Verificar que el segundo hilo (tcb2) está bloqueado
	blocked := false
	for _, blockedTCB := range mutexWrapper.BlockedThreads {
		if blockedTCB.TID == tcb2.TID {
			blocked = true
			break
		}
	}

	if !blocked {
		t.Fatalf("El segundo hilo con TID <%d> no se bloqueó como se esperaba", tcb2.TID)
	}

	t.Logf("El segundo hilo con TID <%d> se bloqueó correctamente al intentar tomar el mutex con ID <%d>", tcb2.TID, mutexID)
}

func TestMutexUnlock(t *testing.T) {
	setup()

	// Crear un PCB y TCB de prueba
	pcb := kerneltypes.PCB{
		PID:   0,
		TIDs:  []int{0, 1, 2}, // Hilos con TID 0, 1 y 2
		Mutex: []int{},
	}

	// Crear tres TCBs asociados al mismo PCB
	tcb1 := kerneltypes.TCB{
		TID:       0,
		Prioridad: 0,
		ConectPCB: &pcb,
	}
	tcb2 := kerneltypes.TCB{
		TID:       1,
		Prioridad: 0,
		ConectPCB: &pcb,
	}
	tcb3 := kerneltypes.TCB{
		TID:       2,
		Prioridad: 0,
		ConectPCB: &pcb,
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
	mutexID := len(pcb.Mutex)

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
	if len(mutexWrapper.BlockedThreads) != 1 {
		t.Fatalf("Se esperaba 1 hilo bloqueado en el mutex con ID <%d>, pero se encontraron %d", mutexID, len(mutexWrapper.BlockedThreads))
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
	if len(mutexWrapper.BlockedThreads) != 1 {
		t.Fatalf("Se esperaba 1 hilo bloqueado en el mutex con ID <%d>, pero se encontraron %d", mutexID, len(mutexWrapper.BlockedThreads))
	}

	// Verificar el estado final de ExecStateThread
	logCurrentState("Estado final después de las pruebas de MutexUnlock")

	t.Logf("El tercer hilo con TID <%d> se bloqueó correctamente al intentar tomar el mutex con ID <%d>", tcb3.TID, mutexID)
}

func TestThreadExit(t *testing.T) {
	setup()

	// Crear un PCB y TCB de prueba
	pcb := kerneltypes.PCB{
		PID:   0,
		TIDs:  []int{0}, // Hilos con TID 0 y 1
		Mutex: []int{},
	}

	// Crear dos TCBs asociados al mismo PCB
	tcb1 := kerneltypes.TCB{
		TID:       0,
		Prioridad: 0,
		ConectPCB: &pcb,
	}

	// Simulamos que el primer hilo (tcb1) es el que está en ejecución
	kernelglobals.ExecStateThread = tcb1

	// Verificamos el estado inicial del hilo en ejecución
	logCurrentState("Estado inicial antes de la syscall ThreadExit")

	// Llamar a la syscall ThreadExit con el hilo actual
	args := []string{}
	syscall := syscalls.Syscall{
		Type:      syscalls.ThreadExit,
		Arguments: args,
	}

	err := ExecuteSyscall(syscall)
	if err != nil {
		t.Fatalf("Error al ejecutar syscall THREAD_EXIT: %v", err)
	}

	// Verificar que el hilo en ejecución ha sido movido a la cola de Exit
	// Comprobar si ConectPCB es nil antes de acceder a él
	if kernelglobals.ExecStateThread.TID != -1 {
		t.Fatalf("El ExecStateThread no se ha vaciado correctamente, se esperaba un TCB vacío, pero se encontró TID <%d> y PID <%d>", kernelglobals.ExecStateThread.TID, kernelglobals.ExecStateThread.ConectPCB.PID)
	}

	// Verificar que el TCB se encuentra en la cola de ExitState
	foundInExit := false
	kernelglobals.ExitStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == tcb1.TID && tcb.ConectPCB == &pcb {
			foundInExit = true
		}
	})

	if !foundInExit {
		t.Fatalf("El TCB del hilo con TID <%d> no se encontró en la cola de ExitStateQueue", tcb1.TID)
	}

	// Mostrar el estado final después de la syscall
	logCurrentState("Estado final después de la syscall ThreadExit")

	t.Logf("El hilo con TID <%d> se ha movido correctamente al estado EXIT", tcb1.TID)
}
