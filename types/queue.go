package types

import "errors"

type Queue[T any] struct {
	elements []T
}

func (c *Queue[T]) Add(t *T) {
	c.elements = append(c.elements, *t)
}

func (c *Queue[T]) GetAndRemoveNext() (*T, error) {
	if len(c.elements) == 0 {
		return nil, errors.New("No hay elementos para quitar de la cola")
	}
	nextThread := &c.elements[0]
	c.elements = c.elements[1:]

	return nextThread, nil
}

func (c *Queue[T]) IsEmpty() bool {
	return len(c.elements) == 0
}
