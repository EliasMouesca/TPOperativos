package main

import (
	"errors"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"strconv"
)

// An Instruction is basically a function that takes an ectx and a slice of parameters (strings)
// and returns the modified ectx and any errors encountered.
type Instruction func(types.ExecutionContext, []string) (types.ExecutionContext, error)

var instructionSet = map[string]Instruction{
	"SET": setInstruction,
	//"SUM": sumInstruction,
	//"SUB": subInstruction,
	//"READ_MEM": readMemInstruction,
	//"WRITE_MEM": writeMemInstruction,
	//"JNZ": jnzInstruction,
	"LOG": logInstruction,
}

func checkArguments(args []string, correctNumberOfArgs int) error {
	if len(args) != correctNumberOfArgs {
		return errors.New("se recibió una cantidad de argumentos no válida")
	}
	return nil
}

func setInstruction(ctx types.ExecutionContext, args []string) (types.ExecutionContext, error) {
	// Check number of arguments
	if err := checkArguments(args, 2); err != nil {
		return ctx, err
	}

	// Get the register to modify
	reg, err := ctx.GetRegister(args[0])
	if err != nil {
		return types.ExecutionContext{}, err
	}

	// Parse second argument as int
	i, err := strconv.Atoi(args[1])
	if err != nil {
		return types.ExecutionContext{}, errors.New("no se pudo parsear '" + args[1] + "' como un entero")
	}

	// Set the register
	*reg = uint32(i)
	return ctx, nil
}

func logInstruction(ctx types.ExecutionContext, args []string) (types.ExecutionContext, error) {
	if err := checkArguments(args, 1); err != nil {
		return ctx, err
	}

	registerString := args[0]

	register, err := ctx.GetRegister(registerString)
	if err != nil {
		return ctx, err
	}

	logger.Info("Logging register '%v': %v", registerString, register)
	return ctx, nil
}
