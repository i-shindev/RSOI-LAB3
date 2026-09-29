# Gateway

Единая точка входа, порт **8080**. Своей базы нет.
Единственный сервис, который обращается к Cars, Rental и Payment: проводит бронирование, завершение и отмену
и собирает ответы из трёх источников.

Контракт — [spec.json](spec.json): это [OpenAPI задания](../v3/%5Binst%5D%5Bv3%5D%20Car%20Rental%20System.yml)
с тремя отличиями — `page` начинается с 1, отменённый платёж имеет статус `CANCELED` (а не `REVERSED`),
добавлен `/manage/health`.

| Метод | Путь | Ответ |
|---|---|---|
| GET | `/manage/health` | 200 |
| GET | `/api/v1/cars?page=&size=&showAll=` | 200 `PaginationResponse` |
| GET | `/api/v1/rental` | 200 `RentalResponse[]` |
| POST | `/api/v1/rental` `{carUid, dateFrom, dateTo}` | 200 `CreateRentalResponse` · 400 |
| GET | `/api/v1/rental/{rentalUid}` | 200 `RentalResponse` · 404 |
| DELETE | `/api/v1/rental/{rentalUid}` | 204 · 404 |
| POST | `/api/v1/rental/{rentalUid}/finish` | 204 · 404 |

Запросы к `/api/v1/rental` требуют заголовок `X-User-Name`.

```
RentalResponse       { rentalUid, status, dateFrom, dateTo, car: CarInfo, payment: PaymentInfo }
CreateRentalResponse { rentalUid, status, carUid, dateFrom, dateTo, payment: PaymentInfo }
CarInfo              { carUid, brand, model, registrationNumber }
PaymentInfo          { paymentUid, status, price }
```

- **Бронь**: зарезервировать автомобиль → создать платёж на (`dateTo − dateFrom`) дней × цена за сутки →
  создать аренду. Если шаг падает, уже сделанное откатывается: снимается резерв, отменяется платёж.
- **Завершение**: аренда → `FINISHED`, резерв снимается. Платёж не меняется.
- **Отмена**: аренда → `CANCELED`, резерв снимается, платёж → `CANCELED`.
- Ответы, которых нет в спеке задания: 400 — нет `X-User-Name`, автомобиль не найден или занят, завершение
  отменённой / отмена завершённой аренды; 503 — один из сервисов недоступен.

## Запуск

| Переменная | По умолчанию |
|---|---|
| `PORT` | `8080` |
| `CARS_SERVICE_URL` | `http://localhost:8070` |
| `RENTAL_SERVICE_URL` | `http://localhost:8060` |
| `PAYMENT_SERVICE_URL` | `http://localhost:8050` |
| `REQUEST_TIMEOUT` | `5s` |

```bash
go test ./...
go run ./src
```
