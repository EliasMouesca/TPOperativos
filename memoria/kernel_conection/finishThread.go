package kernel_conection

import (
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
)

func FinishThreadHandler(w http.ResponseWriter, r *http.Request) {
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

	// Extraer PID y TID del cuerpo JSON enviado por ThreadToExit
	pidS := requestData.PID
	tidS := requestData.TID

	//Según la consigna, se deben borran las estructuras administrativas del thread => hay que borrar el contexto
	/*
		pid, err:= strconv.Atoi(pidS)
		tid, err:= strconv.Atoi(tidS)
		if err != nil {
			logger.Error("Error al parsear el Thread (%v, %v) recibido - %v", pidS, tidS, err)
			http.Error(w, "Thread inválido", http.StatusBadRequest)
			return
		}
		delete(memoriaGlobals.ExecContext, types.Thread{types.Pid(pid), types.Tid(tid)})
		logger.Trace("Contexto eliminado");
	*/

	// Log obligatorio
	logger.Info("## Hilo Destruido - (PID:TID) - (%v,%v)", pidS, tidS)

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Hilo finalizado correctamente"))
	if err != nil {
		logger.Error("Error escribiendo response: %v", err)
	}
}
