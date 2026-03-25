## Ejercicio 3:
>Crear un script de bash `validar-echo-server.sh` que permita verificar el correcto funcionamiento del servidor utilizando el comando `netcat` para interactuar con el mismo. Dado que el servidor es un echo server, se debe enviar un mensaje al servidor y esperar recibir el mismo mensaje enviado.
>
>En caso de que la validación sea exitosa imprimir: `action: test_echo_server | result: success`, de lo contrario imprimir:`action: test_echo_server | result: fail`.
>
>El script deberá ubicarse en la raíz del proyecto. Netcat no debe ser instalado en la máquina _host_ y no se pueden exponer puertos del servidor para realizar la comunicación (hint: `docker network`). `

---

## Solución:
Se desarrollo el script `validar-echo-server.sh` que automatiza la verificacion del servidor siguiendo las restricciones mencionadas. 

Se utiliza busybox para crear un contenedor temporal, este se conecta con el contenedor del echo server a través de una red de Docker, y se utiliza netcat para enviar un mensaje al servidor y verificar la respuesta. 
No expongo puertos, ya que la comunicación se realiza internamente entre los contenedores, en la red privada de Docker