package fileSystem_conection

import (
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
)

func Dump(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		logger.Error("Metodo no permitido")
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	query := r.URL.Query()
	pidS := query.Get("pid")
	tidS := query.Get("tid")

	// Log obligatorio
	logger.Info("## Memory Dump solicitado - (PID:TID) - (%v:%v)", pidS, tidS)

	w.WriteHeader(http.StatusOK)
}
