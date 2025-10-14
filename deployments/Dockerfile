FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest && swag init -g cmd/subs/main.go -o docs
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o main ./cmd/subs

FROM alpine:latest
ENV GIN_MODE=release

RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/docs ./docs
EXPOSE 8080
CMD ["./main"]