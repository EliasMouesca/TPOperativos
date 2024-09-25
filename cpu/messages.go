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

func memoryGiveMeExecutionContext(tid types.Thread) (ectx types.ExecutionContext, err error) {
	logger.Info("T%v P%v - Solicito contexto de ejecución", tid.TID, tid.PID)

	url := fmt.Sprintf("http://%v:%v", config.MemoryAddress, config.MemoryPort)

	data, err := json.Marshal(tid)
	if err != nil {
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &ectx)
	if err != nil {
		return
	}

	return ectx, nil
}

func memoryUpdateExecutionContext(tid types.Thread, ectx types.ExecutionContext) error {
	logger.Info("T%v P%v - Actualizo contexto de ejecución", tid.TID, tid.PID)

	// TODO: Hablar con memoria

	return nil
}

// TODO: Qué onda las memory address? Qué son? uint32?
func memoryIsThisAddressOk(tid types.Thread, physicalAdrress uint32) (bool, error) {
	// TODO: Llamar a memoria

	return true, nil
}

func memoryGiveMeInstruction(thread types.Thread, pc uint32) (string, error) {
	logger.Info("T%v P%v - FETCH PC=%v", thread.TID, thread.PID, pc)

	// TODO: Contactar a memoria

	return "SET AX, 1", nil

}

// La consigna dice "4 bytes" creo, pero vamos a usar uint32 para ahorrarnos la endianess
func memoryRead(thread types.Thread, physicalDirection uint32) (uint32, error) {
	logger.Info("T%v P%v - LEER -> %v", thread.TID, thread.PID, physicalDirection)

	// TODO: Charlar con memoria

	return 0xdeadbeef, nil
}

func memoryWrite(thread types.Thread, physicalDirection uint32, data uint32) error {
	logger.Info("T%v P%v - ESCRIBIR -> %v", thread.TID, thread.PID, physicalDirection)

	// TODO: Chamuyarse a la memoria

	return nil
}

// TODO: Qué le mando el body del post? La interrupción sí, pero cómo, el int?
func kernelYourProcessFinished(thread types.Thread, interruptReceived types.Interruption) (err error) {
	logger.Debug("Kernel, tu proceso terminó! TID: %v, PID: %v", thread.TID, thread.PID)
	logger.Debug("Int. received - %v", interruptReceived.Description)

	//url := fmt.Sprintf("http://%v:%v/kernel/process_finished", config.KernelAddress, config.KernelPort)

	return nil
}
