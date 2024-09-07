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
	logger.Info("T%v P%v - Solicito contexto de ejecución", tid.Tid, tid.Pid)

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

func memoryGiveMeInstruction(thread types.Thread, pc uint32) (string, error) {
	logger.Info("T%v P%v - FETCH PC=%v", thread.Tid, thread.Pid, pc)

	// TODO: Contactar a memoria

	return "SET AX, 1", nil

}

func memoryUpdateExecutionContext(tid types.Thread, ectx types.ExecutionContext) error {
	logger.Info("T%v P%v - Actualizo contexto de ejecución", tid.Tid, tid.Pid)

	// TODO: Hablar con memoria

	return nil
}

func memoryRead(thread types.Thread, physicalDirection byte) ([4]byte, error) {
	logger.Info("T%v P%v - LEER -> %v", thread.Tid, thread.Pid, physicalDirection)

	// TODO: Charlar con memoria

	return [4]byte{0xde, 0xad, 0xbe, 0xef}, nil
}

func memoryWrite(thread types.Thread, physicalDirection byte, data [4]byte) error {
	logger.Info("T%v P%v - ESCRIBIR -> %v", thread.Tid, thread.Pid, physicalDirection)

	// TODO: Chamuyarse a la memoria

	return nil
}

func kernelYourProcessFinished(thread types.Thread, interruptReceived int) (err error) {
	logger.Debug("Kernel, tu proceso terminó! TID: %v, PID: %v", thread.Tid, thread.Pid)
	logger.Debug("Int. received %v", interruptReceived)

	return nil
}
