# Payment Service

Платежи. Порт **8050**, база `payments`, таблица `payment`.

Контракт — [spec.json](spec.json).

| Метод | Путь | Ответ |
|---|---|---|
| GET | `/manage/health` | 200 |
| POST | `/api/v1/payments` `{price}` | 201 `PaymentResponse`, заголовок `Location`; 400 |
| GET | `/api/v1/payments/{paymentUid}` | 200 `PaymentResponse`; 400; 404 |
| DELETE | `/api/v1/payments/{paymentUid}` | 204; 400; 404 |

```
PaymentResponse { paymentUid, status: PAID | CANCELED, price }
```

- Новый платёж создаётся в статусе `PAID`.
- `DELETE` не удаляет запись, а переводит платёж в `CANCELED`. Повторный вызов тоже отвечает 204.
- `price` обязательное неотрицательное число.
- `Location` — HTTP-заголовок ответа `201 Created` с адресом созданного платежа,
  например `Location: /api/v1/payments/f0ca3bcb-98a4-40c0-8163-1a6b6f40e536`.
## Запуск

| Переменная | По умолчанию |
|---|---|
| `PORT` | `8050` |
| `DATABASE_URL` | — (обязательна), например `postgres://program:test@localhost:5432/payments` |

```bash
go test ./...
go run ./src
```
