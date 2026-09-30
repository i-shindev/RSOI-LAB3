FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY src ./src
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/gateway ./src

FROM alpine:3.20
RUN adduser -D -H -s /sbin/nologin gateway
COPY --from=build /out/gateway /usr/local/bin/gateway
USER gateway
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["gateway"]
