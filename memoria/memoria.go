package main

import (
	"encoding/json"
	"fmt"
	cpu "github.com/sisoputnfrba/tp-golang/memoria/cpu_conection"
	fileSystem "github.com/sisoputnfrba/tp-golang/memoria/fileSystem_conection"
	kernel "github.com/sisoputnfrba/tp-golang/memoria/kernel_conection"
	"github.com/sisoputnfrba/tp-golang/memoria/memoria_helpers"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"os"
)

//TODO: Agregar el tiempo de espera para cada petición del CPU

func init() {
	loggerLevel := "INFO"
	err := logger.ConfigureLogger("memoria.log", loggerLevel)
	if err != nil {
		fmt.Println("No se pudo crear el logger - ", err)
		os.Exit(1)
	}
	logger.Debug("Logger creado")

	// Load Config
	configData, err := os.ReadFile("config/Config.json")
	if err != nil {
		logger.Fatal("No se pudo leer el archivo de configuración - %v", err.Error())
	}

	err = json.Unmarshal(configData, &memoria_helpers.Config)
	if err != nil {
		logger.Fatal("No se pudo parsear el archivo de configuración - %v", err.Error())
	}

	if err = memoria_helpers.Config.Validate(); err != nil {
		logger.Fatal("La configuración no es válida - %v", err.Error())
	}
	logger.Debug("Configuración cargada exitosamente")

	err = logger.SetLevel(memoria_helpers.Config.LogLevel)
	if err != nil {
		logger.Fatal("No se pudo leer el log-level - %v", err.Error())
	}

}

func main() {
	logger.Info("--- Comienzo ejecución MEMORIA ---")

	// TRUE RESPONSE
	http.HandleFunc("/", memoria_helpers.BadRequest)
	http.HandleFunc("/memoria/getContext", cpu.GetContext)
	http.HandleFunc("/memoria/saveContext", cpu.SaveContext)

	// STUB FORMAT RESPONSE
	http.HandleFunc("/memoria/getInstruction", cpu.GetInstruction)
	http.HandleFunc("/memoria/readMem", cpu.ReadMemory)
	http.HandleFunc("/memoria/writeMem", cpu.WriteMemory)

	http.HandleFunc("/memoria/createProcess", kernel.CreateProcess)
	http.HandleFunc("/memoria/finishProcess", kernel.FinishProcess)
	http.HandleFunc("/memoria/createThread", kernel.CreateThread)
	http.HandleFunc("/memoria/finishThread", kernel.FinishThread)

	http.HandleFunc("/memoria/memoryDump", fileSystem.DumpMemory)

	self := fmt.Sprintf("%v:%v", memoria_helpers.Config.SelfAddress, memoria_helpers.Config.SelfPort)
	logger.Info("Server activo en %v", self)
	err := http.ListenAndServe(self, nil)
	if err != nil {
		logger.Fatal("No se puede escuchar el puerto 8082: " + err.Error())
	}
}
