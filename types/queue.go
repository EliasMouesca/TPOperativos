package types

import (
	"errors"
	"sync"
)

type Queue[T any] struct {
	elements []T
	mutex    sync.Mutex
	Priority int
}

func (c *Queue[T]) Add(t *T) {
	c.mutex.Lock()
	c.elements = append(c.elements, *t)
	c.mutex.Unlock()
}

func (c *Queue[T]) GetAndRemoveNext() (*T, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if len(c.elements) == 0 {
		return nil, errors.New("no hay elementos para quitar de la cola")
	}
	nextThread := &c.elements[0]
	c.elements = c.elements[1:]

	return nextThread, nil
}

func (c *Queue[T]) IsEmpty() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.elements) == 0
}

func (c *Queue[T]) Size() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.elements)
}

func (c *Queue[T]) Do(f func(*T)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for i := range c.elements {
		f(&c.elements[i])
	}
}

func (c *Queue[T]) Remove(t *T) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for i, element := range c.elements {
		if &element == t {
			c.elements = append(c.elements[:i], c.elements[i+1:]...)
			return nil
		}
	}
	return errors.New("elemento no encontrado en la cola")
}
