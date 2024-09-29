package memoria_helpers

import (
	"github.com/sisoputnfrba/tp-golang/memoria/config"
	"github.com/sisoputnfrba/tp-golang/types"
)

var Config config.MemoriaConfig
var ExecContext = make(map[types.Thread]types.ExecutionContext)
var IndexInstructionsLists = make(map[types.Thread][]string)
var UserMem = make([]byte, Config.MemorySize)
