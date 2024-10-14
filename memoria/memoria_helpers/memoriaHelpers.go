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
	cuatro_mordidas := [4]byte{byte(123), byte(255), byte(111), byte(222)}
	return cuatro_mordidas, nil
}

func GetInstruction(thread types.Thread, pc int) (string, uint32, error) {
	// Verificar si el hilo tiene instrucciones
	instructions, exists := CodeRegionForThreads[thread]
	if !exists {
		return "", 0, fmt.Errorf("No se encontraron instrucciones para el hilo (PID:%d, TID:%d)", thread.PID, thread.TID)
	}

	// puede pasar esto? creeria que no. Lo dejo por las dudas
	if pc == -1 {
		pc = InstructionPointer[thread]
	}

	// Verificar si el PC está dentro de los límites de las instrucciones
	if pc >= len(instructions) {
		return "", 0, fmt.Errorf("No hay más instrucciones para el hilo (PID:%d, TID:%d)", thread.PID, thread.TID)
	}

	// Obtener la instrucción actual en la posición del PC
	instruction := instructions[pc]

	// Actualizar el PC para el siguiente ciclo
	newPC := pc + 1
	InstructionPointer[thread] = newPC // Guardar el nuevo PC en el mapa

	return instruction, uint32(newPC), nil
}

var InstructionPointer = make(map[types.Thread]int) //No estoy usando el map?? por que no? si me viene bien para buscar en el archivo... no entiendo
