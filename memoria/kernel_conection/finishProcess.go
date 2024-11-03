package kernel_conection

import (
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaGlobals"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
)

func FinishProcessHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug("********Entra a finish process handler")
	// Verificar que sea un POST en lugar de GET
	if r.Method != "POST" {
		logger.Error("Método no permitido")
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request de finishProcess recibida de: %v", r.RemoteAddr)

	// Leer el cuerpo de la solicitud (debe contener un JSON con la información del proceso)
	var requestData types.RequestToMemory
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		logger.Error("Error al decodificar el cuerpo de la solicitud: %v", err)
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	// Extraer PID del cuerpo JSON enviado por ProcessToExit
	pidS := requestData.Thread.PID
	var errLiberar error
	errLiberar = memoriaGlobals.SistemaParticiones.LiberarParticion(pidS)
	if errLiberar != nil {
		logger.Error("Error al liberar particion: %v ", errLiberar.Error())
		http.Error(w, "No se pudo liberar la particion", http.StatusInternalServerError)
		return
	}

	// Log obligatorio
	logger.Info("## Proceso Destruido - PID: %v", pidS)
	w.WriteHeader(http.StatusOK)

	//_, err = w.Write([]byte("Proceso destruido correctamente"))
	//if err != nil {
	//	logger.Error("Error escribiendo response: %v", err)
	//}
	logger.Debug("******* FIN de Finish process handler")
}
