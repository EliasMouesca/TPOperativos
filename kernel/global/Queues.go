package global

import "github.com/sisoputnfrba/tp-golang/types"

//var NEW []types.PCB
//var Ready []types.TCB

// Colas de New y Ready usnaod el tipo Queue, quedaria cambiar donde se usan
var NEW types.Queue[types.PCB]
var Ready types.Queue[types.TCB]
