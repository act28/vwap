FROM golang:1.17-alpine

COPY artifacts/build/release/linux/amd64/* /app/bin/

ENV TRADING_PAIRS="BTC-USD ETH-USD ETH-BTC" \
    WEBSOCKET_ENDPOINT="wss://ws-feed.exchange.coinbase.com" \
    WINDOW_SIZE=200

ENTRYPOINT ["/app/bin/server"]
