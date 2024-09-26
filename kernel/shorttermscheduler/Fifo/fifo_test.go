package Fifo

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"testing"
)

var fifo *Fifo

func setupFifoTest() {
	logger.ConfigureLogger("test.log", "INFO")
	fifo = &Fifo{}
}

func TestFifo(t *testing.T) {
	setupFifoTest()

	testSlice := []*kerneltypes.TCB{
		&kerneltypes.TCB{Prioridad: 5, TID: 8},
		&kerneltypes.TCB{Prioridad: 0, TID: 1},
		&kerneltypes.TCB{Prioridad: 1, TID: 3},
		&kerneltypes.TCB{Prioridad: 2, TID: 4},
		&kerneltypes.TCB{Prioridad: 3, TID: 5},
		&kerneltypes.TCB{Prioridad: 0, TID: 2},
		&kerneltypes.TCB{Prioridad: 4, TID: 7},
		&kerneltypes.TCB{Prioridad: 3, TID: 6},
	}

	for _, v := range testSlice {
		fifo.AddToReady(v)
	}

	for _, v := range testSlice {
		planned, _ := fifo.Planificar()
		if v.TID != planned.TID {
			t.Errorf("No se planificó de acuerdo al algoritmo")
			return
		}
	}

}
