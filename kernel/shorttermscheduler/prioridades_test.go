package shorttermscheduler

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"math/rand"
	"testing"
	"time"
)

var prioridades *Prioridades

func setup() {
	logger.ConfigureLogger("test.log", "INFO")
	prioridades = &Prioridades{}
}

// Test: probá _todo junto
func TestPrioridades(t *testing.T) {
	setup()

	correctSlice := []kerneltypes.TCB{
		kerneltypes.TCB{Prioridad: 0, TID: 1},
		kerneltypes.TCB{Prioridad: 0, TID: 2},
		kerneltypes.TCB{Prioridad: 1, TID: 3},
		kerneltypes.TCB{Prioridad: 2, TID: 4},
		kerneltypes.TCB{Prioridad: 3, TID: 5},
		kerneltypes.TCB{Prioridad: 3, TID: 6},
		kerneltypes.TCB{Prioridad: 4, TID: 7},
		kerneltypes.TCB{Prioridad: 5, TID: 8},
	}

	testSlice := []kerneltypes.TCB{
		kerneltypes.TCB{Prioridad: 5, TID: 8},
		kerneltypes.TCB{Prioridad: 0, TID: 1},
		kerneltypes.TCB{Prioridad: 1, TID: 3},
		kerneltypes.TCB{Prioridad: 2, TID: 4},
		kerneltypes.TCB{Prioridad: 3, TID: 5},
		kerneltypes.TCB{Prioridad: 0, TID: 2},
		kerneltypes.TCB{Prioridad: 4, TID: 7},
		kerneltypes.TCB{Prioridad: 3, TID: 6},
	}

	for _, v := range testSlice {
		prioridades.AddToReady(v)
	}

	for _, v := range correctSlice {
		planned, _ := prioridades.Planificar()
		if v != planned {
			t.Errorf("No se planificó de acuerdo al algoritmo")
			return
		}
	}

}

// Test: si shuffleo la lista, sigue insertando por orden de prioridades??
func TestAddToReady(t *testing.T) {
	setup()

	correctSlice := []kerneltypes.TCB{
		kerneltypes.TCB{Prioridad: 0, TID: 1},
		kerneltypes.TCB{Prioridad: 1, TID: 2},
		kerneltypes.TCB{Prioridad: 2, TID: 3},
		kerneltypes.TCB{Prioridad: 3, TID: 4},
		kerneltypes.TCB{Prioridad: 4, TID: 5},
		kerneltypes.TCB{Prioridad: 5, TID: 6},
	}

	var testSlice []kerneltypes.TCB
	testSlice = append(testSlice, correctSlice...)

	copy(testSlice, correctSlice)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(testSlice), func(i, j int) { testSlice[i], testSlice[j] = testSlice[j], testSlice[i] })

	for _, v := range testSlice {
		prioridades.AddToReady(v)
	}

	if len(correctSlice) != len(prioridades.readyThreads) {
		t.Errorf("No son del mismo tamaño\nCorrect slice: %v\nReceived Slice: %v\nTest slice: %v", correctSlice, prioridades.readyThreads, testSlice)
		return
	}

	for i := range correctSlice {
		if correctSlice[i] != prioridades.readyThreads[i] {
			t.Errorf("\nCorrect slice: %v\nReceived Slice: %v\nTest slice: %v", correctSlice, prioridades.readyThreads, testSlice)
			return
		}
	}

	logger.Debug("\nCorrect slice: %v\nReceived Slice: %v\nTest slice: %v\n", correctSlice, prioridades.readyThreads, testSlice)

}

// Ok, inserta por prioridades, pero si llegan dos hilos con misma prioridad, hace FIFO?
func TestAddToReadyFIFO(t *testing.T) {
	setup()

	correctSlice := []kerneltypes.TCB{
		kerneltypes.TCB{Prioridad: 0, TID: 1},
		kerneltypes.TCB{Prioridad: 0, TID: 2},
		kerneltypes.TCB{Prioridad: 1, TID: 3},
		kerneltypes.TCB{Prioridad: 1, TID: 4},
		kerneltypes.TCB{Prioridad: 2, TID: 5},
		kerneltypes.TCB{Prioridad: 2, TID: 6},
	}

	for _, v := range correctSlice {
		prioridades.AddToReady(v)
	}

	if len(correctSlice) != len(prioridades.readyThreads) {
		t.Errorf("No son del mismo tamaño\nCorrect slice: %v\nReceived Slice: %v", correctSlice, prioridades.readyThreads)
		return
	}

	for i := range correctSlice {
		if correctSlice[i] != prioridades.readyThreads[i] {
			t.Errorf("\nCorrect slice: %v\nReceived Slice: %v", correctSlice, prioridades.readyThreads)
			return
		}
	}

	logger.Debug("\nCorrect slice: %v\nReceived Slice: %v\n", correctSlice, prioridades.readyThreads)

}
