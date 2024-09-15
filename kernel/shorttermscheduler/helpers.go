package shorttermscheduler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sisoputnfrba/tp-golang/kernel/kernelglobals"
	"github.com/sisoputnfrba/tp-golang/types"
	"net/http"
)

func cpuInterrupt(interruption types.Interruption) error {
	url := fmt.Sprintf("http://%v:%v/cpu/interrupt",
		kernelglobals.Config.CpuAddress,
		kernelglobals.Config.CpuPort)

	data, err := json.Marshal(&interruption)
	if err != nil {
		return err
	}

	_, err = http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}

	return nil
}

func CpuExecute(thread types.Thread) error {
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

	return nil

}
