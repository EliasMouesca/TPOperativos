package best

import (
	"errors"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaTypes"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
)

type Best struct{}

func (b *Best) BuscarParticion(size int, f *[]memoriaTypes.Particion) (error, *memoriaTypes.Particion) {
	var particionSeleccionada *memoriaTypes.Particion
	encontrada := false
	minSize := 0

	for i, particion := range *f {
		tamanoParticion := particion.Limite - particion.Base
		if minSize == 0 && tamanoParticion >= size && !particion.Ocupado {
			minSize = tamanoParticion
			logger.Debug("MinSize inical: %v", minSize)
		}
		if minSize != 0 && !particion.Ocupado && tamanoParticion >= size && tamanoParticion <= minSize {
			// TODO: Esto es feo si, pero particion es una copia del slice asi que la forma de
			// devolver un puntero al slice que le pasamos por parametro viene a ser esta :)
			particionSeleccionada = &(*f)[i]
			minSize = tamanoParticion
			encontrada = true
		}
	}
	logger.Debug("Particion encontrada: %v", particionSeleccionada)
	if !encontrada {
		return errors.New("no se encontró una partición adecuada"), nil
	}

	return nil, particionSeleccionada
}
