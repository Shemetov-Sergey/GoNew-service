FROM golang:alpine as builder

WORKDIR /GoNew-service

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main ./cmd/

FROM alpine

WORKDIR /GoNew-service

COPY --from=builder /GoNew-service/main /GoNew-service/main
COPY --from=builder /GoNew-service/pkg/config/envs/*.env /GoNew-service/
COPY --from=builder /GoNew-service/config.json /GoNew-service/

RUN chmod +x /GoNew-service/main

CMD ["./main"]