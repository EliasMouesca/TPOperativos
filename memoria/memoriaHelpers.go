package main

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

func writeMemoryPosta(dir int, data [4]byte) error {
	return nil
}

func readMemoryPosta(dir int) ([4]byte, error) {
	var mordida [4]byte = [4]byte{byte(123), byte(255), byte(111), byte(222)}
	return mordida, nil
}

func getInstructionPosta(tread types.Thread, pc int) (string, error) {
	return "SET AX 1", nil
}
