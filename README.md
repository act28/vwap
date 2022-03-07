# Real-time VWAP calculation engine

A real-time VWAP calculation engine using Coinbase's Websocket data feed.

## Design

1. This engine subscribes to the Coinbase Websocket `matches` channel and
   retrieves data for the following trading pairs:
   - BTC-USD
   - ETH-USD
   - ETH-BTC

2. It then calculates the VWAP per trading pair using a sliding window of 200
   (maximum) data points.

3. The VWAP formula used is based on the following resources:
   - <https://academy.binance.com/en/articles/volume-weighted-average-price-vwap-explained>
   - <https://en.wikipedia.org/wiki/Volume-weighted_average_price>

4. The sliding window is implemented as a slice of data points. When a new data
   point is pushed into the window, we check if the window size has been
   exceeded, and pop the oldest data point from the queue.

5. Rather than recalculate the `Product * Quantity` and `Quantity` summations
   each time a new data point is added, the `PQ` and `Q` of the oldest data
   point is first deducted from a running total in O(N) constant time and O(1)
   space.

## Assumptions

For the purposes of this exercise, this engine takes the KISS+YAGNI approach.
Therefore, the following assumptions have been made:

1. No client authentication. The engine uses the publicly available websocket
   feed, and takes rate limiting into consideration. Client authentication can be
   trivially added if required.

2. The engine only handles the `matches` channel. This could be further extended
   to handle the other channel types in the Coinbase websocket client.

3. Only the Coinbase websocket feed is supported. The packages may need to be
   structured differently to handle other exchanges, and client types.

4. Under normal circumstances, the data feed service would be separate from the
   vwap calculation engine. I have opted to combine it in a single service due
   to time constraints.

## Usage

To run the engine...

### As a docker container

```Shell
make run
```

This builds the docker image and runs the docker container.

### As a CLI command

```Shell
make run-cli
```

This runs the command `go run ./cmd/server/main.go`

### Testing

To run the tests:

```Shell
make test
```

Uniting test only. No integration test.

## TODO

Things to do if I had more time:

1. Improve test coverage
2. Split the websocket data feed and VWAP calculation engine into two separate
   service implementions. We use the [Dogma](https://github.com/dogmatiq/dogma) framework to build event-sourcing services like this.
3. Handle websocket timeouts and disconnects.

## Issues

1. The sandbox endpoint operates differently to the live endpoint. Firstly, it
   only accepts `BTC-USD`. All other `product_id` results in a `Failed to
   subscribe` error.
2. Secondly, it sends a single `last_match` response and nothing else for an
   extended period. It intermittently sends `match` responses in short bursts,
   again at extended intervals.
