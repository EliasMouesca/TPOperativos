package fileSystem_conection

import (
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
)

// TODO: Esto le tiene que pedir a filesystem que... vos sabés, haga el dump
func DumpMemoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		logger.Error("Método no permitido")
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	// Leer el cuerpo de la solicitud (debe contener un JSON con la información del hilo)
	var requestData types.RequestToMemory
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		logger.Error("Error al decodificar el cuerpo de la solicitud: %v", err)
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	// Extraer PID y TID del cuerpo JSON enviado
	pid := requestData.Arguments[0]
	tid := requestData.Arguments[1]

	// Log obligatorio
	logger.Info("## Memory Dump solicitado - (PID:TID) - (%v:%v)", pid, tid)

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Memory dump solicitado correctamente"))
	if err != nil {
		logger.Error("Error escribiendo response: %v", err)
	}
}
