package memoriaTypes

import "github.com/sisoputnfrba/tp-golang/types"

type ParticionesInterface interface {
	AsignarProcesoAParticion(pid types.Pid, size int) (base uint32, err error)
	LiberarParticion(pid types.Pid) error
	Init()
	Compactar()
	ObtenerParticionDeProceso(pid types.Pid) (Particion, error)
}
