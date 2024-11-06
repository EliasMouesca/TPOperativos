package main

import (
	"encoding/json"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
)

type MemoryDumpRequest struct {
	Nombre    string `json:"Nombre"`
	Size      int    `json:"Size"`
	Contenido []byte `json:"Contenido"`
}

func persistMemoryDump(w http.ResponseWriter, r *http.Request) {
	logger.Debug("MemoryDump solicitado")
	var dumpRequest MemoryDumpRequest

	err := json.NewDecoder(r.Body).Decode(&dumpRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error("Error al decodificar el cuerpo de la solicitud - %v", err)
		w.Write([]byte("Error al decodificar JSON"))
		return
	}

	if len(dumpRequest.Contenido) != dumpRequest.Size {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error("Se solicitó una dump request que tiene un tamaño distinto al del slice enviado")
		w.Write([]byte("El tamaño del slice y el tamaño envíado en la request no coinciden!"))
		return
	}

	err = writeFile(dumpRequest.Nombre, dumpRequest.Contenido)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error("No se pudo escribir el archivo - %v", err)
		w.Write([]byte("No se pudo escribir el archivo - " + err.Error()))
		return
	}

	// Responder con éxito
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Todo bien, dump persistido :)"))

	logger.Info("## Fin de solicitud - Archivo: %v", dumpRequest.Nombre)
	return

}
