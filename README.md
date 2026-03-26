## Ejercicio N°8:

> Modificar el servidor para que permita aceptar conexiones y procesar mensajes en paralelo. En caso de que el alumno implemente el servidor en Python utilizando _multithreading_,  deberán tenerse en cuenta las [limitaciones propias del lenguaje](https://wiki.python.org/moin/GlobalInterpreterLock).

---

## Solucion:

Agregué multithreading al servidor para permitir que múltiples clientes se atiendan en paralelo. Para esto, utilizo el módulo threading de Python, creando un nuevo hilo por cada conexión entrante. De esta forma, el servidor puede seguir aceptando conexiones mientras otros clientes envían apuestas o consultan resultados, evitando bloqueos y mejorando la capacidad de respuesta.

Para evitar problemas de concurrencia, incorporé mecanismos de sincronización:

Utilizo un mutex (`_bets_file_lock`) para proteger el acceso al archivo de apuestas. Esto evita condiciones de carrera cuando varios threads intentan escribir simultáneamente. Este lock se aplica al momento de persistir las apuestas (`store_bets`), garantizando consistencia en los datos.

Además, implemento un mutex (`_state_lock`) junto con una variable de condición (`_sorteo_condition`) para coordinar la ejecución del sorteo. Esto permite que los threads esperen de manera eficiente hasta que todas las agencias hayan terminado de enviar sus apuestas.

Cada cliente envía sus apuestas y, al finalizar, notifica al servidor con un mensaje de `FIN`.
El servidor registra qué agencias terminaron.
Cuando la última agencia finaliza, se realiza el sorteo una única vez.
Luego, se utiliza `notify_all()` para despertar a los threads que estaban esperando el resultado.
Estos threads pueden entonces consultar y obtener la cantidad de ganadores correspondiente.

```python
def get_winners_count(self, agency_id):
    agency = int(agency_id)

    with self._sorteo_condition:
        while not self._sorteo_done and not self._shutdown_event.is_set():
            self._sorteo_condition.wait(timeout=SECONDS_TO_WAIT_FOR_SORTEO)

        if not self._sorteo_done:
            return None

        return self._winners_by_agency.get(agency, 0)
```



