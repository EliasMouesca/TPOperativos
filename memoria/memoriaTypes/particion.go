package memoriaTypes

import (
	"github.com/sisoputnfrba/tp-golang/types"
)

type Particion struct {
	Base    int
	Limite  int
	Ocupado bool
	Pid     types.Pid
}
