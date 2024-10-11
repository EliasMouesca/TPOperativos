package kernel_conection

import (
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/memoria/memoria_helpers"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
)

func CreateThread(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		logger.Error("Método no permitido")
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	// Leer el cuerpo de la solicitud (debe contener un JSON con la información del hilo)
	var requestData types.Thread
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		logger.Error("Error al decodificar el cuerpo de la solicitud: %v", err)
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	// Extraer PID y TID del cuerpo JSON enviado por NewThreadToReady
	pidS := requestData.PID
	tidS := requestData.TID

	// Log obligatorio
	logger.Info("## Hilo Creado - (PID:TID) - (%v,%v)", pidS, tidS)

	// Guardar el contexto en ExecContext, usando PID y TID como clave
	context := types.ExecutionContext{}
	thread := types.Thread{PID: pidS, TID: tidS}
	memoria_helpers.ExecContext[thread] = context
	logger.Info("Contexto creado para el hilo - (PID:TID): (%v, %v)", pidS, tidS)

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Hilo creado correctamente"))
	if err != nil {
		logger.Error("Error escribiendo response: %v", err)
	}
}
