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

//func GetInstructionPosta(thread types.Thread, pc int) (string, error) {
//	/*
//		instruction, exists := InstructionList[thread]
//		if !exists {
//			return "", errors.New("Instruction Not Found")
//		}
//		return instruction, nil
//	*/
//	return "SET AX 1", nil
//}

// TODOñ TEST YA QUE MEMORIA TODAVIA NO LEE ARCHIVOS. CREO UN SET DE INSTRUCCIONES Y SE LOS VOY MANDANDO A CPU A MEDIDA QUE LOS PIDE

func GetInstruction(thread types.Thread, pc int) (string, error) {
	// Verificar si el hilo tiene instrucciones
	instructions, exists := CodeRegionForThreads[thread]
	if !exists {
		return "", fmt.Errorf("No se encontraron instrucciones para el hilo (PID:%d, TID:%d)", thread.PID, thread.TID)
	}

	// Verificar si ya no hay más instrucciones
	ip := InstructionPointer[thread]
	if ip >= len(instructions) {
		return "", fmt.Errorf("No hay más instrucciones para el hilo (PID:%d, TID:%d)", thread.PID, thread.TID)
	}

	// Devolver la instrucción actual y avanzar el puntero
	instruction := instructions[ip]
	InstructionPointer[thread]++

	return instruction, nil
}

var InstructionPointer = make(map[types.Thread]int)

/*
func LoadTestInstructions() {
	// Instrucciones para el hilo1
	thread1 := types.Thread{PID: 0, TID: 0}
	[thread1] = []string{
		"SET AX 1",
		"PROCESS_EXIT",
		/*
			"SET BX 1",
			"SET PC 5",
			"SUM AX BX",
			"SUB AX BX",
			//"READ_MEM AX BX",
			//"WRITE_MEM AX BX",
			"JNZ AX 4",
			"LOG AX",
			"MUTEX_CREATE RECURSO_1",
			"MUTEX_LOCK RECURSO_1",
			"MUTEX_UNLOCK RECURSO_1",
			"DUMP_MEMORY",
			"IO 1500",
			"PROCESS_CREATE proceso1 256 1",
			"THREAD_CREATE hilo1 3",
			"THREAD_CANCEL 1",
			"THREAD_JOIN 1",
			"THREAD_EXIT",
			"PROCESS_EXIT",
}
*/
