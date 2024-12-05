package shorttermscheduler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"net/http"
	"sync"
)

var mutex = &sync.Mutex{}

func CpuInterrupt(interruption types.Interruption) error {
	mutex.Lock()
	url := fmt.Sprintf("http://%v:%v/cpu/interrupt",
		kernelglobals.Config.CpuAddress,
		kernelglobals.Config.CpuPort)

	data, err := json.Marshal(&interruption)
	if err != nil {
		return err
	}
	logger.Debug("Enviando CPU INTERRUPT: %v", interruption.Description)
	_, err = http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}

	logger.Debug("CPU INTERRUPT enviado correctamente")
	mutex.Unlock()
	return nil
}

func CpuExecute(thread types.Thread) error {
	mutex.Lock()
	url := fmt.Sprintf("http://%v:%v/cpu/execute",
		kernelglobals.Config.CpuAddress,
		kernelglobals.Config.CpuPort)

	data, err := json.Marshal(&thread)
	if err != nil {
		return err
	}

	_, err = http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	mutex.Unlock()
	return nil

}
