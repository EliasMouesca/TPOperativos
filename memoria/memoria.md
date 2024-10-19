### Aclaraciones
    - Los outputs están separados por los direfentes grupos de respuestas que puede tener (no puede tener outputs del 1. y del 2. al mismo tiempo)
    - Todos los llamados corroboran que el verbo sea el adecuado, de no serlo responderá con un StatusMethodNotAllowed
    - En el listado de los Outputs se tiene encuenta que la peticiones siempre estarán bien formadas y que los errores ocurren por un error en el valor de los datos.
      Generalmente, si las peticiones no están bien formadas, se devuelve un BadRequest

# Conecciones con CPU
## Guardar(o actualizar) contexto de ejecución
Funcionamiento
    Guarda (creando uno nuevo o actualizando) el contexto de ejecución para el pid, tid especificado
verbo: POST
Input:
    Body: json con struct ExecutionContext
    Params: pid y tid (a los que pertenece el contexto)
Output:
    OK (si salió bien)

ejemplo de URL: ".../memoria/saveContext?pid=123&tid=123"


## Obtener contexto de ejecución
Funcionamiento:
    Devuelve json del ExecutionContext solicitado a partir de los parametros de la query
Verbo: GET
Input:
    Params: pid y tid
Output:
    1. ExcecutionContext
    2. StatusNotFound

ejemplo URL: ".../memoria/getContext?pid=123&tid=123"


## Obtener siguiente linea de pseudocódigo a ejecutar
Endpoint: _/getInstruction_
Verbo: _GET_
Params: _pid_, _tid_, _pc_
Output: Instrucción a ejecutar (string)

ie: "http://localhost:9999/memoria/getContext?pid=123&tid=123&pc=5"


### Funcionamiento:
    Devuelve un string (sin parsear) que corresponde al pc, del tid y pid (los tres pasados por parametros)
    Verbo: GET
### Input:
    Params: pid, tid y pc
### Output:
    1. string + StatusOK
    2. nil + StatusNotFound



## Leer memoria
Endpoint: /readMem
Funcionamiento:
    Devuelve los primeros 4 bytes (correspondientes a un espacio en memoria de usuario) a partir del byte envíado por dirección (física)
Verbo: GET
Input:
    Params: addr (= base + desplazamiento), tid, pid (estos dos últimos es para el log obligatorio)
Output:
    1. [4]byte
    2. StatusNotFound

## Escribir memoria
Endpoint: /writeMem
Funcionamiento:

Verbo: POST
Input:
    Params: idem "Leer memoria"
    Body: json de [4]byte (con los datos a guardar)
Output:
    1. OK
    2. StatusNotFound