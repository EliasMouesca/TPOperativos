package main

import (
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"testing"
)

var context types.ExecutionContext

func setup() {
	logger.ConfigureLogger("cpu.log", "TRACE")
	context = types.ExecutionContext{
		Ax:         0,
		Bx:         0,
		Cx:         1,
		Dx:         2,
		Pc:         10,
		MemoryBase: 0x0,
		MemorySize: 10,
	}
}

func TestJnzInstruction(t *testing.T) {
	setup()
	err := jnzInstruction(&context, []string{"DX", "4"})
	if err != nil {
		t.Error(err)
	}

	if context.Pc != 4 {
		t.Errorf("JnzInstruction expected pc to be 4, got %d", context.Pc)
	}

	context.Pc = 23
	err = jnzInstruction(&context, []string{"Ax", "4"})
	if err != nil {
		t.Error(err)
	}
	if context.Pc != 23 {
		t.Errorf("JnzInstruction expected pc to be 23, got %d", context.Pc)
	}
}

func TestSumInstruction(t *testing.T) {
	setup()

	err := sumInstruction(&context, []string{"CX", "DX"})
	if err != nil {
		t.Error(err)
	}
	if context.Cx != 3 {
		t.Errorf("SumInstruction expected cx to be 3, got %d", context.Cx)
	}

}

func TestSubInstruction(t *testing.T) {
	setup()

	err := subInstruction(&context, []string{"DX", "dx"})
	if err != nil {
		t.Error(err)
	}
	if context.Dx != 0 {
		t.Errorf("SubInstruction expected cx to be 0, got %d", context.Dx)
	}
}

func TestSetInstruction(t *testing.T) {
	setup()
	err := setInstruction(&context, []string{"hx", "4"})
	if err != nil {
		t.Error(err)
	}
	if context.Hx != 4 {
		t.Errorf("SetInstruction expected hx to be 4, got %d", context.Hx)
	}

	err = setInstruction(&context, []string{"cx", "hx"})
	if err != nil {
		t.Error(err)
	}
	if context.Cx != 4 {
		t.Errorf("SetInstruction expected cx to be 4, got %d", context.Cx)
	}

}

// -- Este test no anda más porque el current thread es nil, pq no está ejecutando la CPU.
// Este test solo sirve hasta el próximo checkpoint, después habría que escribir a una dirección
// y luego leer de la misma dirección a ver si es lo mismo
/*func TestReadMemory(t *testing.T) {
	setup()
	err := readMemInstruction(&context, []string{"ax", "BX"})
	if err != nil {
		t.Error(err)
	}
	if context.Ax != 0xdeadbeef {
		t.Errorf("ReadMemory expected ax to be 0xdeadbeef, got %d", context.Ax)
	}
}
*/
