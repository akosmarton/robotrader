FROM golang:1.23 as builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v

FROM alpine

COPY --from=builder /app/go-trader /app/go-trader
ENV STORAGE_DIR=/data ALPACA_API_KEY= ALPACA_API_SECRET= MATRIX_HOMESERVER= MATRIX_USER_ID= MATRIX_ACCESS_TOKEN= MATRIX_ROOM_ID=
VOLUME [ "/data" ]
ENTRYPOINT [ "/app/go-trader" ]