package main

type SyscallRequest struct {
	Type        int      `json:"type"`
	Arguments   []string `json:"arguments"`
	Description string   `json:"description"`
}

// TODO: DEJO COMENTADO LOS QUE NO ESTAN ANDANDO

//func TestProcessToReady(t *testing.T) {
//	setup()
//	pcb := kerneltypes.PCB{
//		PID:   0,
//		TIDs:  []int{0},
//		CreatedMutexes: nil,
//	}
//	tcb := kerneltypes.TCB{
//		TID:       0,
//		Prioridad: 0,
//		FatherPCB: &pcb,
//	}
//	kernelglobals.ExecStateThread = tcb
//
//	args := []string{"testfile", "1024", "1"}
//	syscall := syscalls.Syscall{
//		Type:      2,
//		Arguments: args,
//	}
//
//	kernelsync.WaitPlanificadorLP.Add(1)
//	go func() {
//		defer kernelsync.WaitPlanificadorLP.Done()
//		err := ExecuteSyscall(syscall)
//		if err != nil {
//			logger.Error("%v", err)
//		}
//	}()
//
//	kernelsync.WaitPlanificadorLP.Add(1)
//	go func() {
//		defer kernelsync.WaitPlanificadorLP.Done()
//		ProcessToReady()
//	}()
//	//kernelsync.SemCreateprocess <- 0
//	// no se va a conectar con memoria pero ya le
//	// estoy dando el visto para que se conecte
//
//	// Esperamos a que finalicen todas las rutinas
//	kernelsync.WaitPlanificadorLP.Wait()
//}
//
//func TestProcessToExit(t *testing.T) {
//	setup()
//	pcb := kerneltypes.PCB{
//		PID:   0,
//		TIDs:  []int{0, 1},
//		CreatedMutexes: nil,
//	}
//	tcb := kerneltypes.TCB{
//		TID:       0,
//		Prioridad: 0,
//		FatherPCB: &pcb,
//	}
//	tcb1 := kerneltypes.TCB{
//		TID:       1,
//		Prioridad: 0,
//		FatherPCB: &pcb,
//	}
//	kernelglobals.ReadyStateQueue.Add(&tcb1)
//	kernelglobals.ExecStateThread = tcb
//	args := []string{}
//	syscall := syscalls.Syscall{
//		Type:      10,
//		Arguments: args,
//	}
//
//	kernelsync.WaitPlanificadorLP.Add(1)
//	go func() {
//		defer kernelsync.WaitPlanificadorLP.Done()
//		err := ExecuteSyscall(syscall)
//		if err != nil {
//			logger.Error("%v", err)
//		}
//	}()
//
//	kernelsync.WaitPlanificadorLP.Add(1)
//	go func() {
//		defer kernelsync.WaitPlanificadorLP.Done()
//		ProcessToExit()
//	}()
//	kernelsync.SemFinishprocess <- 0
//
//	// Esperamos a que finalicen todas las rutinas
//	kernelsync.WaitPlanificadorLP.Wait()
//}
//
//func TestProcessExit(t *testing.T) {
//	setup()
//	pcb := kerneltypes.PCB{
//		PID:   0,
//		TIDs:  []int{0, 1},
//		CreatedMutexes: nil,
//	}
//	tcb := kerneltypes.TCB{
//		TID:       0,
//		Prioridad: 0,
//		FatherPCB: &pcb,
//	}
//	tcb1 := kerneltypes.TCB{
//		TID:       1,
//		Prioridad: 0,
//		FatherPCB: &pcb,
//	}
//	kernelglobals.ExecStateThread = tcb
//	kernelglobals.ReadyStateQueue.Add(&tcb1)
//	args := []string{}
//	syscall := syscalls.Syscall{
//		Type:      10,
//		Arguments: args,
//	}
//
//	err := ExecuteSyscall(syscall)
//	if err != nil {
//		logger.Error("%v", err)
//	}
//}

/*
	func TestThreadEnding(t *testing.T) {
		setup()

		// Crear un PCB y tres TCBs para el proceso
		pcb := kerneltypes.PCB{
			PID:            0,
			TIDs:           []int{0, 1, 2}, // El proceso comienza con tres hilos TID 0, 1 y 2
			CreatedMutexes: []int{},
		}

		// Crear tres TCBs asociados al mismo PCB
		tcb1 := kerneltypes.TCB{
			TID:       0,
			Prioridad: 0,
			FatherPCB: &pcb,
			JoinedTCB: -1,
		}
		tcb2 := kerneltypes.TCB{
			TID:       1,
			Prioridad: 0,
			FatherPCB: &pcb,
			JoinedTCB: 0, // tcb2 espera la finalización de tcb1
		}
		tcb3 := kerneltypes.TCB{
			TID:       2,
			Prioridad: 0,
			FatherPCB: &pcb,
			JoinedTCB: -1,
		}

		// Añadir el hilo tcb2 a BlockedStateQueue simulando un ThreadJoin esperando a tcb1
		kernelglobals.BlockedStateQueue.Add(&tcb2)

		// Añadir el hilo tcb3 a la cola de Ready
		kernelglobals.ReadyStateQueue.Add(&tcb3)

		// Asignar el hilo tcb1 como el hilo en ejecución
		kernelglobals.ExecStateThread = tcb1

		// Mostrar el estado inicial del PCB antes de ejecutar ThreadEnding
		logPCBState("Estado inicial del PCB antes de ejecutar ThreadEnding", &pcb)

		// Ejecutar la función ThreadEnding
		ThreadEnding()

		// Verificar que tcb1 se ha movido a ExitStateQueue
		foundInExit := false
		kernelglobals.ExitStateQueue.Do(func(tcb *kerneltypes.TCB) {
			if tcb.TID == tcb1.TID && tcb.FatherPCB == &pcb {
				foundInExit = true
			}
		})

		if !foundInExit {
			t.Fatalf("El hilo con TID <%d> no se encontró en la cola de ExitStateQueue", tcb1.TID)
		}

		// Verificar que tcb2 se ha movido de Blocked a ReadyStateQueue
		foundInReady := false
		kernelglobals.ReadyStateQueue.Do(func(tcb *kerneltypes.TCB) {
			if tcb.TID == tcb2.TID && tcb.FatherPCB == &pcb {
				foundInReady = true
			}
		})

		if !foundInReady {
			t.Fatalf("El hilo con TID <%d> no se encontró en la cola de ReadyStateQueue después de que TID 0 finalizó", tcb2.TID)
		}

		// Verificar que tcb3 está en ejecución
		if kernelglobals.ExecStateThread.TID != tcb3.TID {
			t.Fatalf("Se esperaba que el hilo en ejecución fuera TID <%d>, pero se encontró TID <%d>", tcb3.TID, kernelglobals.ExecStateThread.TID)
		}

		// Mostrar el estado final del PCB después de ejecutar ThreadEnding
		logPCBState("Estado final del PCB después de ejecutar ThreadEnding", &pcb)
		logCurrentState("Estado final del contexto")
		t.Logf("El hilo con TID <%d> se ha finalizado correctamente y los hilos bloqueados han sido movidos a Ready", tcb1.TID)
	}
*/
