package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"io"
	"os"
	"time"
)

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

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
// writeFile recibe el nombre, la data y crea el archivo
func writeFile(filename string, data []byte) error {
	size := len(data)
	logger.Debug("Cant max bloques: %v", config.BlockCount)
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

	logger.Trace("Persistiendo %v bloques: %v", len(bloquesDato), data)

	// Por cada bloque dato, guarda su índice en el bloque índice Y escribí la data en el bloque dato
	for i, bloqueDato := range bloquesDato {
		time.Sleep(time.Duration(config.BlockAccessDelay) * time.Millisecond)
		//time.Sleep(5 * time.Second)
		// -- Escribimos en el bloque índice --
		buffer := make([]byte, 4)
		bloquesFile, err := os.OpenFile(config.MountDir+"/"+bloquesFilename, os.O_RDWR, 0644)
		binary.LittleEndian.PutUint32(buffer, bloqueDato)
		bytesWritten, err := bloquesFile.WriteAt(buffer, int64(int(bloqueIndice)*config.BlockSize+4*i))
		if err != nil || bytesWritten != 4 {
			if err == nil {
				err = errors.New("no se escribieron 4 bytes")
			}
			logger.Fatal("Al menos un bloque no se pudo escribir - %v", err)
		}

		//-- Escribimos en el bloque dato --
		// Escribí en donde corresponda (bloqueDato * blockSize) un cacho de la data.

		bytesWritten, err = bloquesFile.WriteAt(
			// El cacho de bytes a escribir va desde el número de bloque (i) * el tamaño de bloque
			// hasta el siguiente bloque o el limite de la slice, lo que sea más chico.
			data[i*config.BlockSize:Min((i+1)*config.BlockSize, len(data))],
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

	logger.Info("## Archivo Creado: %s - Tamaño: %v", filename, len(data))

	return nil
}

// allocateBlocks recibe cuantos bloques querés y te devuelve una lista con los que te asginó o error.
// Esta función hace tanto el chequeo como la asignación.
func allocateBlocks(numberOfBlocksToAllocate int) ([]uint32, error) {
	logger.Trace("Asignando %v bloques", numberOfBlocksToAllocate)
	bitmapMutex.Lock()
	defer bitmapMutex.Unlock()
	var err error
	bitmapFile, err := os.OpenFile(config.MountDir+"/"+bitmapFilename, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer bitmapFile.Close()
	// Hacemos una copai del bitmap en memoria, después hay que persistirla si queremos guardar los cambios
	bitmap, err := io.ReadAll(bitmapFile)
	if err != nil {
		return nil, err
	}

	selectedBlocks := make([]uint32, 0, numberOfBlocksToAllocate)

	// Por cada byte en el bitmap...
	for byteInBitmap := 0; byteInBitmap < len(bitmap); byteInBitmap++ {
		//logger.Trace("Probando byte: %v", byteInBitmap)
		// Por cada bit en un byte...
		for bit := 0; bit < 8; bit++ {
			//logger.Trace("Probando bit: %v", bit)
			// Si el bit está seteado en 0
			if (bitmap[byteInBitmap] & (1 << bit)) == 0 {
				logger.Debug("Bloque seleccionado: %v", byteInBitmap*8+bit)
				// Ponelo en 1
				bitmap[byteInBitmap] |= 1 << bit
				// y agregalo a la lista de los seleccionados
				selectedBlocks = append(selectedBlocks, uint32(byteInBitmap*8+bit))
			}
			// Si ya tenemos los que necesitamos, salí
			if len(selectedBlocks) >= numberOfBlocksToAllocate {
				goto FoundNecessaryBlocks
			}
		}
	}

	// Si llego hasta acá y no saltó a "OutsideTheForLoops" es porque leyó el bitmap
	// y no encontró suficientes bloques libres. Simplemente nos vamos sin guardar y chau.
	logger.Warn("No hay espacio suficiente para guardar %v bloques", numberOfBlocksToAllocate)
	return nil, errors.New("no hay espacio suficiente")

FoundNecessaryBlocks:
	// Por cada bloque elegido, actualizá el archivo del bitmap
	for _, block := range selectedBlocks {
		// ie. El bloque es el 38 (= 4 * 8 + 6), entonces el byte en el que esta guardado es el 4
		blockByte := block / 8
		// Actualizamos ese byte
		_, err := bitmapFile.WriteAt([]byte{bitmap[blockByte]}, int64(blockByte))

		// Si algo falló rompé t0do pq el sistema quedó en un estado inconsistente
		if err != nil {
			// por qué es fatal? -> porque puede salir bien la primera y fallar la segunda => estado inconsistente => xd
			logger.Fatal("No se pudieron persistir las asignaciones de bloques")
		}

		// TODO: A este log le falta información que no logeamos pq no la tenemos en este punto! Nombre del archivo y cantidad de bloques libres
		logger.Info("## Bloque asignado: %v", block)
	}

	logger.Debug("MemoryDump completado")

	return selectedBlocks, nil
}
