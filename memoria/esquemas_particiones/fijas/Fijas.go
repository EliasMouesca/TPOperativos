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

func (f *Fijas) Compactar() {

}

func (f *Fijas) Init() {
	logger.Debug("Inicializando particiones fijas")
	base := 0
	for _, tamanio := range memoriaGlobals.Config.Partitions {
		particion := memoriaTypes.Particion{
			Base:    base,
			Limite:  base + tamanio,
			Ocupado: false,
			Pid:     -1,
		}

		f.Particiones = append(f.Particiones, particion)

		base += tamanio
	}
}

func (f *Fijas) AsignarProcesoAParticion(pid types.Pid, size int) (base uint32, err error) {
	err, particionEncontrada := memoriaGlobals.EstrategiaAsignacion.BuscarParticion(size, &f.Particiones)
	if err != nil {
		logger.Warn("La estrategia de asignacion no ha podido asignar el proceso a una particion")
		return 0, err
	}

	particionEncontrada.Ocupado = true
	particionEncontrada.Pid = pid
	logger.Debug("Proceso (< %v >) asignado en particiones fijas", pid)
	logger.Info("Particiones luego de asignar: %v", f.Particiones)
	return uint32(particionEncontrada.Base), nil
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
	for i := range f.Particiones {
		particion := f.Particiones[i]
		if particion.Pid == pid {
			f.Particiones[i].Ocupado = false
			f.Particiones[i].Pid = -1
			encontrada = true
			logger.Debug("Particion encontrada: Base: %v", particion.Base)
			break
		}
	}
	if !encontrada {
		return fmt.Errorf("no se encontro particion que contenga el proceso PID: < %v >", pid)
	}
	logger.Info("Proceso (< %v >) liberado", pid)
	logger.Info("Particiones luego de liberar: %v", f.Particiones)
	return nil
}

func (f *Fijas) ObtenerParticionDeProceso(pid types.Pid) (memoriaTypes.Particion, error) {
	for _, particion := range f.Particiones {
		if particion.Pid == pid {
			return particion, nil
		}
	}
	return memoriaTypes.Particion{}, fmt.Errorf("no se encontro una particion con el proceso PID: %v", pid)
}
