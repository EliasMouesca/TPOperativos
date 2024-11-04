package helpers

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaGlobals"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
)

func BadRequest(w http.ResponseWriter, r *http.Request) {
	logger.Info("Request inválida: %v", r.RemoteAddr)
	w.WriteHeader(http.StatusBadRequest)
	_, err := w.Write([]byte("Request mal formada"))
	if err != nil {
		logger.Error("Error al escribir el response a %v", r.RemoteAddr)
	}
}

func WriteMemory(dir int, data [4]byte) error {
	var err error

	if ValidMemAddress(dir) {
		err = errors.New("No existe la dirección física solicitada")
		return err
	}

	for i := 0; i <= 3; i++ {
		data[i] = memoriaGlobals.UserMem[dir+i]
	}
	return nil
}

func ReadMemory(dir int) (uint32, error) {
	var err error
	if ValidMemAddress(dir) {
		err = errors.New("No existe la dirección física solicitada")
		return 0, err
	}

	var cuatroMordidas [4]byte

	for i := 0; i <= 3; i++ {
		cuatroMordidas[i] = memoriaGlobals.UserMem[dir+i]
	}

	data := binary.BigEndian.Uint32(cuatroMordidas[:])
	return data, nil
}

func GetInstruction(thread types.Thread, pc int) (instruction string, err error) {
	// Verificar si el hilo tiene instrucciones
	instructions, exists := memoriaGlobals.CodeRegionForThreads[thread]
	if !exists {
		logger.Error("Memoria no sabe que este thread exista ! (PID:%d, TID:%d)", thread.PID, thread.TID)
		return "", fmt.Errorf("no se encontraron instrucciones para el hilo (PID:%d, TID:%d)", thread.PID, thread.TID)
	}

	// Verificar si el PC está dentro de los límites de las instrucciones
	if pc > len(instructions) {
		logger.Error("Se pidió la instrucción número '%d' del proceso (PID:%d, TID:%d), la cual no existe",
			pc, thread.PID, thread.TID)
		return "", fmt.Errorf("no hay más instrucciones para el hilo (PID:%d, TID:%d)", thread.PID, thread.TID)
	}

	// Obtener la instrucción actual en la posición del PC
	instruction = instructions[pc]
	//TODO: esto puede romper todo, creemos que es (pc - 1)

	return instruction, nil
}

func ValidMemAddress(dir int) bool {
	if dir > 0 && dir+3 <= len(memoriaGlobals.UserMem) {
		return false
	}
	return true
}
