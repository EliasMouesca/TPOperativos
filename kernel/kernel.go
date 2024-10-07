package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelsync"
	"github.com/sisoputnfrba/tp-golang/types/syscalls"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"os"
)

func init() {
	// Init logger
	err := logger.ConfigureLogger("kernel.log", "INFO")
	if err != nil {
		fmt.Println("No se pudo crear el logger -", err.Error())
		os.Exit(1)
	}
	logger.Debug("Logger creado")

	// Load Config
	configData, err := os.ReadFile("config.json")
	if err != nil {
		logger.Fatal("No se pudo leer el archivo de configuración - %v", err.Error())
	}

	err = json.Unmarshal(configData, &kernelglobals.Config)
	if err != nil {
		logger.Fatal("No se pudo parsear el archivo de configuración - %v", err.Error())
	}

	// Agregar logs para verificar los valores cargados
	logger.Info("Configuración cargada: SelfPort=%d, MemoryPort=%d, CpuPort=%d",
		kernelglobals.Config.SelfPort,
		kernelglobals.Config.MemoryPort,
		kernelglobals.Config.CpuPort)

	if err = kernelglobals.Config.Validate(); err != nil {
		logger.Fatal("La configuración no es válida - %v", err.Error())
	}

	err = logger.SetLevel(kernelglobals.Config.LogLevel)
	if err != nil {
		logger.Fatal("No se pudo leer el log-level - %v", err.Error())
	}

	logger.Info("Configuración cargada exitosamente")
}

func main() {
	logger.Info("-- Comenzó la ejecución del kernel --")

	go planificadorLargoPlazo()
	//go planificadorCortoPlazo()

	// Implementacion GOD
	// Inicializar un proceso sin que CPU mande nada
	// PROCESS_CREATE(// fileName, processSize, TID)

	// Listen and serve
	http.HandleFunc("/kernel/syscall", syscallRecieve)
	http.HandleFunc("/", badRequest)
	http.HandleFunc("POST kernel/process/finished", processFinish)

	url := fmt.Sprintf("%s:%d", kernelglobals.Config.SelfAddress, kernelglobals.Config.SelfPort)
	logger.Info("Server activo en %s", url)
	err := http.ListenAndServe(url, nil)
	if err != nil {
		logger.Fatal("ListenAndServe retornó error - %v", err)
	}

}

func badRequest(w http.ResponseWriter, r *http.Request) {
	logger.Info("Request inválida: %v", r.RequestURI)
	w.WriteHeader(http.StatusBadRequest)
	_, err := w.Write([]byte("Bad request!"))
	if err != nil {
		logger.Error("Error escribiendo response - %v", err)
	}
}

func syscallRecieve(w http.ResponseWriter, r *http.Request) {

	var syscall syscalls.Syscall
	// Parsear la syscall request
	err := json.NewDecoder(r.Body).Decode(&syscall)
	if err != nil {
		logger.Error("Error al decodificar el cuerpo de la solicitud - %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	kernelsync.WaitPlanificadorLP.Add(1)
	go func() {
		defer kernelsync.WaitPlanificadorLP.Done()
		err = ExecuteSyscall(syscall) // map a la libreria de syscalls
		if err != nil {
			// Por alguna razón esto rompe cuando quiero compilar
			logger.Error("Error al ejecutar la syscall: %v - %v", syscalls.SyscallNames[syscall.Type], err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}()
}

func processFinish(w http.ResponseWriter, r *http.Request) {
	// Cosa de largo plazo :)
	// TODO: Volver a poner el hilo que vino de CPU en la cola ready
	// Rami: What?

	//TODO: ESTO NO VA, YA ESTA EN EL PLANI DE LARGO PLAZO - tobi

}

func ExecuteSyscall(syscall syscalls.Syscall) error {
	syscallFunc, exists := syscallSet[syscall.Type]
	if !exists {
		return errors.New("la syscall pedida no es una syscall que el kernel entienda")
	}

	logger.Info("## (<%v>:<%v>) - Solicitó syscall: <%v>",
		kernelglobals.ExecStateThread.FatherPCB.PID,
		kernelglobals.ExecStateThread.TID,
		syscalls.SyscallNames[syscall.Type],
	)

	err := syscallFunc(syscall.Arguments)
	if err != nil {
		return err
	}

	return nil
}
