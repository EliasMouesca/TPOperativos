package types

import (
	"github.com/sisoputnfrba/tp-golang/utils/logger"
	"os"
)

type Queue[T any] struct {
	elements []T
}

func (c *Queue[T]) Add(t *T) {
	c.elements = append(c.elements, *t)
}

func (c *Queue[T]) GetAndRemoveNext() T {
	if len(c.elements) == 0 {
		logger.Error("Se quizo quitar elementos de lista vacia")
		os.Exit(1)
	}
	nextTrhead := c.elements[0]
	c.elements = c.elements[1:]
	return nextTrhead
}

func (c *Queue[T]) IsEmpty() bool {
	return len(c.elements) == 0
}
