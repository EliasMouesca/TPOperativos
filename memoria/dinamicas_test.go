package main

import (
	"github.com/sisoputnfrba/tp-golang/memoria/esquemas_particiones/dinamicas"
	"github.com/sisoputnfrba/tp-golang/memoria/memoriaGlobals"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"testing"
)

func TestAsignarProcesoAParticion(t *testing.T) {
	// Inicializa una nueva instancia de Dinamicas con memoria de tamaño 1024
	d := dinamicas.Dinamicas{}
	memoriaGlobals.Config.MemorySize = 1024
	d.Init()

	// Asigna un proceso de tamaño 100
	_, err := d.AsignarProcesoAParticion(types.Pid(1), 100)
	if err != nil {
		t.Fatalf("Error al asignar el proceso: %v", err)
	}

	if !d.Particiones[0].Ocupado {
		t.Errorf("La primera partición debe estar ocupada")
	}

	if d.Particiones[0].Pid != types.Pid(1) {
		t.Errorf("El PID de la partición debe coincidir con el proceso asignado")
	}

	if d.Particiones[1].Base != 100 {
		t.Errorf("La segunda partición debe comenzar en el límite del proceso asignado")
	}

	if d.Particiones[1].Ocupado {
		t.Errorf("La segunda partición debe estar libre")
	}

	// Asigna otro proceso de tamaño 200 y verifica el fraccionamiento correcto
	_, err = d.AsignarProcesoAParticion(types.Pid(2), 200)
	if err != nil {
		t.Fatalf("Error al asignar el segundo proceso: %v", err)
	}

	if !d.Particiones[1].Ocupado {
		t.Errorf("La segunda partición debe estar ocupada después de asignar el segundo proceso")
	}

	if d.Particiones[1].Pid != types.Pid(2) {
		t.Errorf("El PID de la segunda partición debe coincidir con el proceso asignado")
	}
}

func TestLiberarParticion(t *testing.T) {
	// Inicializa una nueva instancia de Dinamicas
	d := dinamicas.Dinamicas{}
	memoriaGlobals.Config.MemorySize = 1024
	d.Init()

	logger.Info("Test: Se crea un proceso y se lo borra => Al final solo queda una única partición libre")
	_, err := d.AsignarProcesoAParticion(types.Pid(1), 100)
	err = d.LiberarParticion(types.Pid(1))
	if err != nil {
		t.Fatalf("Error al liberar el proceso: %v", err)
	}
	if len(d.Particiones) > 1 && d.Particiones[0].Ocupado && d.Particiones[0].Pid != types.Pid(0) {
		t.Fatalf("Debería haber solo una partición libre con PID-0")
	}

	logger.Info("Test: Se crean 2 procesos y se borra el 1ro y despues el 2do => Al final solo queda una única partición libre")
	_, err = d.AsignarProcesoAParticion(types.Pid(1), 100)
	_, err = d.AsignarProcesoAParticion(types.Pid(2), 100)
	err = d.LiberarParticion(types.Pid(1))
	err = d.LiberarParticion(types.Pid(2))

	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if len(d.Particiones) > 1 && d.Particiones[0].Ocupado && d.Particiones[0].Pid != types.Pid(0) {
		t.Fatalf("Debería haber solo una partición libre con PID-0")
	}

	logger.Info("Test: Se crean 2 procesos, se borra el 2do y despues el 1ro => Al final solo queda una única partición libre")
	_, err = d.AsignarProcesoAParticion(types.Pid(1), 100)
	_, err = d.AsignarProcesoAParticion(types.Pid(2), 100)
	err = d.LiberarParticion(types.Pid(2))
	err = d.LiberarParticion(types.Pid(1))
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if len(d.Particiones) > 1 && d.Particiones[0].Ocupado && d.Particiones[0].Pid != types.Pid(0) {
		t.Fatalf("Debería haber solo una partición libre con PID-0")
	}
}

/*
func TestCompactarParticiones(t *testing.T) {
	// Inicializa una nueva instancia de Dinamicas
	d := dinamicas.Dinamicas{}
	memoriaGlobals.Config.MemorySize = 1024
	d.Init()

	// Asigna varios procesos
	_ = d.AsignarProcesoAParticion(types.Pid(1), 100)
	_ = d.AsignarProcesoAParticion(types.Pid(2), 50)
	_ = d.AsignarProcesoAParticion(types.Pid(3), 300)
	_ = d.AsignarProcesoAParticion(types.Pid(4), 200)
	_ = d.AsignarProcesoAParticion(types.Pid(5), 100)
	_ = d.AsignarProcesoAParticion(types.Pid(6), 274)

	// Libera uno de los procesos para crear fragmentación
	_ = d.LiberarParticion(types.Pid(2))
	_ = d.LiberarParticion(types.Pid(5))

	_ = d.AsignarProcesoAParticion(types.Pid(7), 120)
	logger.Debug("Particiones actuales: %v", d.Particiones)

}

func TestCrearCuatroYLiberarPrimera(t *testing.T) {
	// Inicializa una nueva instancia de Dinamicas
	d := dinamicas.Dinamicas{}
	memoriaGlobals.Config.MemorySize = 1024
	d.Init()

	// Asigna cuatro procesos
	_ = d.AsignarProcesoAParticion(types.Pid(1), 100)
	_ = d.AsignarProcesoAParticion(types.Pid(2), 200)
	_ = d.AsignarProcesoAParticion(types.Pid(3), 300)
	_ = d.AsignarProcesoAParticion(types.Pid(4), 400)
	_ = d.AsignarProcesoAParticion(types.Pid(5), 24)
	// Libera el primer proceso
	err := d.LiberarParticion(types.Pid(1))
	if err != nil {
		t.Fatalf("Error al liberar el proceso: %v", err)
	}

	// Verifica que la primera partición esté libre
	if d.Particiones[0].Ocupado {
		t.Errorf("La primera partición debe estar libre después de liberar el proceso")
	}
	if d.Particiones[0].Base != 0 || d.Particiones[0].Limite != 100 {
		t.Errorf("La primera partición debe tener Base: 0 y Limite: 100, pero tiene Base: %v y Limite: %v", d.Particiones[0].Base, d.Particiones[0].Limite)
	}

	// Verifica que las otras particiones permanezcan ocupadas con los procesos asignados
	if !d.Particiones[1].Ocupado || d.Particiones[1].Pid != types.Pid(2) {
		t.Errorf("La segunda partición debe estar ocupada por el proceso con PID 2")
	}
	if !d.Particiones[2].Ocupado || d.Particiones[2].Pid != types.Pid(3) {
		t.Errorf("La tercera partición debe estar ocupada por el proceso con PID 3")
	}
	if !d.Particiones[3].Ocupado || d.Particiones[3].Pid != types.Pid(4) {
		t.Errorf("La cuarta partición debe estar ocupada por el proceso con PID 4")
	}
}
*/
