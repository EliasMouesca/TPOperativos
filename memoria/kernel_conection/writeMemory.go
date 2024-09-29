package kernel_conection

import (
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/memoria/memoria_helpers"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"net/http"
	"strconv"
)

func WriteMemory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		logger.Error("Metodo no permitido")
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	query := r.URL.Query()
	dirS := query.Get("dir")
	tidS := query.Get("tid")
	pidS := query.Get("pid")
	// Log obligatorio
	logger.Info("## Escritura - (PID:TID) - (%v:%v) - Dir.Física: %v - Tamaño: %v", tidS, pidS, dirS)

	dir, err := strconv.Atoi(dirS)
	if err != nil {
		logger.Error("Dirección física mal formada: %v", dirS)
		http.Error(w, "Dirección física mal formada", http.StatusNotFound)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("No se pudo leer el cuerpo del request")
		http.Error(w, "Dirección física mal formada", http.StatusNotFound)
		return
	}
	defer r.Body.Close()

	// Decode JSON from body
	var data [4]byte
	err = json.Unmarshal(body, &data)
	if err != nil {
		logger.Error("No se pudo decodificar el cuerpo del request")
		memoria_helpers.BadRequest(w, r)
		return
	}

	err = memoria_helpers.WriteMemoryPosta(dir, data)
	if err != nil {
		logger.Error("Error al escribir en memoria de usuario")
		memoria_helpers.BadRequest(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
}
