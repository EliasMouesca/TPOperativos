package memoria_helpers

import (
	"fmt"
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
	return nil
}

func ReadMemory(dir int) ([4]byte, error) {
	// cuatro mordidas, ja ja    ~eli
	cuatroMordidas := [4]byte{byte(123), byte(255), byte(111), byte(222)}
	return cuatroMordidas, nil
}

func GetInstruction(thread types.Thread, pc int) (instruction string, err error) {
	// Verificar si el hilo tiene instrucciones
	instructions, exists := CodeRegionForThreads[thread]
	if !exists {
		logger.Error("Memoria no sabe que este thread exista ! (PID:%d, TID:%d)", thread.PID, thread.TID)
		return "", fmt.Errorf("no se encontraron instrucciones para el hilo (PID:%d, TID:%d)", thread.PID, thread.TID)
	}

	// puede pasar esto? creeria que no. Lo dejo por las dudas
	// Esto no puede pasar
	/*if pc == -1 {
		pc = InstructionPointer[thread]
	}*/

	// Verificar si el PC está dentro de los límites de las instrucciones
	if pc > len(instructions) {
		logger.Error("Se pidió la instrucción número '%d' del proceso (PID:%d, TID:%d), la cual no existe",
			pc, thread.PID, thread.TID)
		return "", fmt.Errorf("no hay más instrucciones para el hilo (PID:%d, TID:%d)", thread.PID, thread.TID)
	}

	// Obtener la instrucción actual en la posición del PC
	instruction = instructions[pc]

	// Actualizar el PC para el siguiente ciclo
	//newPC := pc + 1     // Lo hace CPU !
	//InstructionPointer[thread] = newPC

	return instruction, nil
}
