package main

import (
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"os"
)

func writeFilePhysically(filename string, data []byte) error {
	logger.Trace("Se está creando el archivo '%s' físicamente en la computadora, fopen(), fwrite()...", filename)
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func writeFile(filename string, data []byte) error {
	// 1
	checkEspacio(len(data))

	// 2
	reservarBloques

	// 3
	crearMetadata()

	// 4
	writeIndexBlock()

	// 5
	writeDataBlocks()

	return nil
}
