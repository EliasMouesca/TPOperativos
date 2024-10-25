package kernel_conection

import (
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaGlobals"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"strconv"
)

func CreateProcessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		logger.Error("Método no permitido")
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	// Leer el cuerpo de la request
	var requestData types.RequestToMemory
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		logger.Error("Error al decodificar el cuerpo de la solicitud: %v", err)
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	// Extraer el PID y el tamaño desde el cuerpo JSON
	pidS := requestData.Arguments[0]
	sizeS := requestData.Arguments[1]
	pid, _ := strconv.Atoi(pidS)
	size, _ := strconv.Atoi(sizeS)

	err = memoriaGlobals.SistemaParticiones.AsignarProcesoAParticion(types.Pid(pid), size)
	if err != nil {
		logger.Error("Error al asignar el proceso (<%v>) de tamanio %v a una particion de memoria", pidS, sizeS)
	}

	// Log obligatorio
	logger.Info("## Proceso Creado - FileName: %v - Tamaño: %v", pidS, sizeS)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Proceso creado correctamente"))
}
