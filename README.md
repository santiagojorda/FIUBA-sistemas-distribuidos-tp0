## Ejercicio N°1:
> Definir un script de bash `generar-compose.sh` que permita crear una definición de Docker Compose con una cantidad configurable de clientes.  El nombre de los containers deberá seguir el formato propuesto: client1, client2, client3, etc. 
>
> El script deberá ubicarse en la raíz del proyecto y recibirá por parámetro el nombre del archivo de salida y la cantidad de clientes esperados:
> 
> `./generar-compose.sh docker-compose-dev.yaml 5`
> 
> Considerar que en el contenido del script pueden invocar un subscript de Go o Python:
> 
>```
>#!/bin/bash
>echo "Nombre del archivo de salida: $1"
>echo "Cantidad de clientes: $2"
>python3 mi-generador.py $1 $2
>```
>
>En el archivo de Docker Compose de salida se pueden definir volúmenes, variables de entorno y redes con libertad, pero recordar actualizar este script cuando se modifiquen tales definiciones en los sucesivos ejercicios.

---

## Resolucion
Se creo un script llamada `generar-compose.sh` el cual se encarga de generar un archivo `docker-compose-dev.yaml` con la cantidad de clientes que se le indique por parámetro. Para ejecutarlo se tiene que pasar el nombre del archivo de salida y la cantidad de clientes esperados, por ejemplo:

```bash
./generar-compose.sh docker-compose-dev.yaml 5
```

En el se definen variables de entorno, y redes para la comunicacion entre clientes y el servidor. 
Cada cliente va a tener un `ID` unico.