package memoriaTypes

type EstrategiasAsignacionInterface interface {
	BuscarParticion(int) (error, Particion) // Recibe tamanio de proceso y devuelve la particion a la que se le asigno
}
