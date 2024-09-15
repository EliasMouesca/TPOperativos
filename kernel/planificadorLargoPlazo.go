package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
)

// nose todavia bien como vamos a hacer el planificador
// si poner estas funciones dentro de las syscalls
// o crear un hilo en Kernel.go que quede corriendo
// o varios hilos en kernel.go
// o un map? no creo q sea buena opcion
// depende de como hagamos la sincronizacion

var available int = 0

func processToReady(processSize int, prioridad int) {
	for true {
		availableMemory(processSize) // seguro hay qui enviarle el psudoCodigo a memoria
		if !NEW.IsEmpty() && available == 1 {
			pcb, err := NEW.GetAndRemoveNext()
			if err != nil {
				logger.Error("Error en la cola NEW - %v", err)
			}
			_ = TCB{
				TID:       0,
				Prioridad: prioridad,
				ConectPCB: pcb,
			}
			available = 0 // reiniciar available
			// algorithm.AddToReady(hiloMain) preguntar a eli o juan como hicieron al final la cola Ready
		}
	}
}

func processToExit(pcb *PCB) {
	// aca hay que hacer sincronizacion
	// por que hay q informar a memoria
	// y despues volver al flujo de PROCESS_EXIT
	// y despues volver al flujo de esta funcion con processToReady
	// vamos por buen camino processToReady se tiene que inicializar por el enunciado
	// hay que ver el tema de sincroo!!
}

// esto hay que mejorarlo seguro quiza hacerlo de alguna manera
// polimorfica, ya que lo unico que hace  basicamente
// largo plazo es comunicarse con memoria
func availableMemory(processSize int) {

	logger.Debug("Preguntando a memoria si tiene espacio disponible. ")

	// Serializar mensaje
	processSize_json, err := json.Marshal(processSize)
	if err != nil {
		logger.Fatal("Error al serializar processSize - %v", err)
		return
	}

	// Hacer request a memoria
	memoria := &http.Client{}
	url := fmt.Sprintf("http://%s:%d/memoria/availableMemory", Config.MemoryAddress, Config.MemoryPort)
	logger.Debug("Enviando request a memoria")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(processSize_json))
	if err != nil {
		logger.Fatal("Error al conectar con memoria - %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Recibo repuesta de memoria
	resp, err := memoria.Do(req)
	if err != nil {
		logger.Fatal("Error al obtener mensaje de respuesta por parte de memoria - %v", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		available = 1
	} else {
		logger.Info("No hay espacio disponible en memoria")
	}
}
