FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY src ./src
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/rental-service ./src

FROM alpine:3.20
RUN adduser -D -H -s /sbin/nologin rental
COPY --from=build /out/rental-service /usr/local/bin/rental-service
USER rental
ENV PORT=8060
EXPOSE 8060
ENTRYPOINT ["rental-service"]
