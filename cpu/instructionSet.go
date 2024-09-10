package main

import (
	"errors"
	"github.com/sisoputnfrba/tp-golang/types"
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"strconv"
)

// An Instruction is basically a function that takes an ectx and a slice of parameters (strings)
// and returns the modified ectx and any errors encountered.
type Instruction func(*types.ExecutionContext, []string) error

var instructionSet = map[string]Instruction{
	"SET": setInstruction,
	"SUM": sumInstruction,
	"SUB": subInstruction,
	//"READ_MEM": readMemInstruction,
	//"WRITE_MEM": writeMemInstruction,
	"JNZ": jnzInstruction,
	"LOG": logInstruction,
	"MUTEX_CREATE",
}

func jnzInstruction(context *types.ExecutionContext, arguments []string) error {
	if err := checkArguments(arguments, 2); err != nil {
		return err
	}

	register, err := context.GetRegister(arguments[0])
	if err != nil {
		return err
	}

	jump, err := strconv.Atoi(arguments[1])
	if err != nil {
		return err
	}

	if *register == 0 {
		context.Pc = uint32(jump)
	}

	return nil

}

func sumInstruction(context *types.ExecutionContext, args []string) error {
	if err := checkArguments(args, 2); err != nil {
		return err
	}

	firstRegister, err := context.GetRegister(args[0])
	if err != nil {
		return err
	}

	secondRegister, err := context.GetRegister(args[1])
	if err != nil {
		return err
	}

	*firstRegister = *firstRegister + *secondRegister

	return nil

}

func subInstruction(context *types.ExecutionContext, args []string) error {
	if err := checkArguments(args, 2); err != nil {
		return err
	}

	firstRegister, err := context.GetRegister(args[0])
	if err != nil {
		return err
	}

	secondRegister, err := context.GetRegister(args[1])
	if err != nil {
		return err
	}

	*firstRegister = *firstRegister - *secondRegister

	return nil

}

func setInstruction(ctx *types.ExecutionContext, args []string) error {
	// Check number of arguments
	if err := checkArguments(args, 2); err != nil {
		return err
	}

	// Get the register to modify
	reg, err := ctx.GetRegister(args[0])
	if err != nil {
		return err
	}

	// Parse second argument as int
	i, err := strconv.Atoi(args[1])
	if err != nil {
		return errors.New("no se pudo parsear '" + args[1] + "' como un entero")
	}

	// Set the register
	*reg = uint32(i)
	return nil
}

func logInstruction(ctx *types.ExecutionContext, args []string) error {
	if err := checkArguments(args, 1); err != nil {
		return err
	}

	registerString := args[0]

	register, err := ctx.GetRegister(registerString)
	if err != nil {
		return err
	}

	logger.Info("Logging register '%v': %v", registerString, register)
	return nil
}

func checkArguments(args []string, correctNumberOfArgs int) error {
	if len(args) != correctNumberOfArgs {
		return errors.New("se recibió una cantidad de argumentos no válida")
	}
	return nil
}
