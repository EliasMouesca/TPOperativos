package worst

import (
	"errors"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaTypes"
)

type Worst struct {
}

func (w *Worst) BuscarParticion(size int, f *[]memoriaTypes.Particion) (error, *memoriaTypes.Particion) {
	var particionSeleccionada *memoriaTypes.Particion
	encontrada := false
	maxSize := 0

	for i, particion := range *f {
		if !particion.Ocupado {
			tamanoParticion := particion.Limite - particion.Base
			if maxSize == 0 && tamanoParticion >= size {
				maxSize = tamanoParticion
			}
			if tamanoParticion >= size && tamanoParticion >= maxSize {
				particionSeleccionada = &(*f)[i]
				maxSize = tamanoParticion
				encontrada = true
			}
		}

	}

	if !encontrada {
		return errors.New("no se encontró una partición adecuada"), nil
	}

	return nil, particionSeleccionada
}
