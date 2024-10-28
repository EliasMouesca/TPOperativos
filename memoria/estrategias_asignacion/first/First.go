package first

import (
	"errors"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaTypes"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

type First struct{}

func (s *First) BuscarParticion(size int, f *[]memoriaTypes.Particion) (error, *memoriaTypes.Particion) {
	var particionSeleccionada *memoriaTypes.Particion
	encontrada := false
	logger.Debug("Particiones recibidas: %v", f)
	for i, particion := range *f {
		tamanoParticion := particion.Limite - particion.Base
		if !particion.Ocupado && tamanoParticion >= size {
			particionSeleccionada = &(*f)[i]
			encontrada = true
			break
		}
	}

	if !encontrada {
		return errors.New("no se encontró una partición adecuada"), nil
	}

	return nil, particionSeleccionada
}
