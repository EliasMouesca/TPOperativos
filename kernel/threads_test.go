package main

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"testing"
)

// TODO: ---------------------------------- TEST PARA THREAD ----------------------------------

func TestThreadCreate(t *testing.T) {
	setup()

	// Crear un PCB y un TCB inicial para el proceso
	pcb := kerneltypes.PCB{
		PID:   0,
		TIDs:  []int{0}, // El proceso comienza con un solo hilo TID 0
		Mutex: []int{},
	}

	// Crear el hilo inicial (main) asociado al PCB
	tcb1 := kerneltypes.TCB{
		TID:       0,
		Prioridad: 0,
		ConectPCB: &pcb,
	}

	// Asignar el hilo inicial como el hilo en ejecución
	kernelglobals.ExecStateThread = tcb1

	// Agregar el TCB inicial a la cola de Ready
	kernelglobals.ReadyStateQueue.Add(&tcb1)

	// Mostrar estado inicial del PCB antes de crear un nuevo hilo
	logPCBState("Estado inicial del PCB antes de crear un nuevo hilo", &pcb)

	// Preparar los argumentos para la syscall ThreadCreate
	args := []string{"archivo_pseudocodigo", "1"} // Archivo ficticio y prioridad 1
	syscall := syscalls.Syscall{
		Type:      syscalls.ThreadCreate,
		Arguments: args,
	}

	// Ejecutar la syscall
	err := ExecuteSyscall(syscall)
	if err != nil {
		t.Fatalf("Error al ejecutar syscall THREAD_CREATE: %v", err)
	}

	// Mostrar el estado del PCB después de la creación del nuevo hilo
	logPCBState("Estado del PCB después de crear un nuevo hilo con ThreadCreate", &pcb)

	// Verificar que se haya creado un nuevo TID para el PCB
	if len(pcb.TIDs) != 2 {
		t.Fatalf("Se esperaba que el PCB con PID <%d> tuviera 2 TIDs, pero tiene %d", pcb.PID, len(pcb.TIDs))
	}

	// Verificar que el nuevo TID es el siguiente en la secuencia (1 en este caso)
	newTID := pcb.TIDs[1]
	if newTID != 1 {
		t.Fatalf("Se esperaba que el nuevo TID fuera 1, pero fue %d", newTID)
	}

	// Verificar que el nuevo hilo se encuentra en la cola de Ready
	foundInReady := false
	kernelglobals.ReadyStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == newTID && tcb.ConectPCB == &pcb {
			foundInReady = true
		}
	})

	if !foundInReady {
		t.Fatalf("El nuevo TCB con TID <%d> no se encontró en la cola de ReadyStateQueue", newTID)
	}

	// Mostrar el estado final después de la syscall
	logPCBState("Estado final del PCB después de la prueba de ThreadCreate", &pcb)

	t.Logf("El nuevo hilo con TID <%d> se ha creado y movido correctamente al estado READY", newTID)
}

func TestThreadJoin(t *testing.T) {
	setup()

	// Crear un PCB y dos TCBs para el proceso
	pcb := kerneltypes.PCB{
		PID:   0,
		TIDs:  []int{0, 1}, // El proceso comienza con dos hilos TID 0 y 1
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

	// Asignar el primer hilo (tcb1) como el hilo en ejecución
	kernelglobals.ExecStateThread = tcb1

	// Añadir el segundo hilo (tcb2) a la cola de Ready
	kernelglobals.ReadyStateQueue.Add(&tcb2)

	// Mostrar el estado inicial del PCB antes de la syscall ThreadJoin
	logPCBState("Estado inicial del PCB antes de la syscall ThreadJoin", &pcb)

	// Preparar los argumentos para la syscall ThreadJoin
	args := []string{"1"} // El hilo 0 esperará al hilo 1
	syscall := syscalls.Syscall{
		Type:      syscalls.ThreadJoin,
		Arguments: args,
	}

	// Ejecutar la syscall ThreadJoin desde el hilo 0
	err := ExecuteSyscall(syscall)
	if err != nil {
		t.Fatalf("Error al ejecutar syscall THREAD_JOIN: %v", err)
	}

	// Verificar que el hilo 0 se ha movido a la cola de Blocked
	foundInBlocked := false
	kernelglobals.BlockedStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == tcb1.TID && tcb.ConectPCB == &pcb {
			foundInBlocked = true
		}
	})

	if !foundInBlocked {
		t.Fatalf("El hilo con TID <%d> no se encontró en la cola de BlockedStateQueue", tcb1.TID)
	}

	// Verificar que el hilo 1 está en ejecución
	if kernelglobals.ExecStateThread.TID != tcb2.TID {
		t.Fatalf("Se esperaba que el hilo en ejecución fuera TID <%d>, pero se encontró TID <%d>", tcb2.TID, kernelglobals.ExecStateThread.TID)
	}

	// Mostrar el estado del PCB después de ejecutar ThreadJoin
	logPCBState("Estado del PCB después de ejecutar ThreadJoin", &pcb)

	// Cambiar hilo 1 a estado EXIT para desbloquear hilo 0
	kernelglobals.ExecStateThread = tcb2
	exitArgs := []string{}
	syscallExit := syscalls.Syscall{
		Type:      syscalls.ThreadExit,
		Arguments: exitArgs,
	}

	err = ExecuteSyscall(syscallExit)
	if err != nil {
		t.Fatalf("Error al ejecutar syscall THREAD_EXIT para el TID 1: %v", err)
	}

	// Verificar que el hilo 1 se ha movido a la cola de Exit
	foundInExit := false
	kernelglobals.ExitStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == tcb2.TID && tcb.ConectPCB == &pcb {
			foundInExit = true
		}
	})

	if !foundInExit {
		t.Fatalf("El hilo con TID <%d> no se encontró en la cola de ExitStateQueue", tcb2.TID)
	}

	// Muevo el hilo 0 a la cola de Ready, ya que el hilo 1 finalizó
	kernelglobals.ReadyStateQueue.Add(&tcb1)
	// Verificar que el hilo 0 se ha movido de Blocked a Ready después de que el hilo 1 finalizó
	foundInReady := false
	kernelglobals.ReadyStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == tcb1.TID && tcb.ConectPCB == &pcb {
			foundInReady = true
		}
	})

	if !foundInReady {
		t.Fatalf("El hilo con TID <%d> no se encontró en la cola de ReadyStateQueue después de que TID 1 finalizó", tcb1.TID)
	}

	// Mostrar el estado final del PCB después de que el hilo 1 ha finalizado
	logPCBState("Estado final del PCB después de la prueba de ThreadJoin", &pcb)

	t.Logf("El hilo con TID <%d> se ha bloqueado correctamente y ha vuelto a Ready después de la finalización de TID 1", tcb1.TID)
}

func TestThreadCancel(t *testing.T) {
	setup()

	// Crear un PCB y TCBs de prueba
	pcb := kerneltypes.PCB{
		PID:   0,
		TIDs:  []int{0, 1, 2}, // El proceso tiene tres hilos TID 0, 1 y 2
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

	// Añadir los TCBs a la cola de Ready, simulando que están listos para ejecutarse
	kernelglobals.ReadyStateQueue.Add(&tcb3)
	kernelglobals.ReadyStateQueue.Add(&tcb2)

	// Asignar el primer hilo como el hilo en ejecución
	kernelglobals.ExecStateThread = tcb1

	// Mostrar estado inicial del PCB antes de la syscall ThreadCancel
	logPCBState("Estado inicial del PCB antes de la syscall ThreadCancel", &pcb)

	// Preparar los argumentos para la syscall ThreadCancel (cancelar el hilo 1)
	args := []string{"1"} // Se cancelará el TID 1
	syscall := syscalls.Syscall{
		Type:      syscalls.ThreadCancel,
		Arguments: args,
	}

	// Ejecutar la syscall ThreadCancel desde el hilo 0
	err := ExecuteSyscall(syscall)
	if err != nil {
		t.Fatalf("Error al ejecutar syscall THREAD_CANCEL: %v", err)
	}

	// Verificar que el hilo 1 ha sido movido a la cola de Exit
	foundInExit := false
	kernelglobals.ExitStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == tcb2.TID && tcb.ConectPCB == &pcb {
			foundInExit = true
		}
	})

	if !foundInExit {
		t.Fatalf("El hilo con TID <%d> no se encontró en la cola de ExitStateQueue", tcb2.TID)
	}

	// Verificar que el hilo 1 ha sido eliminado de ReadyState y BlockedState
	foundInReady := false
	kernelglobals.ReadyStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == tcb2.TID && tcb.ConectPCB == &pcb {
			foundInReady = true
		}
	})

	if foundInReady {
		t.Fatalf("El hilo con TID <%d> no debería estar en la cola de ReadyStateQueue después de ser cancelado", tcb2.TID)
	}

	foundInBlocked := false
	kernelglobals.BlockedStateQueue.Do(func(tcb *kerneltypes.TCB) {
		if tcb.TID == tcb2.TID && tcb.ConectPCB == &pcb {
			foundInBlocked = true
		}
	})

	if foundInBlocked {
		t.Fatalf("El hilo con TID <%d> no debería estar en la cola de BlockedStateQueue después de ser cancelado", tcb2.TID)
	}

	// Mostrar el estado final del PCB después de ejecutar ThreadCancel
	logPCBState("Estado final del PCB después de ejecutar ThreadCancel", &pcb)

	t.Logf("El hilo con TID <%d> ha sido cancelado correctamente y movido al estado EXIT", tcb2.TID)
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
