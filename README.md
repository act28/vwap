# Real-time VWAP calculation engine

A real-time VWAP calculation engine using Coinbase's Websocket data feed.

## Design

This engine subscribes to the Coinbase Websocket `matches` channel and retrieves
data for the following trading pairs:

- BTC-USD
- ETH-USD
- ETH-BTC

It then calculates the VWAP per trading pair using a sliding window of 200 data
points.

The VWAP formula used is based on the following resources:

- https://academy.binance.com/en/articles/volume-weighted-average-price-vwap-explained
- https://en.wikipedia.org/wiki/Volume-weighted_average_price


## Assumptions

For the purposes of this exercise, this engine takes a YAGNI approach.
Therefore, the following assumptions have been made:

1. No client authentication. The engine uses the publicly available
websocket feed, and takes rate limiting into consideration. Client
authentication can be trivially added if required.

2. The engine only handles the `matches` channel. This could be further extended to
handle the other channel types in the Coinbase websocket client.

3. Only the Coinbase websocket feed is supported. The packages may need to be
structured differently to handle other exchanges, and client types.

## Usage

To run the engine:

```Shell
make run
```
