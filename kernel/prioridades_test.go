package main

import (
	fmt "fmt"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"math/rand"
	"testing"
	"time"
)

var prioridades *Prioridades

func setup() {
	logger.ConfigureLogger("test.log", "DEBUG")
	prioridades = &Prioridades{}
}

// Test: probá _todo junto
func TestPrioridades(t *testing.T) {
	setup()
	logger.Debug("== Test algoritmo de prioridades ==")

	correctSlice := []TCB{
		TCB{Prioridad: 0, TID: 1},
		TCB{Prioridad: 0, TID: 2},
		TCB{Prioridad: 1, TID: 3},
		TCB{Prioridad: 2, TID: 4},
		TCB{Prioridad: 3, TID: 5},
		TCB{Prioridad: 3, TID: 6},
		TCB{Prioridad: 4, TID: 7},
		TCB{Prioridad: 5, TID: 8},
	}

	testSlice := []TCB{
		TCB{Prioridad: 5, TID: 8},
		TCB{Prioridad: 0, TID: 1},
		TCB{Prioridad: 1, TID: 3},
		TCB{Prioridad: 2, TID: 4},
		TCB{Prioridad: 3, TID: 5},
		TCB{Prioridad: 0, TID: 2},
		TCB{Prioridad: 4, TID: 7},
		TCB{Prioridad: 3, TID: 6},
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
	logger.Debug("== Running TestAddToReady ==")

	correctSlice := []TCB{
		TCB{Prioridad: 0, TID: 1},
		TCB{Prioridad: 1, TID: 2},
		TCB{Prioridad: 2, TID: 3},
		TCB{Prioridad: 3, TID: 4},
		TCB{Prioridad: 4, TID: 5},
		TCB{Prioridad: 5, TID: 6},
	}

	var testSlice []TCB
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

	fmt.Printf("\nCorrect slice: %v\nReceived Slice: %v\nTest slice: %v\n", correctSlice, prioridades.readyThreads, testSlice)

}

// Ok, inserta por prioridades, pero si llegan dos hilos con misma prioridad, hace FIFO?
func TestAddToReadyFIFO(t *testing.T) {
	setup()
	logger.Debug("== Running TestAddToReady FIFO==")

	correctSlice := []TCB{
		TCB{Prioridad: 0, TID: 1},
		TCB{Prioridad: 0, TID: 2},
		TCB{Prioridad: 1, TID: 3},
		TCB{Prioridad: 1, TID: 4},
		TCB{Prioridad: 2, TID: 5},
		TCB{Prioridad: 2, TID: 6},
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

	fmt.Printf("\nCorrect slice: %v\nReceived Slice: %v\n", correctSlice, prioridades.readyThreads)

}
