package types

type RequestToDumpMemory struct {
	Contenido []byte
	Nombre    string
	Size      int
}
