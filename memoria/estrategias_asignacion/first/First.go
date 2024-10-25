package first

import (
	"errors"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaTypes"
)

type First struct {
}

func BuscarParticion(size int, f []memoriaTypes.Particion) (error, memoriaTypes.Particion) {
	var particionSeleccionada memoriaTypes.Particion
	encontrada := false

	for _, particion := range f {
		tamanoParticion := particion.Limite - particion.Base
		if !particion.Ocupado && tamanoParticion >= size {
			particionSeleccionada = particion
			encontrada = true
			break
		}
	}

	if !encontrada {
		return errors.New("no se encontró una partición adecuada"), memoriaTypes.Particion{}
	}

	return nil, particionSeleccionada
}
