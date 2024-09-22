package main

import (
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"os"
)

var config MemoriaConfig
var execContext = make(map[types.Thread]types.ExecutionContext)

func init() {
	loggerLevel := "INFO"
	err := logger.ConfigureLogger("memoria.log", loggerLevel)
	if err != nil {
		fmt.Println("No se pudo crear el logger - ", err)
		os.Exit(1)
	}

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

	// --- INICIALIZAR EL SERVER ---
	http.HandleFunc("/", TerribleRequest)
	http.HandleFunc("/memoria/getContext", getContext)
	http.HandleFunc("/memoria/saveContext", saveContext)
	// TODO: -----------------
	http.HandleFunc("POST /memoria/updateExecutionContext", GoodRequest)
	http.HandleFunc("POST /memoria/getInstruction", GoodRequest)
	http.HandleFunc("POST /memoria/ReadMem", GoodRequest)
	http.HandleFunc("POST /WriteMem", GoodRequest)
	// -----------------------

	self := fmt.Sprintf("%v:%v", config.SelfAddress, config.SelfPort)
	logger.Info("Server activo en %v", self)
	err := http.ListenAndServe(self, nil)
	if err != nil {
		logger.Fatal("No se puede escuchar el puerto 8082: " + err.Error())
	}
	// --- FIN DE INICIALIZACION DE SERVER ---
}
func TerribleRequest(w http.ResponseWriter, r *http.Request) {
	logger.Info("Ke ac vo aca? Request inválida: %v", r.RemoteAddr)
	w.WriteHeader(http.StatusBadRequest)
	_, err := w.Write([]byte("Request mal formada"))
	if err != nil {
		logger.Error("Error al escribir el response a %v", r.RemoteAddr)
	}
}
