FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY src ./src
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/payment-service ./src

FROM alpine:3.20
RUN adduser -D -H -s /sbin/nologin payment
COPY --from=build /out/payment-service /usr/local/bin/payment-service
USER payment
ENV PORT=8050
EXPOSE 8050
ENTRYPOINT ["payment-service"]
