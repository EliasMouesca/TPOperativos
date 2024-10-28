package dinamicas

import (
	"fmt"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaGlobals"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaTypes"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

type Dinamicas struct {
	// TODO: Si la lista fueran punteros a particiones seria todo mas facil :)
	Particiones []memoriaTypes.Particion
}

func (d *Dinamicas) init() {
	logger.Debug("******** Inicializando Particiones Dinámicas")
	memSize := memoriaGlobals.Config.MemorySize
	d.Particiones = make([]memoriaTypes.Particion, memSize)
	d.Particiones = append(d.Particiones, memoriaTypes.Particion{
		Base:    0,
		Limite:  memSize,
		Ocupado: false,
	})
}

func (d *Dinamicas) AsignarProcesoAParticion(pid types.Pid, size int) error {
	err, particionEncontrada := memoriaGlobals.EstrategiaAsignacion.BuscarParticion(size, &d.Particiones)
	if err != nil {
		if d.hayEspacioLibreSuficiente(size) {
			// TODO: Que onda esto?
			// notificarKernelNecesidadDeCompactar
			// if kernel me dijo que si
			d.compactarParticiones()
		} else {
			logger.Error("La estrategia de asignacion no ha podido asignar el proceso a una particion")
			return err
		}
	}
	tamParticion := particionEncontrada.Limite - particionEncontrada.Base

	// Si a la particion encontrada le sobra espacio
	if tamParticion > size {
		// Se divide en una particion del espacio requerido por el proceso
		d.Particiones = append(d.Particiones, memoriaTypes.Particion{
			Base:    particionEncontrada.Base,
			Limite:  size,
			Ocupado: true,
			Pid:     pid,
		},
			// Y una partcion del espacio que sobra
			memoriaTypes.Particion{
				Base:    particionEncontrada.Base + size,
				Limite:  tamParticion,
				Ocupado: false,
			})
		logger.Debug("Se fracciono la particion Base: %v Limite: %v ", particionEncontrada.Base, particionEncontrada.Limite)

		// Si llega al else la particion encontrada tiene el justo tamaño del proceso asi que se le asgina y listo
	} else {
		particionEncontrada.Ocupado = false
		particionEncontrada.Pid = pid
	}

	logger.Debug("Se asigno el proceso PID: < %v > a la particion Base: %v Limite %v", pid, particionEncontrada.Base, particionEncontrada.Limite)
	return nil
}

func (d *Dinamicas) hayEspacioLibreSuficiente(espacioRequerido int) bool {
	espacioLibre := 0
	for _, particion := range d.Particiones {
		if !particion.Ocupado {
			espacioLibre = particion.Limite - particion.Base
		}
	}
	if espacioLibre >= espacioRequerido {
		return true
	}
	return false
}

func (d *Dinamicas) compactarParticiones() {
	// TODO: Aca hay un tema porque particion no es un puntero es una copia asi que en realidad no hace nada
	proximaBase := 0
	for _, particion := range d.Particiones {
		if particion.Ocupado {
			particion.Base = proximaBase
			particion.Limite = particion.Limite - particion.Base + proximaBase
		}
	}
}

func (d *Dinamicas) LiberarParticion(pid types.Pid) error {
	encontrada := false
	for i, particion := range d.Particiones {
		if particion.Pid == pid {
			particion.Ocupado = false
			encontrada = true
			// Consolidacion de particiones libres aledañas
			if !d.Particiones[i-1].Ocupado {

				d.Particiones[i-1].Limite = particion.Limite
				d.Particiones = append(d.Particiones[:i], d.Particiones[i+1:]...)

			} else if !d.Particiones[i+1].Ocupado {

				particion.Limite = d.Particiones[i+1].Limite
				d.Particiones = append(d.Particiones[:i+1], d.Particiones[i+2:]...)

			} else if !d.Particiones[i+1].Ocupado && !d.Particiones[i-1].Ocupado {

				particion.Base = d.Particiones[i-1].Base
				particion.Limite = d.Particiones[i+1].Limite
				d.Particiones = append(d.Particiones[:i+1], d.Particiones[i+2:]...)
				d.Particiones = append(d.Particiones[:i-1], d.Particiones[i:]...)
			}
		}
		break
	}
	if !encontrada {
		return fmt.Errorf("no se encontro particion que contenga el proceso PID: < %v >", pid)
	}
	logger.Debug("Proceso (< %v >) liberado", pid)
	return nil
}
