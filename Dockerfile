FROM docker.io/library/node:slim as node_builder
WORKDIR /app/
COPY . .
RUN npm install && npm run build

FROM golang:1.23.7 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v

FROM alpine
WORKDIR /app
COPY --from=node_builder /app/dist /app/dist
COPY --from=builder /app/robotrader /app/robotrader
ENV STORAGE_DIR=/data ALPACA_API_KEY= ALPACA_API_SECRET= MATRIX_HOMESERVER= MATRIX_USER_ID= MATRIX_ACCESS_TOKEN= MATRIX_ROOM_ID=
VOLUME [ "/data" ]
ENTRYPOINT [ "/app/robotrader" ]