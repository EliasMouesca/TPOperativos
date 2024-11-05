package main

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

func TestInitialize(t *testing.T) {
	if _, err := os.Stat(bitmapFilename); os.IsNotExist(err) {
		config.BlockCount = 1024
		err := initialize()
		if err != nil {
			t.Error(err)
		}
		defer bitmapFile.Close()
		defer bloquesFile.Close()

		defer os.Remove(bitmapFilename)

		info, err := os.Stat(bitmapFilename)
		if err != nil {
			t.Error(err)
		}

		if info.Size() != int64((config.BlockCount+7)/8) {
			t.Error(errors.New("los tamaños no coinciden"))
		}
	}
}

func TestAllocateBlock(t *testing.T) {
	initialize()

	allocated, err := allocateBlocks(8)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(allocated)
}

func TestWriteFile(t *testing.T) {
	initialize()

	data := make([]byte, 32+8)

	for i := range data {
		data[i] = 0xff
	}

	err := writeFile("lal.dat", data)
	if err != nil {
		t.Error(err)
	}
}
