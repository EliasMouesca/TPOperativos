package dinamicas

import (
	"errors"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaGlobals"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaTypes"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

type Dinamicas struct {
	Particiones []memoriaTypes.Particion
}

func (d *Dinamicas) Init() {
	logger.Debug("******** Inicializando Particiones Dinámicas")
	memSize := memoriaGlobals.Config.MemorySize
	//	d.Particiones = make([]memoriaTypes.Particion, memSize)
	d.Particiones = append(d.Particiones, memoriaTypes.Particion{
		Base:    0,
		Limite:  memSize,
		Ocupado: false,
		Pid:     -1,
	})
}

func (d *Dinamicas) AsignarProcesoAParticion(pid types.Pid, size int) (base uint32, err error) {
	err, particionEncontrada := memoriaGlobals.EstrategiaAsignacion.BuscarParticion(size, &d.Particiones)
	logger.Debug("Particion encontrada: %v", particionEncontrada)
	if err != nil {
		logger.Debug("Error en buscarParticion")
		if d.hayEspacioLibreSuficiente(size) {
			logger.Debug("si hay espacio suficiente")
			return 0, errors.New(types.Compactacion)
		} else {
			logger.Warn("La estrategia de asignacion no ha podido asignar el proceso a una particion")
			return 0, err
		}
	}

	tamParticion := particionEncontrada.Limite - particionEncontrada.Base
	var nuevasParticiones []memoriaTypes.Particion

	for i, particion := range d.Particiones {
		if particion.Base == particionEncontrada.Base {
			// Si a la particion encontrada le sobra espacio
			if tamParticion > size { // Chequeo si el tamanio de la particion es mayor a lo que el proceso requiere
				nuevasParticiones = append(
					nuevasParticiones, // vacia
					memoriaTypes.Particion{ // Particion del proceso
						Base:    particionEncontrada.Base,
						Limite:  size + particionEncontrada.Base,
						Ocupado: true,
						Pid:     pid,
					},
					memoriaTypes.Particion{ //Particion de lo que sobro al asignar el proceso
						Base:    particionEncontrada.Base + size,
						Limite:  particionEncontrada.Base + tamParticion,
						Ocupado: false,
					},
				)

				logger.Debug("Nuevas particiones: %v", nuevasParticiones)
				logger.Debug("Particiones antes de agregar: %v", d.Particiones)

				nuevasParticiones = append(nuevasParticiones, d.Particiones[i+1:]...)
				d.Particiones = append(d.Particiones[:i], nuevasParticiones...)

				logger.Debug("Particiones luego de agregar las nuevas: %v", d.Particiones)
				logger.Debug("Se fracciono la particion Base: %v Limite: %v ", particionEncontrada.Base, particionEncontrada.Limite)

				// Si llega al else la particion encontrada tiene el justo tamaño del proceso asi que se le asgina y listo
			} else if tamParticion == size {
				particionEncontrada.Ocupado = true
				particionEncontrada.Pid = pid
			}
			break
		}

	}
	logger.Debug("Se asigno el proceso PID: < %v > a la particion Base: %v Limite %v", pid, particionEncontrada.Base, particionEncontrada.Base+size)
	logger.Info("Particiones: %v", d.Particiones)
	return uint32(particionEncontrada.Base), nil
}

func (d *Dinamicas) hayEspacioLibreSuficiente(espacioRequerido int) bool {
	espacioLibre := 0

	for _, particion := range d.Particiones {
		if !particion.Ocupado {
			espacioLibre += particion.Limite - particion.Base
			logger.Debug("Espacion leido: %v", particion.Limite-particion.Base)
		}
	}
	logger.Debug("Espacion libre leido: %v", espacioLibre)
	if espacioLibre >= espacioRequerido {
		return true
	}
	return false
}

func (d *Dinamicas) Compactar() {
	logger.Debug("Entra a compactar")
	var particionesOcupadas []memoriaTypes.Particion
	espacioLibreTotal := 0

	// Recorre las particiones, mueve las ocupadas al inicio y suma el tamaño de las libres
	for _, particion := range d.Particiones {
		if particion.Ocupado {
			particionesOcupadas = append(particionesOcupadas, particion)
		} else {
			espacioLibreTotal += particion.Limite - particion.Base
		}
	}

	// Calcula nuevas bases y límites para las particiones ocupadas
	baseActual := 0
	for i := range particionesOcupadas {
		tamano := particionesOcupadas[i].Limite - particionesOcupadas[i].Base
		particionesOcupadas[i].Base = baseActual
		particionesOcupadas[i].Limite = baseActual + tamano
		baseActual = particionesOcupadas[i].Limite
	}

	// Crea una partición libre con el espacio total restante
	if espacioLibreTotal > 0 {
		particionLibre := memoriaTypes.Particion{
			Base:    baseActual,
			Limite:  baseActual + espacioLibreTotal,
			Ocupado: false,
			Pid:     0,
		}
		particionesOcupadas = append(particionesOcupadas, particionLibre)
	}

	// Actualiza la lista de particiones
	d.Particiones = particionesOcupadas
	logger.Info("Compactación completada. Particiones actuales: %v", d.Particiones)
}

func (d *Dinamicas) LiberarParticion(pid types.Pid) error {
	logger.Debug("Particiones: %v", d.Particiones)
	encontrada := false
	for i := range d.Particiones {
		particion := d.Particiones[i]

		if particion.Pid == pid {
			// Marcar la partición como libre
			d.Particiones[i].Ocupado = false
			d.Particiones[i].Pid = -1
			encontrada = true

			// Consolidación de particiones libres adyacentes
			if i > 0 && !d.Particiones[i-1].Ocupado && i+1 < len(d.Particiones) && !d.Particiones[i+1].Ocupado {
				// Caso 1: La partición actual y ambas adyacentes son libres
				d.Particiones[i-1].Limite = d.Particiones[i+1].Limite
				/*
					Acá no estaría eliminando la partición i+1
					d.Particiones = append(d.Particiones[:i], d.Particiones[i+1:]...)
				*/
				d.Particiones = append(d.Particiones[:i], d.Particiones[i+2:]...)
			} else if i > 0 && !d.Particiones[i-1].Ocupado {
				// Caso 2: La partición anterior es libre
				d.Particiones[i-1].Limite = particion.Limite
				d.Particiones = append(d.Particiones[:i], d.Particiones[i+1:]...)

			} else if i+1 < len(d.Particiones) && !d.Particiones[i+1].Ocupado {
				// Caso 3: La partición siguiente es libre
				d.Particiones[i+1].Base = d.Particiones[i].Base
				d.Particiones = append(d.Particiones[:i], d.Particiones[i+1:]...)
			}
			break
		}
	}
	if !encontrada {
		return fmt.Errorf("no se encontro particion que contenga el proceso PID: < %v >", pid)
	}

	logger.Debug("Proceso (< %v >) liberado", pid)
	logger.Info("Particiones actuales: %v", d.Particiones)
	return nil
}

func (d *Dinamicas) ObtenerParticionDeProceso(pid types.Pid) (memoriaTypes.Particion, error) {
	for _, particion := range d.Particiones {
		if particion.Pid == pid {
			return particion, nil
		}
	}
	return memoriaTypes.Particion{}, fmt.Errorf("no se encontro una particion con el proceso PID: %v", pid)
}
