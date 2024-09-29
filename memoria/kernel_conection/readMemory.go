package kernel_conection

import (
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/memoria/memoria_helpers"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"strconv"
)

func ReadMemory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		logger.Error("Metodo no permitido")
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	query := r.URL.Query()
	dirS := query.Get("dir")
	tidS := query.Get("tid")
	pidS := query.Get("pid")

	// Que es el Tamaño????????
	// Log obligatorio
	logger.Info("## Lectura - (PID:TID) - (%v:%v) - Dir.Física: %v - Tamaño: %v", tidS, pidS, dirS)

	dir, err := strconv.Atoi(dirS)
	if err != nil {
		logger.Error("Dirección física mal formada: %v", dirS)
		http.Error(w, "Dirección física mal formada", http.StatusNotFound)
	}

	data, err := memoria_helpers.ReadMemoryPosta(dir)
	if err != nil {
		logger.Error("Error al leer la dirección: %v", dir)
		http.Error(w, "No se pudo leer la dirección de memoria", http.StatusNotFound)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Error("Error al escribir el response - %v", err.Error())
		memoria_helpers.BadRequest(w, r)
		return
	}
}
