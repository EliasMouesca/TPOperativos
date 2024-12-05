package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"strconv"
	"strings"
)

// An Instruction is basically a function that takes an ectx and a slice of parameters (strings)
// and returns the modified ectx and any errors encountered.
type Instruction func(context *types.ExecutionContext, arguments []string) error

var instructionSet = map[string]Instruction{
	"SET":       setInstruction,
	"READ_MEM":  readMemInstruction,
	"WRITE_MEM": writeMemInstruction,
	"SUM":       sumInstruction,
	"SUB":       subInstruction,
	"JNZ":       jnzInstruction,
	"LOG":       logInstruction,
	// De acá en adelante son syscalls y las soluciona querido kernel
	"DUMP_MEMORY":    dumpMemoryInstruction,
	"IO":             ioInstruction,
	"PROCESS_CREATE": processCreateInstruction,
	"THREAD_CREATE":  threadCreateInstruction,
	"THREAD_JOIN":    threadJoinInstruction,
	"THREAD_CANCEL":  threadCancelInstruction,
	"MUTEX_CREATE":   mutexCreateInstruction,
	"MUTEX_LOCK":     mutexLockInstruction,
	"MUTEX_UNLOCK":   mutexUnlockInstruction,
	"THREAD_EXIT":    threadExitInstruction,
	"PROCESS_EXIT":   processExitInstruction,
}

func writeMemInstruction(context *types.ExecutionContext, arguments []string) error {
	dataRegister, err := context.GetRegister(arguments[1])
	if err != nil {
		return err
	}

	virtualAddressRegister, err := context.GetRegister(arguments[0])
	if err != nil {
		return err
	}

	physicalAddress := context.MemoryBase + *virtualAddressRegister
	logger.Debug("Escribiendo en physicalAddres: %v", physicalAddress)
	if *virtualAddressRegister >= context.MemorySize {
		logger.Warn("Se trató de escribir una dirección no perteneciente al proceso! Interrumpiendo...")
		interruptionChannel <- types.Interruption{
			Type:        types.InterruptionSegFault,
			Description: "La dirección no forma parte del espacio del memoria del proceso"}

		logger.Debug("Intentadno liberar mutexInterruption")
		MutexInterruption.Unlock()
		logger.Debug("Liberado mutexInterruption")
		return nil
	}

	err = memoryWrite(*currentThread, physicalAddress, *dataRegister)
	if err != nil {
		return err
	}

	logger.Info("## P%v T%v - Escribió '%v' en la dirección física <%v>",
		currentThread.PID, currentThread.TID, *dataRegister, physicalAddress)

	return nil
}

func readMemInstruction(context *types.ExecutionContext, arguments []string) error {
	dataRegister, err := context.GetRegister(arguments[0])
	logger.Debug("DataRegister: %v", *dataRegister)
	if err != nil {
		return err
	}

	virtualAddressRegister, err := context.GetRegister(arguments[1])
	logger.Debug("VirtualAddressRegister: %v", *virtualAddressRegister)
	if err != nil {
		return err
	}
	logger.Debug("MemoryBase: %v", context.MemoryBase)
	physicalAddress := context.MemoryBase + *virtualAddressRegister
	logger.Debug("Physical Address: %v", physicalAddress)

	if *virtualAddressRegister >= context.MemorySize {
		logger.Warn("Se trató de leer una dirección no perteneciente al proceso! Interrumpiendo...")
		interruptionChannel <- types.Interruption{
			Type:        types.InterruptionSegFault,
			Description: "La dirección no forma parte del espacio del memoria del proceso"}

		logger.Debug("Intentadno liberar mutexInterruption")
		MutexInterruption.Unlock()
		logger.Debug("Liberado mutexInterruption")
		return nil
	}

	if currentThread == nil {
		logger.Error("Se mando a ejecutar la instrucción readMemory pero no hay ningún hilo en ejecución ?")
		return nil
	}

	*dataRegister, err = memoryRead(*currentThread, physicalAddress)
	if err != nil {
		return err
	}
	logger.Info("## P%v T%v - Leyó '%v' de la dirección física <%v>",
		currentThread.PID, currentThread.TID, *dataRegister, physicalAddress)

	return nil
}

func jnzInstruction(context *types.ExecutionContext, arguments []string) error {
	if err := checkArguments(arguments, 2); err != nil {
		return err
	}

	register, err := context.GetRegister(arguments[0])
	if err != nil {
		return err
	}

	jump, err := strconv.Atoi(arguments[1])
	if err != nil {
		return err
	}

	if *register != 0 {
		context.Pc = uint32(jump)
		logger.Trace("actualizando PC: %v", context.Pc)
	}

	return nil

}

func sumInstruction(context *types.ExecutionContext, args []string) error {
	if err := checkArguments(args, 2); err != nil {
		return err
	}

	firstRegister, err := context.GetRegister(args[0])
	if err != nil {
		return err
	}

	secondRegister, err := context.GetRegister(args[1])
	if err != nil {
		return err
	}

	*firstRegister = *firstRegister + *secondRegister

	return nil

}

func subInstruction(context *types.ExecutionContext, args []string) error {
	if err := checkArguments(args, 2); err != nil {
		return err
	}

	firstRegister, err := context.GetRegister(args[0])
	if err != nil {
		return err
	}

	secondRegister, err := context.GetRegister(args[1])
	if err != nil {
		return err
	}

	*firstRegister = *firstRegister - *secondRegister

	return nil

}

func setInstruction(ctx *types.ExecutionContext, args []string) error {
	// Check number of arguments
	if err := checkArguments(args, 2); err != nil {
		return err
	}

	// Get the register to modify
	reg, err := ctx.GetRegister(args[0])
	if err != nil {
		return err
	}

	// Try parsing second argument as int
	i, err := strconv.Atoi(args[1])
	if err != nil {
		// Ok not int, but is it a register?
		if reg2, err := ctx.GetRegister(args[1]); err != nil {
			// It is not a register nor an int
			return errors.New("no se pudo parsear '" + args[1] + "' como un entero o un registro")
		} else {
			// If it IS a register, set it to that value
			*reg = *reg2
		}
	} else {
		// Set the register
		*reg = uint32(i)
	}

	return nil
}

func logInstruction(ctx *types.ExecutionContext, args []string) error {
	if err := checkArguments(args, 1); err != nil {
		return err
	}

	registerString := args[0]

	register, err := ctx.GetRegister(registerString)
	if err != nil {
		return err
	}

	logger.Info("Logging register '%v': %v", registerString, *register)
	fmt.Println(*register)
	return nil
}

func checkArguments(args []string, correctNumberOfArgs int) error {
	if len(args) != correctNumberOfArgs {
		return errors.New("se recibió una cantidad de argumentos no válida")
	}
	return nil
}

// A partir de acá las syscalls
func doSyscall(ctx types.ExecutionContext, syscall syscalls.Syscall) error {
	interruptionChannel <- types.Interruption{
		Type:        types.InterruptionSyscall,
		Description: "Interrupción por syscall",
	}

	url := fmt.Sprintf("http://%v:%v/kernel/syscall", config.KernelAddress, config.KernelPort)
	jsonData, err := json.Marshal(syscall)
	if err != nil {
		return fmt.Errorf("error al empaquetar syscall: %v", err)
	}

	logger.Debug("Liberando MutexInterruption")
	MutexInterruption.Unlock()
	logger.Debug("MutexInterruption liberado")

	resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error al enviar syscall al kernel: %v", err)
	}

	logger.Debug("Syscall enviada al kernel")

	return nil

}

func mutexCreateInstruction(context *types.ExecutionContext, arguments []string) error {
	if err := doSyscall(
		*context,
		syscalls.New(syscalls.MutexCreate, arguments),
	); err != nil {
		return err
	}

	return nil
}

func processExitInstruction(context *types.ExecutionContext, arguments []string) error {
	if err := doSyscall(
		*context,
		syscalls.New(syscalls.ProcessExit, arguments),
	); err != nil {
		return err
	}

	return nil
}

func threadExitInstruction(context *types.ExecutionContext, arguments []string) error {
	if err := doSyscall(
		*context,
		syscalls.New(syscalls.ThreadExit, arguments),
	); err != nil {
		return err
	}

	return nil
}

func mutexLockInstruction(context *types.ExecutionContext, arguments []string) error {
	if err := doSyscall(
		*context,
		syscalls.New(syscalls.MutexLock, arguments),
	); err != nil {
		return err
	}

	return nil

}

func mutexUnlockInstruction(context *types.ExecutionContext, arguments []string) error {
	if err := doSyscall(
		*context,
		syscalls.New(syscalls.MutexUnlock, arguments),
	); err != nil {
		return err
	}

	return nil

}

func threadCancelInstruction(context *types.ExecutionContext, arguments []string) error {
	if err := doSyscall(
		*context,
		syscalls.New(syscalls.ThreadCancel, arguments),
	); err != nil {
		return err
	}

	return nil

}

func threadCreateInstruction(context *types.ExecutionContext, arguments []string) error {
	if err := doSyscall(
		*context,
		syscalls.New(syscalls.ThreadCreate, arguments),
	); err != nil {
		return err
	}

	return nil

}

func threadJoinInstruction(context *types.ExecutionContext, arguments []string) error {
	if err := doSyscall(
		*context,
		syscalls.New(syscalls.ThreadJoin, arguments),
	); err != nil {
		return err
	}

	return nil
}

func processCreateInstruction(context *types.ExecutionContext, arguments []string) error {
	if err := doSyscall(
		*context,
		syscalls.New(syscalls.ProcessCreate, arguments),
	); err != nil {
		return err
	}

	return nil
}

func ioInstruction(context *types.ExecutionContext, arguments []string) error {
	if err := doSyscall(
		*context,
		syscalls.New(syscalls.IO, arguments),
	); err != nil {
		return err
	}

	return nil

}

func dumpMemoryInstruction(context *types.ExecutionContext, arguments []string) error {
	if err := doSyscall(
		*context,
		syscalls.New(syscalls.DumpMemory, arguments),
	); err != nil {
		return err
	}

	return nil

}
