# Memory dump
Esto se lo pide kernel a memoria y memoria a filesystem, esta es la especificación de la comunicación memoria -> fs
- Endpoint: /memoryDump
- Queryparams: pid, tid, size (en bytes)
- Data: contenido actual de la memoria ([]byte)

