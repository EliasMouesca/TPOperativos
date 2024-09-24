package main

import (
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"os"
)

//TODO: Agregar el tiempo de espera para cada petición del CPU

// GOBALES DE LA MEMORIA
var config MemoriaConfig
var execContext = make(map[types.Thread]types.ExecutionContext)
var indexInstructionsLists = make(map[types.Thread][]string)
var userMem = make([]byte, config.MemorySize)

// ------------------------------------------------------------------------

func init() {
	loggerLevel := "INFO"
	err := logger.ConfigureLogger("memoria.log", loggerLevel)
	if err != nil {
		fmt.Println("No se pudo crear el logger - ", err)
		os.Exit(1)
	}
	logger.Debug("Logger creado")

	// Load config
	configData, err := os.ReadFile("config.json")
	if err != nil {
		logger.Fatal("No se pudo leer el archivo de configuración - %v", err.Error())
	}

	err = json.Unmarshal(configData, &config)
	if err != nil {
		logger.Fatal("No se pudo parsear el archivo de configuración - %v", err.Error())
	}

	if err = config.validate(); err != nil {
		logger.Fatal("La configuración no es válida - %v", err.Error())
	}
	logger.Debug("Configuración cargada exitosamente")

	err = logger.SetLevel(config.LogLevel)
	if err != nil {
		logger.Fatal("No se pudo leer el log-level - %v", err.Error())
	}

}

func main() {
	logger.Info("--- Comienzo ejecución MEMORIA ---")

	// TRUE RESPONSE
	http.HandleFunc("/", BadRequest)
	http.HandleFunc("/memoria/getContext", getContext)
	http.HandleFunc("/memoria/saveContext", saveContext)
	// STUB FORMAT RESPONSE
	http.HandleFunc("/memoria/getInstruction", getInstruction)
	http.HandleFunc("/memoria/readMem", readMemory)
	http.HandleFunc("/memoria/writeMem", writeMemory)
	http.HandleFunc("/memoria/createProcess", createProcess)
	http.HandleFunc("/memoria/finishProcess", finishProcess)
	http.HandleFunc("/memoria/createThread", createThread)
	http.HandleFunc("/memoria/finishThread", finishThread)
	http.HandleFunc("/memoria/dump", dump)
	// TODO: -----------------
	// -----------------------

	self := fmt.Sprintf("%v:%v", config.SelfAddress, config.SelfPort)
	logger.Info("Server activo en %v", self)
	err := http.ListenAndServe(self, nil)
	if err != nil {
		logger.Fatal("No se puede escuchar el puerto 8082: " + err.Error())
	}
}
