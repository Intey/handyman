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

# Установка rpm
```bash
rpm -Uihv handyman-0.1.0-1-any.rpm
```

# Удаление rpm
Пусть мы собрали rpm, который имеет следующий вид
```bash
handyman-0.1.0-1-any.rpm
```

*ВНИМАНИЕ*: для удаления пакета нужно использовать только ту часть, которая относится к названию пакета
```bash
rpm -e handyman-0.1.0-1
```

При попытке удалить пакет по названию файла (с суффиксом any.rpm) будет ошибка
