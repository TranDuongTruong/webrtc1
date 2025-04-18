# Sử dụng image chính thức của Golang
FROM golang:1.21-alpine

# Thiết lập thư mục làm việc trong container
WORKDIR /app

# Sao chép các file Go vào thư mục làm việc của container
COPY . .

# Chuyển đến thư mục chứa main.go
WORKDIR /app/examples/whip-whep

# Cài đặt các phụ thuộc và build ứng dụng Go
RUN go mod tidy
RUN go build -o server .

# Mở cổng 8083 cho server 2
EXPOSE 8084

# Lệnh chạy khi container được khởi động
CMD ["./server", "--port", "8084"]
