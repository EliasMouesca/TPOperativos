package fileSystem_conection

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/memoria/helpers"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaGlobals"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"time"
)

func DumpMemoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		logger.Error("Método no permitido")
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	logger.Debug("Request recibida de: %v", r.RemoteAddr)
	var requestData types.RequestToMemory
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		logger.Error("Error al decodificar el cuerpo de la solicitud: %v", err)
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}

	pid := requestData.Thread.PID
	tid := requestData.Thread.TID
	logger.Debug("Request de proceso: %v", pid)
	particion, err := memoriaGlobals.SistemaParticiones.ObtenerParticionDeProceso(pid)
	if err != nil {
		logger.Debug("No se ha encontrado la particion para el memory dump...")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	size := particion.Limite - particion.Base
	var contenidoMemProceso []byte
	for i := 0; i < size; i += 4 {
		cuatroMordidas, err := helpers.ReadMemory(particion.Base + i)
		if err != nil {
			logger.Error("Error al leer memoria: %v", err)
			return
		}
		contenidoMemProceso = append(contenidoMemProceso, cuatroMordidas...)
	}
	logger.Trace("Bytes leidos: %v", contenidoMemProceso)

	logger.Trace("Size: %v, len: %v", size, len(contenidoMemProceso))
	request := types.RequestToDumpMemory{
		Contenido: contenidoMemProceso,
		Nombre:    fmt.Sprintf("%d-%d-%s.dmp", pid, tid, time.Now().Format("2006-01-02T15:04:05")),
		Size:      size,
	}

	err = enviarAFileSystem(request)
	if err != nil {
		logger.Debug("No se ha enviado la request de memory dump a filesystem: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// Log obligatorio
	logger.Info("## Memory Dump solicitado - (PID:TID) - (%v:%v)", pid, tid)

	w.WriteHeader(http.StatusOK)
}

func enviarAFileSystem(request types.RequestToDumpMemory) error {
	// Convertir el request a JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return err
	}

	// Definir la URL de destino
	url := fmt.Sprintf("http://%v:%v/filesystem/memoryDump", memoriaGlobals.Config.IpFilesystem, memoriaGlobals.Config.PortFilesystem)

	// Enviar la solicitud POST
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	logger.Debug("DumpMemory enviado a FileSystem: %v", request)
	// Puedes verificar el código de estado de la respuesta si es necesario
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error en la solicitud: %v", resp.Status)
	}

	return nil
}
