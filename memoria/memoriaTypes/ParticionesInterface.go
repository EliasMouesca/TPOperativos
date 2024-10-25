package memoriaTypes

import "github.com/sisoputnfrba/tp-golang/types"

type ParticionesInterface interface {
	AsignarProcesoAParticion(pid types.Pid, size int) error
	LiberarParticion(pid types.Pid) error
}
