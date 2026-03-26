## Ejercicio N°7:

>Modificar los clientes para que notifiquen al servidor al finalizar con el envío de todas las apuestas y así proceder con el sorteo.
>Inmediatamente después de la notificacion, los clientes consultarán la lista de ganadores del sorteo correspondientes a su agencia.
>Una vez el cliente obtenga los resultados, deberá imprimir por log: `action: consulta_ganadores | result: success | cant_ganadores: ${CANT}`.
>
>El servidor deberá esperar la notificación de las 5 agencias para considerar que se realizó el sorteo e imprimir por log: `action: sorteo | result: success`.
>Luego de este evento, podrá verificar cada apuesta con las funciones `load_bets(...)` y `has_won(...)` y retornar los DNI de los ganadores de la agencia en cuestión. Antes del sorteo no se podrán responder consultas por la lista de ganadores con información parcial.
>
>Las funciones `load_bets(...)` y `has_won(...)` son provistas por la cátedra y no podrán ser modificadas por el alumno.
>
>No es correcto realizar un broadcast de todos los ganadores hacia todas las agencias, se espera que se informen los DNIs ganadores que correspondan a cada una de ellas.

---

## Solucion:

Se incorporaron dos nuevos mensajes, `ASK_WINNERS` y `WAIT`, para permitir que los clientes consulten por los ganadores y que el servidor gestione la sincronización del proceso.

El servidor espera a que las cinco agencias finalicen el envío de sus apuestas. Una vez cumplida esta condición, se realiza el sorteo. A partir de ese momento, cuando un cliente consulta por los ganadores, el servidor verifica si el sorteo ya fue efectuado y responde con la cantidad de ganadores correspondiente a la agencia del cliente.

Si un cliente consulta antes de que el sorteo haya sido realizado, envía un mensaje `ASK_WINNERS` y el servidor responde con `WAIT`, indicando que debe aguardar antes de volver a intentar. Dado que el servidor procesa los mensajes de forma secuencial (sin mecanismos de concurrencia), el cliente implementa una espera activa: tras recibir `WAIT`, aguarda un intervalo de tiempo y vuelve a realizar la consulta.

Una vez realizado el sorteo, el servidor responde a cada cliente con los resultados de ganadores de su agencia, y cada cliente registra en su log la cantidad de ganadores obtenida.