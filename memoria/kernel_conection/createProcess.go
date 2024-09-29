package kernel_conection

import (
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
)

func CreateProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		logger.Error("Metodo no permitido")
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	query := r.URL.Query()
	sizeS := query.Get("size")
	pidS := query.Get("ptd")

	//Log obligatorio
	logger.Info("## Proceso Creado -  PID: %v - Tamaño: %v", pidS, sizeS)

	w.WriteHeader(http.StatusOK)
}
