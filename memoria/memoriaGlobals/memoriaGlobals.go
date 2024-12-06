package memoriaGlobals

import (
	"github.com/sisoputnfrba/tp-golang/memoria/config"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaTypes"
	"github.com/sisoputnfrba/tp-golang/types"
	"sync"
)

var EstrategiaAsignacion memoriaTypes.EstrategiasAsignacionInterface
var SistemaParticiones memoriaTypes.ParticionesInterface

var Config config.MemoriaConfig
var ExecContext = make(map[types.Thread]types.ExecutionContext)
var CodeRegionForThreads = make(map[types.Thread][]string)

var UserMem []byte

var MutexContext sync.Mutex
