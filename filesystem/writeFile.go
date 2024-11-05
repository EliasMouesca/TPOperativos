package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
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

// TODO: Chequear que no haya ya un archivo con el mismo nombre
func writeFile(filename string, data []byte) error {
	size := len(data)

	// Si lo que quieren guardar es más grande que usar todos los bloques del disco menos uno para indexar => ?
	if size > config.BlockSize*(config.BlockCount-1) {
		return errors.New("archivo demasiado grande")
	}

	// Si lo que quieren guardar es más grande que usar todos los bloques que podemos direccionar con un bloque => impostor among us
	if size > (config.BlockSize/4)*config.BlockSize {
		return errors.New("archivo demasiado grande")
	}

	// Chequear espacio disponible y reservarlo
	bloques, err := allocateBlocks((size+config.BlockSize-1)/config.BlockSize + 1)
	if err != nil {
		return err
	}

	bloqueIndice := bloques[0]
	bloquesDato := bloques[1:]

	logger.Trace("Persistiendo %v bytes", len(bloquesDato))

	// Guardamos los indices de los bloquesDato en el bloqueIndice
	for i, bloqueDato := range bloquesDato {
		buffer := make([]byte, 4)
		binary.LittleEndian.PutUint32(buffer, bloqueDato)
		bytesWritten, err := bloquesFile.WriteAt(buffer, int64(int(bloqueIndice)*config.BlockSize+4*i))
		if err != nil || bytesWritten != 4 {
			if err == nil {
				err = errors.New("no se escribieron 4 bytes")
			}
			logger.Fatal("Al menos un bloque no se pudo escribir - %v", err)
		}

		// Escribí en donde corresponda (bloqueDato * blockSize) un cacho de la data.
		// De que tamaño el cacho? -> del tamaño de un bloque o de lo que te quede por escribir, lo que sea más chico
		bytesWritten, err = bloquesFile.WriteAt(
			data[i*config.BlockSize:min((i+1)*config.BlockSize, len(data))],
			int64(int(bloqueDato)*config.BlockSize))
		if err != nil {
			logger.Fatal("Al menos un bloque no se pudo escribir - %v", err)
		}
	}

	// Si anduvo bien, creamos la metadata
	var fcb = FCB{bloqueIndice, len(bloquesDato)}
	file, err := os.Create(config.MountDir + "/files/" + filename)
	defer file.Close()
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	if err = encoder.Encode(fcb); err != nil {
		return err
	}

	return nil
}

func allocateBlocks(numberOfBlocksToAllocate int) ([]uint32, error) {
	logger.Trace("Asignando %v bloques", numberOfBlocksToAllocate)
	bitmapMutex.Lock()
	defer bitmapMutex.Unlock()

	// Leemos el bitmap, después hay que persistirlo si lo cambiamos
	bitmap, err := io.ReadAll(bitmapFile)
	if err != nil {
		return nil, err
	}

	selectedBlocks := make([]uint32, 0, numberOfBlocksToAllocate)

	for byteInBitmap := 0; byteInBitmap < len(bitmap); byteInBitmap++ {
		logger.Trace("Probando byte: %v", byteInBitmap)
		for bit := 0; bit < 8; bit++ {
			logger.Trace("Probando bit: %v", bit)
			if (bitmap[byteInBitmap] & (1 << bit)) == 0 {
				logger.Trace("Bloque seleccionado: %v", byteInBitmap*8+bit)
				bitmap[byteInBitmap] |= 1 << bit
				selectedBlocks = append(selectedBlocks, uint32(byteInBitmap*8+bit))
			}

			if len(selectedBlocks) >= numberOfBlocksToAllocate {
				goto OutsideTheForLoops
			}
		}
	}

	// Si llego hasta acá y no saltó a "OutsideTheForLoops" es porque leyó el bitmap
	// y no encontró suficientes bloques libres. Simplemente nos vamos sin persistir y chau.
	return nil, errors.New("no hay espacio suficiente")

OutsideTheForLoops:
	// Esto escribe varias veces la misma posición si es que dos bloques fueron asignados en el mismo byte, pero bueno
	for _, block := range selectedBlocks {
		// ie. El bloque es el 38 (= 4 * 8 + 6), entonces el byte en el que esta guardado es el 4
		blockByte := block / 8
		// Actualizamos ese byte
		_, err := bitmapFile.WriteAt([]byte{bitmap[blockByte]}, int64(blockByte))
		if err != nil {
			// por qué es fatal? -> porque puede salir bien la primera y fallar la segunda => estado inconsistente => xd
			logger.Fatal("No se pudieron persistir las asignaciones de bloques")
		}
	}
	return selectedBlocks, nil
}
