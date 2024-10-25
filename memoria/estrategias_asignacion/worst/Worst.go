package worst

import (
	"errors"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaTypes"
)

type Worst struct {
}

func BuscarParticion(size int, f []memoriaTypes.Particion) (error, memoriaTypes.Particion) {
	var particionSeleccionada memoriaTypes.Particion
	encontrada := false
	maxSize := 0

	for _, particion := range f {
		tamanoParticion := particion.Limite - particion.Base
		if maxSize == 0 {
			maxSize = tamanoParticion
		}
		if !particion.Ocupado && tamanoParticion >= size && tamanoParticion > maxSize {
			particionSeleccionada = particion
			maxSize = tamanoParticion
			encontrada = true
		}
	}

	if !encontrada {
		return errors.New("no se encontró una partición adecuada"), memoriaTypes.Particion{}
	}

	return nil, particionSeleccionada
}
