package memoria_helpers

import (
	"github.com/sisoputnfrba/tp-golang/types"
)

var Config MemoriaConfig
var ExecContext = make(map[types.Thread]types.ExecutionContext)
var CodeRegionForThreads = make(map[types.Thread][]string)
var UserMem = make([]byte, Config.MemorySize)
