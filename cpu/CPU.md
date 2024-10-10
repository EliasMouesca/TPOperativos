# CPU Documentation
Una lista de los endpoints expuestos y cómo usarlos.

## Execute
Pone a ejecutar el proceso indicado. Si la CPU está ocupada, queda esperando.
- Endpoint: /execute
- Verbo: POST
- Query params: tid, pid

## Interrupt
- Endpoint: /interrupt
- Verbo: POST
- Body: Estructura types.Interruption{} en formato json
