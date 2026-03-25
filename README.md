## Ejercicio N°2:
> Modificar el cliente y el servidor para lograr que realizar cambios en el archivo de configuración no requiera reconstruír las imágenes de Docker para que los mismos sean efectivos. La configuración a través del archivo correspondiente (`config.ini` y `config.yaml`, dependiendo de la aplicación) debe ser inyectada en el container y persistida por fuera de la imagen (hint: `docker volumes`).

---

## Solucion:
Se vinculo el archivo de configuracion `config.ini` y `config.yaml` a los contenedores utilizando `docker volumes`. De esta manera, cualquier cambio realizado en los archivos de configuración fuera del contenedor se refla automáticamente dentro del contenedor sin necesidad de reconstruir las imágenes de Docker.

Se modifico el Dockerfile para eliminar la copia de los archivos de configuración dentro de la imagen, y en su lugar, se montaron los archivos de configuración como volúmenes en el contenedor 

Esto nos da flexibilidad para modificar la configuración de las aplicaciones cliente y servidor sin tener que pasar por el proceso de construcción de las imágenes, lo que ahorra tiempo y facilita la gestión de la configuración.