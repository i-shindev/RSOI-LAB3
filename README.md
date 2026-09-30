# Лабораторная работа #3

![GitHub Classroom Workflow](../../workflows/GitHub%20Classroom%20Workflow/badge.svg?branch=master)

## Fault Tolerance

### Формулировка

На базе [Лабораторной работы #2](https://github.com/bmstu-rsoi/lab2-template) реализовать механизмы, увеличивающие
отказоустойчивость системы.

### Требования

1. На Gateway Service для _всех операций_ чтения реализовать паттерн Circuit Breaker. Накапливать статистику в памяти, и
   если система не ответила N раз, то в N + 1 раз вместо запроса сразу отдавать fallback. Через небольшой timeout
   выполнить запрос к реальной системе, чтобы проверить ее состояние.
2. На каждом сервисе сделать специальный endpoint `GET /manage/health`, отдающий 200 ОК, он будет использоваться для
   проверки доступности сервиса (в [Github Actions](.github/workflows/classroom.yml) в скрипте проверки готовности всех
   сервисов [wait-script.sh](scripts/wait-script.sh) и в тестах [test-script.sh](scripts/test-script.sh)).
   ```shell
   "$path"/wait-for.sh -t 120 "http://localhost:$port/manage/health" -- echo "Host localhost:$port is active"
   ```
4. В случае недоступности данных из некритичного источника (не основного), возвращается fallback-ответ. В зависимости от
   ситуации, это может быть:
    * пустой объект или массив;
    * объект, с заполненным полем (`uid` или подобным), по которому идет связь с другой системой;
    * default строка (если при этом не меняется тип переменной).
5. В задании описаны две операции, изменяющие состояния нескольких систем. В случае недоступности одной из систем,
   участвующих в этой операции, выполнить:
    1. откат всей операции;
    2. возвращать пользователю ответ об успешном завершении операции, а на Gateway Service поставить этот запрос в
       очередь для повторного выполнения.
6. Для автоматических прогонов тестов в файле [autograding.json](.github/classroom/autograding.json)
   и [classroom.yml](.github/workflows/classroom.yml) заменить `<variant>` на ваш вариант.
7. В [docker-compose.yml](docker-compose.yml) прописать сборку и запуск docker контейнеров.
8. Код хранить на Github, для сборки использовать Github Actions.
9. Каждый сервис должен быть завернут в docker.
10. В classroom.yml дописать шаги на сборку, прогон unit-тестов.

### Пояснения

1. Для локальной разработки можно использовать Postgres в docker.
2. Схема взаимодействия сервисов остается как в [Лабораторной работы #2](https://github.com/bmstu-rsoi/lab2-template).
3. Для реализации очереди можно использовать language native реализацию (например, BlockingQueue для Java), либо
   какую-то готовую реализацию типа RabbitMQ, Redis, ZeroMQ и т.п. Крайне нежелательно использовать реляционную базу
   данных как средство эмуляции очереди.
4. Можно использовать внешнюю очередь или запускать ее в docker.
5. Для проверки отказоустойчивости используется остановка и запуск контейнеров docker, это делает
   скрипт [test-script.sh](scripts/test-script.sh). Скрипт нужно запускать из корня проекта, т.к. он обращается к папке
   postman по вариантам.
   ```shell
   # запуск тестового сценария:
   # * <variant> – номер варианта (v1 | v2 | v3 | v4 )
   # * <service> – имя сервиса в Docker Compose
   # * <port>    – порт, на котором запущен сервис
   $ scripts/test-script.sh <variant> <service> <port>
   ```

### Прием задания

1. При получении задания у вас создается fork этого репозитория для вашего пользователя.
2. После того как все тесты успешно завершатся, в Github Classroom на Dashboard будет отмечено успешное выполнение
   тестов.

### Варианты заданий

Распределение вариантов заданий аналогично [ЛР #2](https://github.com/bmstu-rsoi/lab2-template).

1. [Flight Booking System](v1/README.md)
2. [Hotels Booking System](v2/README.md)
3. [Car Rental System](v3/README.md)
4. [Library System](v4/README.md)

## Реализация

Вариант 3 — [Car Rental System](v3/README.md). Все сервисы написаны на Go 1.25
(стандартный `net/http`, `pgx/v5` для Postgres).

| Сервис | Порт | База | Описание |
|---|---|---|---|
| [gateway](gateway/README.md) | 8080 | — | единая точка входа, оркестрация и сборка ответов |
| [cars-service](cars-service/README.md) | 8070 | `cars` | автомобили, пагинация, резервирование |
| [rental-service](rental-service/README.md) | 8060 | `rentals` | аренды пользователя и их статусы |
| [payment-service](payment-service/README.md) | 8050 | `payments` | платежи |

У каждого сервиса есть `spec.json` (OpenAPI 3) и `GET /manage/health`.
Сервисы друг к другу не обращаются: всё, что требует данных двух сервисов, делает Gateway.
Каждый сервис работает только со своей базой; схема и тестовые данные создаются init-скриптом
[db-v3.sql](postgres/scripts/db-v3.sql).

Внутри каждого сервиса одни и те же слои (`src/`):

| Пакет | Что |
|---|---|
| `domain` | сущности, валидация, ошибки, интерфейс репозитория (у Gateway — интерфейсы клиентов сервисов) |
| `service` | бизнес-логика и unit-тесты на mock-репозитории |
| `repository` | реализация репозитория на Postgres (у Gateway — `clients`, HTTP-клиенты сервисов) |
| `api` | HTTP-хендлеры, DTO, перевод ошибок в коды ответа |

### Отказоустойчивость

Всё реализовано только на Gateway (`gateway/src`):

* **Circuit Breaker** (`clients/circuit_breaker.go`) — у каждого HTTP-клиента свой. Считает отказы (нет соединения,
  таймаут, 5xx) в скользящем окне из 5 запросов; при 3 отказах открывается. Через 3 секунды переходит в half-open.
* **Fallback** (`service/gateway.go`) — если при `GET /api/v1/rental` и
  `GET /api/v1/rental/{rentalUid}` недоступен Cars Service или Payment Service, вместо ошибки отдаются данные аренды,
  а `car`/`payment` заполняются заглушкой.
* **Откат** — бронирование автомобиля при отказе Payment Service или Rental Service отменяет уже выполненные шаги и возвращает 503
* **Очередь повторов** (`service/retry_queue.go`) — отмена и завершение аренды после успешного шага в Rental Service
  возвращают 204, неудавшиеся запросы к Cars Service и Payment Service ставятся в очередь и повторяются раз в
  секунду, пока сервис не ответит. Туда же попадают неудавшиеся шаги отката при бронировании.

### Запуск

```bash
docker compose up -d --build
docker compose --profile test run --rm newman
docker compose down
```

Второй командой прогоняется папка `success` коллекции `v3/postman` по Gateway внутри сети compose.
Полный сценарий с остановкой и запуском Payment Service (нужны `newman` и `wget`):

```bash
./scripts/test-script.sh v3 payment-service 8050
```

После изменения `postgres/scripts/db-v3.sql` нужен `docker compose down -v`: init-скрипт выполняется только
на пустом томе.

### Unit-тесты

```bash
for s in gateway cars-service rental-service payment-service; do (cd $s && go test ./...); done
```

В CI (`.github/workflows/classroom.yml`) для каждого сервиса выполняются `go vet`, `go build` и `go test`,
затем образы собираются через compose, и `scripts/test-script.sh` прогоняет сценарий с остановкой Payment Service.
