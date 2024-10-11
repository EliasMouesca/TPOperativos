package memoria_helpers

import (
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

func WriteMemoryPosta(dir int, data [4]byte) error {
	return nil
}

func ReadMemoryPosta(dir int) ([4]byte, error) {
	cuatro_mordidas := [4]byte{byte(123), byte(255), byte(111), byte(222)}
	return cuatro_mordidas, nil
}

func GetInstructionPosta(thread types.Thread, pc int) (string, error) {
	/*
		instruction, exists := InstructionList[thread]
		if !exists {
			return "", errors.New("Instruction Not Found")
		}
		return instruction, nil
	*/
	return "PROCESS_EXIT", nil
}
