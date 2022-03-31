FROM golang:latest

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

ENV COINBASE_SOCKET_URL wss://ws-feed.exchange.coinbase.com

RUN go build

CMD ["./coinbase-websocket-trading-pairs"]
