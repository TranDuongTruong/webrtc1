# Dockerfile
FROM golang:1.21

WORKDIR /app

COPY . .

WORKDIR /app/examples/whip-whep

RUN go build -o server main.go

EXPOSE 8084

CMD ["./server"]
