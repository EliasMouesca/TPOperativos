package main

import (
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/utils/dino"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"os"
)

var config fsConfig

// TODO Chequear que no haya ya un archivo con el mismo nombre cuando mandan memdump
func init() {
	loggerLevel := "INFO"
	err := logger.ConfigureLogger("filesystem.log", loggerLevel)
	if err != nil {
		fmt.Println("No se pudo crear el logger - ", err)
	}

	data, err := os.ReadFile("config.json")
	if err != nil {
		logger.Error("No se pudo leer la config - %v", err)
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		logger.Error("Error parseando la config - %v", err)
	}

	err = logger.SetLevel(config.LogLevel)
	if err != nil {
		logger.Error("Error seteando el log level - %v", err)
	}

}

func main() {
	dino.Pterodactyl()

	err := initialize()
	if err != nil {
		logger.Fatal("EL filesystem no se pudo inicializar - %v", err)
	}
	//defer bitmapFile.Close()
	//defer bloquesFile.Close()

	logger.Info("--- Comienzo ejecución del filesystem ---")
	http.HandleFunc("/", notFound)
	http.HandleFunc("POST /filesystem/memoryDump", persistMemoryDump)

	self := fmt.Sprintf("%v:%v", config.SelfAddress, config.SelfPort)
	logger.Debug("Corriendo filesystem en %v", self)
	err = http.ListenAndServe(self, nil)
	if err != nil {
		logger.Fatal("ListenAndServe terminó con un error - %v", err)
	}
}

// Funciones init e initialize kajajsjasj y.. bueno
func initialize() error {
	var err error

	// Creamos el mount dir
	err = os.MkdirAll(config.MountDir+"/files", 0755)
	if err != nil {
		logger.Fatal("No se pudo crear el mountpoint - %v", err)
	}

	// Existe "bitmap.dat"?
	infoBitmap, errBitmap := os.Stat(config.MountDir + "/" + bitmapFilename)
	if errBitmap != nil {
		if !os.IsNotExist(errBitmap) {
			return errBitmap
		}
	}

	// Existe "bloques.dat"?
	infoBloques, errBloques := os.Stat(config.MountDir + "/" + bloquesFilename)
	if errBloques != nil {
		if !os.IsNotExist(errBloques) {
			return errBloques
		}
	}

	// Si alguno de los dos no existe, vamos de cero
	if errBitmap != nil || errBloques != nil {
		if errBitmap == nil {
			logger.Warn("El archivo '%s' existe, pero '%s' no; se crean ambos de cero.",
				bitmapFilename, bloquesFilename)
		}

		if errBloques == nil {
			logger.Warn("El archivo '%s' existe, pero '%s' no; se crean ambos de cero.",
				bloquesFilename, bitmapFilename)
		}

		// Creamos el archivo bitmap
		logger.Debug("Creando el archivo '%s'", bitmapFilename)
		bitmapFile, err := os.OpenFile(config.MountDir+"/"+bitmapFilename, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}

		// lo llenamos de ceros (redondeado para arriba)
		buffer := make([]byte, (config.BlockCount+7)/8)
		_, err = bitmapFile.Write(buffer)
		if err != nil {
			return err
		}

		// Hacemos seek al comienzo porque este archivo no se va a cerrar hasta que cerremos el programa
		_, err = bitmapFile.Seek(0, 0)
		if err != nil {
			return nil
		}

		// "bloques.dat"
		logger.Debug("Creando el archivo '%s'", bloquesFilename)
		bloquesFile, err := os.OpenFile(config.MountDir+"/"+bloquesFilename, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}

		defer bitmapFile.Close()
		defer bloquesFile.Close()

		buffer = make([]byte, config.BlockCount*config.BlockSize)
		_, err = bloquesFile.Write(buffer)
		if err != nil {
			return err
		}

		_, err = bloquesFile.Seek(0, 0)
		if err != nil {
			return nil
		}

		return nil
	}

	// Si se obtuvo correctamente la informacion de los dos archivos
	// Tienen el tamaño que esperamos ?
	if infoBitmap.Size() != int64((config.BlockCount+7)/8) ||
		infoBloques.Size() != int64(config.BlockCount*config.BlockSize) {
		logger.Fatal("La configuración no coincide con los archivos encontrados ('%s' y '%s')",
			bitmapFilename, bloquesFilename)
	}

	return nil

}

func notFound(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Request inválida %v, desde %v", r.RequestURI, r.RemoteAddr)
	w.WriteHeader(http.StatusBadRequest)
	_, err := w.Write([]byte("Request inválida"))
	if err != nil {
		logger.Error("No se pudo escribir la respuesta - %v", err.Error())
	}
}
