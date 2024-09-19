# GoNew-service
Gonews microservice gRPC

### Описание:

Данный микросервис осуществляет чтение данные RSS из источников, указанных в файле конфигурации config.json

### Запуск через утилиту make

1) Для запуска локально использовать команду make server и сервис запуститься с локальным конфигом (dev.env)

2) Для запуска с продовым конфигом нужно создать файл prod.env в той же директории, что и dev.env -> pkg/config/envs.

Также, в случае запуска в контейнере нужно также создать в корне приложения файл .env и указать в нем переменные PORT и DEBUG, 
данные значения подтягиваются файлом docker-compose.yaml.

### Для запуска в контейнере через docker compose

Нужно запустить команду в корне проекта

Если версия docker и docker compose из последних:

1. sudo docker compose build
2. sudo docker compose up

Если версия docker и docker compose из установлена через apt:

1. sudo docker-compose build
2. sudo docker-compose up