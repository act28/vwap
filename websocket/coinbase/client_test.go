package coinbase_test

import (
	"context"
	"testing"
	"time"

	"github.com/act28/vwap/websocket"
	"github.com/act28/vwap/websocket/coinbase"
	"github.com/stretchr/testify/require"
)

func TestNewWebsocketClient(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	_, err := coinbase.NewClient(ctx, coinbase.SandboxWSURL)
	require.NoError(t, err)
}

func TestWebsocketChannelSubscribe(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name         string
		tradingPairs []string
		wantErr      bool
	}{
		{
			name:         "EmptyTradingPairs",
			tradingPairs: []string{},
			wantErr:      true,
		},
		{
			name: "ValidTradingPairs",
			tradingPairs: []string{
				"BTC-USD",
				"ETH-USD",
				"ETH-BTC",
			},
			wantErr: false,
		},
		{
			name: "InvalidTradingPairs",
			tradingPairs: []string{
				"xxx-USD",
				"BTC-xxx",
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			// we have to use the production websocket endpoint because the
			// sandbox only supports BTC-USD.
			ws, err := coinbase.NewClient(ctx, coinbase.ProdWSURL)
			require.NoError(t, err)

			receiver := make(chan websocket.DataPoint)
			err = ws.Subscribe(ctx, tc.tradingPairs, receiver)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				var cnt int
				for m := range receiver {
					select {
					case <-time.After(3 * time.Second):
						cancel()
						return
					case <-ctx.Done():
					default:
						cnt++
						require.Contains(t, tc.tradingPairs, m.Pair)
						if cnt >= 3 {
							cancel()
							return
						}
					}
				}
			}
		})
	}
}
