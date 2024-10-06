package ColasMultinivel

import (
	"github.com/sisoputnfrba/tp-golang/kernel/kerneltypes"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"testing"
)

var colasMultinivel *ColasMultiNivel
var tcbSliceTest []*kerneltypes.TCB

func setup() {
	logger.ConfigureLogger("test.log", "INFO")
	colasMultinivel = &ColasMultiNivel{}

	tcbSliceTest = []*kerneltypes.TCB{
		&kerneltypes.TCB{Prioridad: 1, TID: 0},
		&kerneltypes.TCB{Prioridad: 0, TID: 1},
		&kerneltypes.TCB{Prioridad: 2, TID: 2},
		&kerneltypes.TCB{Prioridad: 1, TID: 3},
		&kerneltypes.TCB{Prioridad: 2, TID: 4},
		&kerneltypes.TCB{Prioridad: 0, TID: 5},
	}
}

func TestColasMultiNivel_Planificar(t *testing.T) {
	setup()
	correctSlice := []*kerneltypes.TCB{
		&kerneltypes.TCB{Prioridad: 0, TID: 1},
		&kerneltypes.TCB{Prioridad: 0, TID: 5},
		&kerneltypes.TCB{Prioridad: 1, TID: 0},
		&kerneltypes.TCB{Prioridad: 1, TID: 3},
		&kerneltypes.TCB{Prioridad: 2, TID: 2},
		&kerneltypes.TCB{Prioridad: 2, TID: 4},
	}
	// Mando a Ready
	for _, v := range tcbSliceTest {
		colasMultinivel.AddToReady(v)
	}

	for _, correctTcb := range correctSlice {
		tcb, _ := colasMultinivel.Planificar()
		if tcb.TID != correctTcb.TID {
			t.Errorf("No se planifico adecuadamente: Expected TID %v but got %v", correctTcb.TID, tcb.TID)
			return
		}
	}
}

// Verifico si se agregaron cada tcb a su cola correspondiente
func TestColasMultiNivel_AddToReady(t *testing.T) {
	setup()

	correctZero := []kerneltypes.TCB{
		kerneltypes.TCB{Prioridad: 0, TID: 1},
		kerneltypes.TCB{Prioridad: 0, TID: 5},
	}
	correctOne := []kerneltypes.TCB{
		kerneltypes.TCB{Prioridad: 1, TID: 0},
		kerneltypes.TCB{Prioridad: 1, TID: 3},
	}
	correctTwo := []kerneltypes.TCB{
		kerneltypes.TCB{Prioridad: 2, TID: 2},
		kerneltypes.TCB{Prioridad: 2, TID: 4},
	}
	correctQueue := [][]kerneltypes.TCB{
		correctZero,
		correctOne,
		correctTwo,
	}
	// Mando a Ready
	for _, v := range tcbSliceTest {
		colasMultinivel.AddToReady(v)
	}

	for p, cola := range colasMultinivel.ReadyQueue {
		for _, v := range correctQueue[p] {
			tcb, _ := cola.GetAndRemoveNext()
			if v.TID != tcb.TID {
				t.Errorf("No se agrego a ready correctamente: Expected TID %v but got %v", v.TID, tcb.TID)
				return
			}
		}
	}
}

func TestColasMultiNivel_AddNewQueue(t *testing.T) {
	setup()
	tcb := kerneltypes.TCB{
		Prioridad: 32,
		TID:       88,
	}

	colasMultinivel.AddToReady(&tcb)

	readyQueue := colasMultinivel.ReadyQueue
	// Tiene una sola cola en el slice por eso esto anda
	for _, queue := range readyQueue {
		if queue.Priority != tcb.Prioridad {
			t.Errorf("No se creo la cola con la priorirad correcta")
		}
	}
}
