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
	// Subscribe subscribes to a data feed and receives a stream of data points
	// for the specified trading pairs in the `receiver` channel.
	Subscribe(ctx context.Context, tradingPairs []string, receiver chan<- DataPoint) error
}
