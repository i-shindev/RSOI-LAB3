# Cars Service

Каталог автомобилей, изменение статуса доступности машины. Порт **8070**, база `cars`, таблица `cars`.

Контракт — [spec.json](spec.json).

| Метод | Путь | Ответ |
|---|---|---|
| GET | `/manage/health` | 200 |
| GET | `/api/v1/cars?page=&size=&showAll=` | 200 `PaginationResponse`; 400 |
| GET | `/api/v1/cars/{carUid}` | 200 `CarResponse`; 400; 404 |
| POST | `/api/v1/cars/{carUid}/reservation` | 204; 400; 404; 409 |
| DELETE | `/api/v1/cars/{carUid}/reservation` | 204; 400; 404 |

```
PaginationResponse { page, pageSize, totalElements, items: CarResponse[] }
CarResponse        { carUid, brand, model, registrationNumber, power, type, price, available }
```

- `page` начинается с 1 (по умолчанию 1), `size` от 1 до 100 (по умолчанию 10).
- Без `showAll=true` в выдаче только свободные автомобили. `pageSize` — число элементов на этой странице.
- Резерв атомарный: `availability` меняется с `true` на `false` одним `UPDATE`.
  Если автомобиль уже занят — 409.
- Снятие резерва со свободного автомобиля тоже отвечает 204.

## Запуск

| Переменная | По умолчанию |
|---|---|
| `PORT` | `8070` |
| `DATABASE_URL` | — (обязательна), например `postgres://program:test@localhost:5432/cars` |

```bash
go test ./...
go run ./src
```
