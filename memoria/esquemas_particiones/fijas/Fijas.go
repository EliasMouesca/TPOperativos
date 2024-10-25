package fijas

import (
	"errors"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaGlobals"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaTypes"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

type Fijas struct {
	Particiones []memoriaTypes.Particion
}

func (f *Fijas) init() {
	logger.Debug("Inicializando particiones fijas")
	base := 0
	for _, tamanio := range memoriaGlobals.Config.Partitions {
		particion := memoriaTypes.Particion{
			Base:    base,
			Limite:  base + tamanio,
			Ocupado: false,
		}

		f.Particiones = append(f.Particiones, particion)

		base += tamanio
	}

}

func (f *Fijas) AsignarProcesoAParticion(pid types.Pid, size int) error {
	err, particionEncontrada := memoriaGlobals.EstrategiaAsignacion.BuscarParticion(size)
	if err != nil {
		logger.Error("La estrategia de asignacion no ha podido asignar el proceso a una particion")
	}

	particion := f.obtenerParticion(particionEncontrada.Base, particionEncontrada.Limite)

	particion.Ocupado = true
	particion.Pid = pid
	logger.Debug("Proceso (< %v >) asignado en particiones fijas", pid)

	return nil
}

func (f *Fijas) obtenerParticion(base int, limite int) *memoriaTypes.Particion {
	for i := range f.Particiones {
		particion := &f.Particiones[i]
		if particion.Base == base && particion.Limite == limite {
			return particion
		}
	}
	return nil
}

func (f *Fijas) LiberarParticion(pid types.Pid) error {
	encontrada := false
	for _, particion := range f.Particiones {
		if particion.Pid == pid {
			particion.Ocupado = false
			encontrada = true
			break
		}
	}
	if !encontrada {
		return errors.New("no se pudo liberar la particion del proceso")
	}
	logger.Debug("Proceso (< %v >) liberado", pid)
	return nil
}
