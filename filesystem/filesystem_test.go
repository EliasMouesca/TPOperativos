package main

import (
	"errors"
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
