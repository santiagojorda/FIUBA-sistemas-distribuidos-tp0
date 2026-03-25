## Ejercicio N°4:
Modificar servidor y cliente para que ambos sistemas terminen de forma _graceful_ al recibir la signal SIGTERM. Terminar la aplicación de forma _graceful_ implica que todos los _file descriptors_ (entre los que se encuentran archivos, sockets, threads y procesos) deben cerrarse correctamente antes que el thread de la aplicación principal muera. Loguear mensajes en el cierre de cada recurso (hint: Verificar que hace el flag `-t` utilizado en el comando `docker compose down`).

---

## Solucion:
En un principio los entrypoints estaban seteados como `/bin/sh/`, la shell no maneja señales de forma adecuada, recibe el SIGTERM pero no lo procesa correctamente, no se lo comunica a los procesos hijos.A l Cambiar los entrypoints del cliente y el servidor para que cuando les llegue la señal de terminación, los coloco como procesos principales, PID 1, por ende pueden escuchar y procesar los SIGTERM, cierren los recursos y terminen de forma _graceful_. Para esto, se utiliza el paquete `os/signal` de Go para capturar la señal SIGTERM y ejecutar una función de limpieza antes de salir. Sino, al recibir la señal, el proceso se termina abruptamente sin cerrar los recursos correctamente.

Modularice el cliente para manejar las conexiones y el loop de mensajes en funciones separadas, cree el archivo `client/connection.go` para esto. Luego implemente el manejo de la señal SIGTERM utilizando el paquete `os/signal` de Go, asegurándome de cerrar correctamente los recursos antes de que la aplicación termine.

```go
# Creacion de señal para manejar el cierre graceful
signal.Notify(signalChan, syscall.SIGTERM)
```

Dentro del bucle principal del cliente, agrego un bloque `select` para escuchar la señal de terminación y ejecutar el proceso de cierre _graceful_:
Esto lo que hace es monitorear constantemente el canal `c.shutdown_event`.
```go
# Manejo de la señal SIGTERM para un cierre graceful
select {
case <-c.shutdown_event:
  log.Infof("action: graceful_shutdown | result: in_progress | client_id: %v", c.config.ID)
  if c.conn != nil {
    c.conn.Close()
  }
  log.Infof("action: graceful_shutdown | result: success | client_id: %v", c.config.ID)
  return
default:
}
```

Y en el servidor es algo parecido, se utiliza la libreria `signal` para capturar la señal SIGTERM y ejecutar una función de limpieza antes de salir. Se asegura de cerrar correctamente los recursos antes de que la aplicación termine.


creo un `self.__shutdown_event` para manejar el evento de cierre y un método `__handle_graceful_shutdown` para cerrar los recursos correctamente al recibir la señal SIGTERM. Luego, registro el manejador de señales para SIGTERM utilizando `signal.signal`.

```python
signal.signal(signal.SIGTERM, self.__handle_graceful_shutdown)
```

y en el loop principal del servidor chequeo constantemente el evento de cierre `self._shutdown_event` para determinar si se debe cerrar el servidor de forma _graceful_:
```python
while not self._shutdown_event.is_set():
    client_sock = self.__accept_new_connection()
    if client_sock is None:
        continue
    self.__handle_client_connection(client_sock)
try:
    self._server_socket.close()
except OSError:
    pass
```

