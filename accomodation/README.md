# Сборка rpm
Здесь описаны шаги, выполненные для сборки rpm в контейнере. Необходим установленный докер.

### Билдим образ контейнера
```bash
cd accomodation/build_rpm
sh accomodation/build_rpm/docker_build_image.sh
```

### Запускаем сборку rpm
```bash
sh accomodation/build_rpm/docker_run.sh
```
В консоли будет происходить процесс сборки. В конце сборки будут строчки:
```bash
RPM is ready
```
После этого в папке accomodation/build_rpm появится rpm-пакет.
