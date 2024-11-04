package types

import (
	"errors"
	"sync"
)

type Nulleable[T any] interface {
	Null() T
	Equal(T) bool
}

// Queue Cola de procesos o hilos
type Queue[T Nulleable[T]] struct {
	elements []T
	mutex    sync.Mutex
	Priority int
}

func (c *Queue[T]) Contains(t T) bool {
	for _, e := range c.elements {
		if e.Equal(t) {
			return true
		}
	}
	// Si llega acá es porque la queue está vacia => No contiene el elemento
	return false
}

func (c *Queue[T]) GetElements() []T {
	return c.elements
}

func (c *Queue[T]) Add(t T) {
	c.mutex.Lock()
	c.elements = append(c.elements, t)
	c.mutex.Unlock()
}

func (c *Queue[T]) GetAndRemoveNext() (T, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if len(c.elements) == 0 {
		return T.Null(*new(T)), errors.New("no hay elementos para quitar de la cola")
	}
	nextThread := c.elements[0]
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

// TODO: Bonito pero quizás no está bueno que cualquiera pueda hacer cualquier cosa con la cola
func (c *Queue[T]) Do(f func(T)) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for i := range c.elements {
		f(c.elements[i])
	}
}

func (c *Queue[T]) Remove(t T) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for i := range c.elements {
		if c.elements[i].Equal(t) { // Comparación de punteros directamente
			c.elements = append(c.elements[:i], c.elements[i+1:]...)
			return nil
		}
	}
	return errors.New("elemento no encontrado en la cola")
}
