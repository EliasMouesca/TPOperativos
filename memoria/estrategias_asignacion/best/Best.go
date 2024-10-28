package best

import (
	"errors"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaTypes"
)

type Best struct{}

func (b *Best) BuscarParticion(size int, f *[]memoriaTypes.Particion) (error, *memoriaTypes.Particion) {
	var particionSeleccionada *memoriaTypes.Particion
	encontrada := false
	minSize := 0

	for i, particion := range *f {
		tamanoParticion := particion.Limite - particion.Base
		if minSize == 0 {
			minSize = tamanoParticion
		}
		if !particion.Ocupado && tamanoParticion >= size && tamanoParticion < minSize {
			// TODO: Esto es feo si, pero particion es una copia del slice asi que la forma de
			// devolver un puntero al slice que le pasamos por parametro viene a ser esta :)
			particionSeleccionada = &(*f)[i]
			minSize = tamanoParticion
			encontrada = true
		}
	}

	if !encontrada {
		return errors.New("no se encontró una partición adecuada"), nil
	}

	return nil, particionSeleccionada
}
