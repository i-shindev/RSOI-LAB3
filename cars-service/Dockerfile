FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY src ./src
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/cars-service ./src

FROM alpine:3.20
RUN adduser -D -H -s /sbin/nologin cars
COPY --from=build /out/cars-service /usr/local/bin/cars-service
USER cars
ENV PORT=8070
EXPOSE 8070
ENTRYPOINT ["cars-service"]
