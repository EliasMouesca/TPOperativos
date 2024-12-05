package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"net/http"
)

func memoryGiveMeExecutionContext(thread types.Thread) (ectx types.ExecutionContext, err error) {
	logger.Info("T%v P%v - Solicito contexto de ejecución", thread.TID, thread.PID)
	url := fmt.Sprintf("http://%v:%v/memoria/getContext?pid=%v&tid=%v",
		config.MemoryAddress, config.MemoryPort, thread.PID, thread.TID)
	err = receiveThatFromHere(url, &ectx)
	return ectx, err
}

func memoryUpdateExecutionContext(thread types.Thread, ectx types.ExecutionContext) error {
	logger.Info("T%v P%v - Actualizo contexto de ejecución", thread.TID, thread.PID)
	url := fmt.Sprintf("http://%v:%v/memoria/saveContext?tid=%v&pid=%v", config.MemoryAddress, config.MemoryPort, thread.TID, thread.PID)
	err := sendThisToThere(url, ectx)
	logger.Debug("Se ha avisado a Memoria del update execution context")
	return err
}

func memoryGiveMeInstruction(thread types.Thread, pc uint32) (instruction string, err error) {
	logger.Info("T%v P%v - FETCH PC=%v", thread.TID, thread.PID, pc)
	url := fmt.Sprintf("http://%v:%v/memoria/getInstruction?tid=%v&pid=%v&pc=%v",
		config.MemoryAddress, config.MemoryPort, thread.TID, thread.PID, pc)

	err = receiveThatFromHere(url, &instruction)
	if err != nil {
		return "", err
	}

	return instruction, nil
}

func memoryRead(thread types.Thread, physicalDirection uint32) (uint32, error) {
	logger.Info("T%v P%v - LEER -> %v", thread.TID, thread.PID, physicalDirection)
	url := fmt.Sprintf("http://%v:%v/memoria/readMem?tid=%v&pid=%v&addr=%v",
		config.MemoryAddress, config.MemoryPort, thread.TID, thread.PID, physicalDirection)
	var valueRead uint32
	err := receiveThatFromHere(url, &valueRead)
	return valueRead, err
}

func memoryWrite(thread types.Thread, physicalDirection uint32, data uint32) error {
	logger.Info("T%v P%v - ESCRIBIR -> %v", thread.TID, thread.PID, physicalDirection)
	url := fmt.Sprintf("http://%v:%v/memoria/writeMem?tid=%v&pid=%v&addr=%v",
		config.MemoryAddress, config.MemoryPort, thread.TID, thread.PID, physicalDirection)
	err := sendThisToThere(url, data)
	return err
}

func kernelYourProcessFinished(thread types.Thread, interruptReceived types.Interruption) (err error) {
	logger.Debug("Kernel, tu proceso terminó! TID: %v, PID: %v", thread.TID, thread.PID)
	logger.Debug("Int. received - %v", interruptReceived.Description)
	url := fmt.Sprintf("http://%v:%v/kernel/process_finished?pid=%v&tid=%v",
		config.KernelAddress, config.KernelPort, thread.PID, thread.TID)
	err = sendThisToThere(url, interruptReceived)
	logger.Debug("Se a avisado a Kernel del process finished")
	return err
}

// Helpers
func sendThisToThere(url string, data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	response, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return nil
}

func receiveThatFromHere(url string, data any) (err error) {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, data)
	if err != nil {
		return err
	}

	return nil

}
