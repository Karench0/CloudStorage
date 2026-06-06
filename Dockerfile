FROM golang:1.25.8-alpine AS builder

WORKDIR /build


COPY go.mod go.sum ./
RUN go mod download


COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /main ./main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /main .
COPY --from=builder /build/templates ./templates

EXPOSE 9091

CMD ["./main"]