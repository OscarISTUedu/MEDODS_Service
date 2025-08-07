# Тестовое задание на позицию Junior Backend Developer

![Логотип проекта](https://github.com/OscarISTUedu/Images-for-MEDODS/blob/main/medods_logo.png)

Проект состоит из 2 Docker контейнеров: Postgres16, Go1.24

- В качестве основного framwork используется Gin
- Реализовано автопродление access токена
- Отсутвие тегов latest гарантирует стабильность проекта

## Оглавление

- [Установка](#установка)
- [Env-файл](#env-файл)
- [Swagger-документация](#swagger-документация)

## Установка

Пошаговые инструкции по установке:

```bash
git clone https://github.com/OscarISTUedu/MEDODS_Service.git
cd docker
docker-compose -f docker-compose.yml up -d
```

## Env-файл
В связи с тем что github не позволяет загрузить .env файл в репозиторий, его потребуется создать (в папке docker), а затем заполнить.
Содержание .env файла
```bash
#Database
DB_HOST=db
DB_PORT=5432
DB_USER=admin
DB_PASSWORD=h1Dk+d2Gb@
DB_NAME=auth
#App
APP_PORT=8000
SECRET_KEY='13?*JfdK;JNA2+1-~_)+QCfhb'
```

## Swagger-документация
Для каждого endpoint была написана аннотация с описанием запроса. 
Просмотреть их можно предварительно развернув сервис и перейдя по ссылке: http://127.0.0.1:8000/swagger/index.html
Далее будут приведены фрагменты документации для каждого запроса

- Запрос на получение пары токенов (access и refresh) для пользователя с идентификатором (GUID) указанным в параметре запроса
![скриншот](https://github.com/OscarISTUedu/Images-for-MEDODS/blob/main/get_user.PNG)

- Запрос на обновление пары токенов
![скриншот](https://github.com/OscarISTUedu/Images-for-MEDODS/blob/main/update_tokens.PNG)

- Запрос на получение GUID текущего пользователя (роут защищен)
![скриншот](https://github.com/OscarISTUedu/Images-for-MEDODS/blob/main/get_id.PNG)

- Запрос на деавторизацию пользователя (поле выполнения этого запроса с access токеном, пользователю больше не доступен роут на получение его GUID и операция обновления токенов)
![скриншот](https://github.com/OscarISTUedu/Images-for-MEDODS/blob/main/update_tokens.PNG)
