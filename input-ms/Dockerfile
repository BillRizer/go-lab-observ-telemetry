FROM golang:1.21 as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/microservice/main.go

FROM scratch

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"] 