package helpers

import (
	"errors"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaGlobals"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
)

func BadRequest(w http.ResponseWriter, r *http.Request) {
	logger.Error("Request inválida: %v", r.RemoteAddr)
	w.WriteHeader(http.StatusBadRequest)
	_, err := w.Write([]byte("Request mal formada"))
	if err != nil {
		logger.Error("Error al escribir el response a %v", r.RemoteAddr)
	}
}

func WriteMemory(dir int, data []byte) error {
	var err error

	if !ValidMemAddress(dir) {
		err = errors.New("no existe la dirección física solicitada")
		return err
	}

	for i := 0; i <= 3; i++ {
		memoriaGlobals.UserMem[dir+i] = data[i]
	}
	return nil
}

func ReadMemory(dir int) ([]byte, error) {
	var err error
	if !ValidMemAddress(dir) {
		err = errors.New("no existe la dirección física solicitada")
		return nil, err
	}

	var cuatroMordidas = make([]byte, 4)

	// Esto se hace al revés, para poder comparar el funcionamiento del dump de fs con este issue 'https://github.com/sisoputnfrba/foro/issues/4463'
	for i := 0; i < 4; i++ {
		cuatroMordidas[3-i] = memoriaGlobals.UserMem[dir+i]
	}

	logger.Trace("cuatroMordidas: %v", cuatroMordidas)

	return cuatroMordidas, nil
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

	return instruction, nil
}

func ValidMemAddress(dir int) bool {
	logger.Trace("Es '%v' una dirección válida? -> %v", dir,
		!(dir < 0 || dir+3 >= len(memoriaGlobals.UserMem)))
	return !(dir < 0 || dir+3 >= len(memoriaGlobals.UserMem))
}
