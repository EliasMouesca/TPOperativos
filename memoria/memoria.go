package main

import (
	"encoding/json"
	"fmt"
	cpu "github.com/sisoputnfrba/tp-golang/memoria/cpu_conection"
	"github.com/sisoputnfrba/tp-golang/memoria/esquemas_particiones/dinamicas"
	"github.com/sisoputnfrba/tp-golang/memoria/esquemas_particiones/fijas"
	"github.com/sisoputnfrba/tp-golang/memoria/estrategias_asignacion/best"
	"github.com/sisoputnfrba/tp-golang/memoria/estrategias_asignacion/first"
	"github.com/sisoputnfrba/tp-golang/memoria/estrategias_asignacion/worst"
	fileSystem "github.com/sisoputnfrba/tp-golang/memoria/fileSystem_conection"
	"github.com/sisoputnfrba/tp-golang/memoria/helpers"
	kernel "github.com/sisoputnfrba/tp-golang/memoria/kernel_conection"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaGlobals"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaTypes"
	"github.com/sisoputnfrba/tp-golang/utils/dino"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"os"
)

var EstrategiaAsignacionMap = map[string]memoriaTypes.EstrategiasAsignacionInterface{
	"FIRST": &first.First{},
	"BEST":  &best.Best{},
	"WORST": &worst.Worst{},
}
var EstrategiaParticionesMap = map[string]memoriaTypes.ParticionesInterface{
	"FIJAS":     &fijas.Fijas{},
	"DINAMICAS": &dinamicas.Dinamicas{},
}

func init() {
	dino.Triceraptops()
	loggerLevel := "TRACE"
	err := logger.ConfigureLogger("memoria.log", loggerLevel)
	if err != nil {
		fmt.Println("No se pudo crear el logger - ", err)
		os.Exit(1)
	}
	logger.Debug("Logger creado")

	// Load Config
	configData, err := os.ReadFile("config.json")
	if err != nil {
		logger.Fatal("No se pudo leer el archivo de configuración - %v", err.Error())
	}

	err = json.Unmarshal(configData, &memoriaGlobals.Config)
	if err != nil {
		logger.Fatal("No se pudo parsear el archivo de configuración - %v", err.Error())
	}

	//	if err = memoria_helpers.Config.Validate(); err != nil {
	//		logger.Fatal("La configuración no es válida - %v", err.Error())
	//	}
	logger.Debug("Configuración cargada exitosamente")

	err = logger.SetLevel(memoriaGlobals.Config.LogLevel)
	if err != nil {
		logger.Fatal("No se pudo leer el log-level - %v", err.Error())
	}

	memoriaGlobals.EstrategiaAsignacion = EstrategiaAsignacionMap[memoriaGlobals.Config.SearchAlgorithm]
	memoriaGlobals.SistemaParticiones = EstrategiaParticionesMap[memoriaGlobals.Config.Scheme]
	memoriaGlobals.SistemaParticiones.Init()
}

func main() {
	logger.Info("--- Comienzo ejecución MEMORIA ---")
	memoriaGlobals.UserMem = make([]byte, memoriaGlobals.Config.MemorySize)

	http.HandleFunc("/memoria/getContext", cpu.GetContextHandler)
	http.HandleFunc("/memoria/saveContext", cpu.SaveContextHandler)
	http.HandleFunc("/memoria/getInstruction", cpu.GetInstructionHandler)
	http.HandleFunc("/memoria/readMem", cpu.ReadMemoryHandler)
	http.HandleFunc("/memoria/writeMem", cpu.WriteMemoryHandler)

	http.HandleFunc("/memoria/createProcess", kernel.CreateProcessHandler)
	http.HandleFunc("/memoria/finishProcess", kernel.FinishProcessHandler)
	http.HandleFunc("/memoria/createThread", kernel.CreateThreadHandler)
	http.HandleFunc("/memoria/finishThread", kernel.FinishThreadHandler)
	http.HandleFunc("/memoria/compactar", kernel.CompactarHandler)

	http.HandleFunc("/memoria/memoryDump", fileSystem.DumpMemoryHandler)
	http.HandleFunc("/", helpers.BadRequest)

	self := fmt.Sprintf("%v:%v", memoriaGlobals.Config.SelfAddress, memoriaGlobals.Config.SelfPort)
	logger.Info("Server activo en %v", self)
	err := http.ListenAndServe(self, nil)
	if err != nil {
		logger.Fatal("No se puede escuchar el puerto 8082: " + err.Error())
	}
}
