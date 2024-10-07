package kernel_conection

import (
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
)

func FinishProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		logger.Error("Metodo no permitido")
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	query := r.URL.Query()
	//TODO: rami: El kernel no te pasa el size
	sizeS := query.Get("size")
	pidS := query.Get("ptd")

	//Log obligatorio
	logger.Info("## Proceso Destruido -  PID: %v - Tamaño: %v", pidS, sizeS)

	w.WriteHeader(http.StatusOK)
}

func createProcess(w http.ResponseWriter, r *http.Request) {
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
