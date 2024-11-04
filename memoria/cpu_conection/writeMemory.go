package cpu_conection

import (
	"encoding/binary"
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/memoria/helpers"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaGlobals"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"net/http"
	"strconv"
	"time"
)

func WriteMemoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		logger.Error("Metodo no permitido")
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)

	query := r.URL.Query()
	dirS := query.Get("addr")
	tidS := query.Get("tid")
	pidS := query.Get("pid")

	// Log obligatorio
	logger.Info("## Escritura - (PID:TID) - (%v:%v) - Dir.Física: %v - Tamaño: %v", tidS, pidS, dirS, "")
	time.Sleep(time.Duration(memoriaGlobals.Config.ResponseDelay))

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
	var data uint32
	err = json.Unmarshal(body, &data)
	if err != nil {
		logger.Error("No se pudo decodificar el cuerpo del request")
		http.Error(w, "No se pudo decodificar el cuerpo del request", http.StatusBadRequest)
		return
	}

	var cuatromordidas []byte
	// Bit más significativo va a l principio del
	binary.BigEndian.PutUint32(cuatromordidas[:], data)

	err = helpers.WriteMemory(dir, cuatromordidas)
	if err != nil {
		logger.Error("Error al escribir en memoria de usuario")
		http.Error(w, "Error al escribir en memoria", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
