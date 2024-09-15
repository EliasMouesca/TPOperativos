package main

import (
	"errors"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

type syscallFunction func(args []string) error

// TODO: Dónde ponemos esto? en qué carpeta?

var syscallDescriptions = map[int]string{
	syscalls.ProcessCreate: "CREATE_PROCESS",
	syscalls.ProcessExit:   "PROCESS_EXIT",
	syscalls.ThreadCreate:  "THREAD_CREATE",
}

var syscallSet = map[int]syscallFunction{
	syscalls.ProcessCreate: ProcessCreate,
	syscalls.ProcessExit:   ProcessExit,
	syscalls.ThreadCreate:  ThreadCreate,
	// "THREAD_JOIN": THREAD_JOIN,
	// "THREAD_CANCEL": THREAD_CANCEL
	// "THREAD_EXIT": THREAD_CREATE,
	// "MUTEX_CREATE": MUTEX_CREATE,
	// "MUTEX_LOCK": MUTEX_LOCK,
	// "MUTEX_UNLOCK": MUTEX_UNLOCK,
}

func ExecuteSyscall(syscall syscalls.Syscall) error {
	syscallFunc, exists := syscallSet[syscall.Type]
	if !exists {
		return errors.New("la syscall pedida no es una syscall que el kernel entienda")
	}

	logger.Info("## (%v:%v) - Solicitó syscall: <%v>",
		kernelglobals.ExecStateThread.ConectPCB.PID,
		kernelglobals.ExecStateThread.TID,
		syscallDescriptions[syscall.Type],
	)

	err := syscallFunc(syscall.Arguments)
	if err != nil {
		return err
	}

	return nil
}

var PIDcount int = 0

func ProcessCreate(args []string) error {

	// Se crea el PCB y el Hilo 0

	var procesoCreate kerneltypes.PCB
	PIDcount++
	procesoCreate.PID = PIDcount
	procesoCreate.TIDs[0] = 0

	logger.Info("## (<%d>:<0>) Se crea el proceso - Estado: NEW", procesoCreado.PID)

	// Se agrega el proceso a NEW
	kernelglobals.NewStateQueue.Add(&procesoCreate)
	kernelsync.ChannelProcessCreate <- procesoCreate

	return nil
}

func ProcessExit(args []string) error {
	// nose si estara bien pero el valor TCB ya esta en el canal
	kernelglobals.ExecStateThread // sino usar una lista de un elemento consultar con los pibes
	if tcb.TID == 0 {             // tiene que ser el hiloMain
		conectedProcess := tcb.ConectPCB
		processToExit(conectedProcess)
	} else {
		return errors.New("El hilo que quizo eliminar el proceso, no es el hilo main")
	}

	return nil
}

func ThreadCreate(args []string) error {
	// len(Ready) forma autoincremental ajustable
	// TIDcount forma autoincremental crecientei ndeterminadamente
	// nose que opcion es mejor
	fmt.Println("Creando hilo...")

	return nil
}

func THREAD_JOIN(args []string) {
	fmt.Println("Esperando a que el hilo termine...")
}

func THREAD_CANCEL(args []string) {
	fmt.Println("Cancelando hilo...")
}

func THREAD_EXIT(args []string) {
	fmt.Println("Saliendo del hilo...")
}

func MUTEX_CREATE(args []string) {
	fmt.Println("Creando mutex...")
}

func MUTEX_LOCK(args []string) {
	fmt.Println("Bloqueando mutex...")
}

func MUTEX_UNLOCK(args []string) {
	fmt.Println("Desbloqueando mutex...")
}
