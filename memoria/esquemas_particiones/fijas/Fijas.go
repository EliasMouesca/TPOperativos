package fijas

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaGlobals"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaTypes"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

type Fijas struct {
	Particiones []memoriaTypes.Particion
}

func (f *Fijas) Init() {
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
	err, particionEncontrada := memoriaGlobals.EstrategiaAsignacion.BuscarParticion(size, &f.Particiones)
	if err != nil {
		logger.Error("La estrategia de asignacion no ha podido asignar el proceso a una particion")
	}

	particionEncontrada.Ocupado = true
	particionEncontrada.Pid = pid
	logger.Debug("Proceso (< %v >) asignado en particiones fijas", pid)

	return nil
}

// No hace falta
//func (f *Fijas) obtenerParticion(base int, limite int) *memoriaTypes.Particion {
//	for i := range f.Particiones {
//		particion := &f.Particiones[i]
//		if particion.Base == base && particion.Limite == limite {
//			return particion
//		}
//	}
//	return nil
//}

func (f *Fijas) LiberarParticion(pid types.Pid) error {
	encontrada := false
	for _, particion := range f.Particiones {
		if particion.Pid == pid {
			particion.Ocupado = false
			encontrada = true
			logger.Debug("Particion encontrada: Base: %v")
			break
		}
	}
	if !encontrada {
		return fmt.Errorf("no se encontro particion que contenga el proceso PID: < %v >", pid)
	}
	logger.Debug("Proceso (< %v >) liberado", pid)
	return nil
}
