# Rental Service

Аренды пользователей. Порт **8060**, база `rentals`, таблица `rental`.
Хранит только UUID автомобиля и платежа.

Контракт — [spec.json](spec.json). Во всех запросах к `/api/v1/rentals` обязателен заголовок `X-User-Name`.

| Метод | Путь | Ответ |
|---|---|---|
| GET | `/manage/health` | 200 |
| GET | `/api/v1/rentals` | 200 `RentalResponse[]`; 400 |
| POST | `/api/v1/rentals` `{carUid, paymentUid, dateFrom, dateTo}` | 201 `RentalResponse`, заголовок `Location`; 400 |
| GET | `/api/v1/rentals/{rentalUid}` | 200 `RentalResponse`; 400; 404 |
| DELETE | `/api/v1/rentals/{rentalUid}` | 204; 400; 404; 409 |
| POST | `/api/v1/rentals/{rentalUid}/finish` | 204; 400; 404; 409 |

```
RentalResponse { rentalUid, username, carUid, paymentUid, dateFrom, dateTo, status: IN_PROGRESS | FINISHED | CANCELED }
```

- Даты в формате `YYYY-MM-DD`, `dateFrom` раньше `dateTo`. Новая аренда создаётся в статусе `IN_PROGRESS`.
- `Location` — HTTP-заголовок ответа `201 Created` с адресом созданной аренды,
  например `Location: /api/v1/rentals/4fd4fc0c-7840-483c-bcf5-3e2be7d4ea69`.
- Чужая аренда неотличима от несуществующей, вовзращается 404.
- `DELETE` переводит аренду в `CANCELED`, `finish` — в `FINISHED`. Менять статус можно только у аренды
  `IN_PROGRESS`. Повторный вызов того же действия отвечает 204, попытка сменить закрытую аренду на другой
  статус — 409.

## Запуск

| Переменная | По умолчанию |
|---|---|
| `PORT` | `8060` |
| `DATABASE_URL` | — (обязательна), например `postgres://program:test@localhost:5432/rentals` |

```bash
go test ./...
go run ./src
```
