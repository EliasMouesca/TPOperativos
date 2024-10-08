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

	//TODO: PARA INICIALIZAR EL KERNEL HAY QUE PONER EN CONSOL:
	// go run .\kernel.go .\planificadorLargoPlazo.go .\syscalls.go

	// Capturar los argumentos pasados al kernel por consola
	if len(os.Args) < 3 {
		logger.Fatal("Se requieren al menos dos argumentos: archivo de pseudocódigo y tamaño del proceso.")
	}
	fileName := os.Args[1]    // Primer argumento: nombre del archivo de pseudocódigo
	processSize := os.Args[2] // Segundo argumento: tamaño del proceso

	// Crear el primer proceso
	logger.Info("Creando el primer proceso inicial (archivo: %s, tamaño: %s)", fileName, processSize)
	initProcess(fileName, processSize)

	go planificadorLargoPlazo()
	//go planificadorCortoPlazo()

	// Listen and serve
	http.HandleFunc("/kernel/syscall", syscallRecieve)
	http.HandleFunc("/", badRequest)

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

	err = ExecuteSyscall(syscall) // map a la libreria de syscalls
	if err != nil {
		// Por alguna razón esto rompe cuando quiero compilar
		logger.Error("Error al ejecutar la syscall: %v - %v", syscalls.SyscallNames[syscall.Type], err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

}

func ExecuteSyscall(syscall syscalls.Syscall) error {
	syscallFunc, exists := syscallSet[syscall.Type]
	if !exists {
		return errors.New("la syscall pedida no es una syscall que el kernel entienda")
	}

	// Verificar si hay un thread en ejecución
	if kernelglobals.ExecStateThread != nil {
		logger.Info("## (<%v>:<%v>) - Solicitó syscall: <%v>",
			kernelglobals.ExecStateThread.FatherPCB.PID,
			kernelglobals.ExecStateThread.TID,
			syscalls.SyscallNames[syscall.Type],
		)
	} else {
		logger.Info("Syscall solicitada <%v>, pero no hay un thread en ejecución actualmente", syscalls.SyscallNames[syscall.Type])
	}

	kernelsync.WaitPlanificadorLP.Add(1)
	go func() {
		defer kernelsync.WaitPlanificadorLP.Done()
		err := syscallFunc(syscall.Arguments)
		if err != nil {
			logger.Error("La syscall devolvio un error - %v", err)
		}
	}()
	kernelsync.WaitPlanificadorLP.Wait()
	return nil
}

func initProcess(fileName, processSize string) {
	logger.Info("Inicializando el proceso inicial con archivo: %s, tamaño: %s", fileName, processSize)

	// Crear los argumentos para la syscall ProcessCreate
	args := []string{fileName, processSize, "0"}
	logger.Info("Archivo y tamanio guardados.")

	// Enviar la syscall ProcessCreate para que el kernel la maneje
	syscall := syscalls.Syscall{
		Type:      syscalls.ProcessCreate, // Tipo de syscall
		Arguments: args,                   // Argumentos del proceso (archivo y tamaño)
	}

	// Ejecutar la syscall
	err := ExecuteSyscall(syscall)
	if err != nil {
		logger.Fatal("Error al ejecutar la syscall ProcessCreate: %v", err)
	}

	logger.Info("Proceso inicial creado correctamente.")
}
