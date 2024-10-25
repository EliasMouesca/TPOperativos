package best

import (
	"errors"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaTypes"
)

func BuscarParticion(size int, f []memoriaTypes.Particion) (error, memoriaTypes.Particion) {
	var particionSeleccionada memoriaTypes.Particion
	encontrada := false
	minSize := 0

	for _, particion := range f {
		tamanoParticion := particion.Limite - particion.Base
		if minSize == 0 {
			minSize = tamanoParticion
		}
		if !particion.Ocupado && tamanoParticion >= size && tamanoParticion < minSize {
			particionSeleccionada = particion
			minSize = tamanoParticion
			encontrada = true
		}
	}

	if !encontrada {
		return errors.New("no se encontró una partición adecuada"), memoriaTypes.Particion{}
	}

	return nil, particionSeleccionada
}
