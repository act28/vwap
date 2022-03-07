package websocket

import (
	"context"

	"github.com/shopspring/decimal"
)

// DataPoint represents a trading data point.
type DataPoint struct {
	Pair   string
	Volume decimal.Decimal
	Price  decimal.Decimal
}

// Client is an interface for implementing websocket clients that receive a
// stream of data points for a trading pair.
type Client interface {
	// Subscribe subscribes to a data feed, and returns an error, or nil.
	Subscribe(ctx context.Context, tradingPairs []string) error
	// Receive listens on the data channel, and sends datapoints to the
	// receiver.
	Receive(ctx context.Context, receiver chan<- DataPoint)
}
