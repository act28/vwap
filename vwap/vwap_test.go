package vwap_test

import (
	"context"
	"log"
	"testing"

	"github.com/act28/vwap/vwap"
	"github.com/act28/vwap/websocket"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestVWAPCalculation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name       string
		dataPoints []websocket.DataPoint
		want       []vwap.Result
		size       uint
	}{
		{
			name:       "EmptyDataPoints",
			dataPoints: []websocket.DataPoint{},
			want:       []vwap.Result{},
			size:       4,
		},
		{
			name: "InvalidWindowSize",
			dataPoints: []websocket.DataPoint{
				{Price: decimal.NewFromInt(10), Volume: decimal.NewFromInt(11), Pair: "BTC-USD"},
				{Price: decimal.NewFromInt(20), Volume: decimal.NewFromInt(23), Pair: "BTC-USD"},
				{Price: decimal.NewFromInt(8), Volume: decimal.NewFromInt(15), Pair: "BTC-USD"},
				{Price: decimal.NewFromInt(11), Volume: decimal.NewFromInt(165), Pair: "BTC-USD"},
			},
			want: []vwap.Result{
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("10")},
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("16.7647058823529412")},
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("14.0816326530612245")},
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("11.705607476635514")},
			},
			size: 1000,
		},
		{
			name: "SinglePairBigWindow",
			dataPoints: []websocket.DataPoint{
				{Price: decimal.NewFromInt(10), Volume: decimal.NewFromInt(11), Pair: "BTC-USD"},
				{Price: decimal.NewFromInt(20), Volume: decimal.NewFromInt(23), Pair: "BTC-USD"},
				{Price: decimal.NewFromInt(8), Volume: decimal.NewFromInt(15), Pair: "BTC-USD"},
				{Price: decimal.NewFromInt(11), Volume: decimal.NewFromInt(165), Pair: "BTC-USD"},
			},
			want: []vwap.Result{
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("10")},
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("16.7647058823529412")},
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("14.0816326530612245")},
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("11.705607476635514")},
			},
			size: 5,
		},
		{
			name: "SinglePairSmallWindow",
			dataPoints: []websocket.DataPoint{
				{Price: decimal.NewFromInt(10), Volume: decimal.NewFromInt(11), Pair: "BTC-USD"},
				{Price: decimal.NewFromInt(20), Volume: decimal.NewFromInt(23), Pair: "BTC-USD"},
				{Price: decimal.NewFromInt(8), Volume: decimal.NewFromInt(15), Pair: "BTC-USD"},
				{Price: decimal.NewFromInt(11), Volume: decimal.NewFromInt(165), Pair: "BTC-USD"},
				{Price: decimal.NewFromInt(19), Volume: decimal.NewFromInt(1000), Pair: "BTC-USD"},
			},
			want: []vwap.Result{
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("10")},
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("16.7647058823529412")},
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("14.0816326530612245")},
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("11.7980295566502463")},
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("17.7415254237288136")},
			},
			size: 3,
		},
		{
			name: "MixedPairsBigWindow",
			dataPoints: []websocket.DataPoint{
				{Pair: "BTC-USD", Price: decimal.NewFromInt(10), Volume: decimal.NewFromInt(10)},
				{Pair: "BTC-USD", Price: decimal.NewFromInt(10), Volume: decimal.NewFromInt(10)},
				{Pair: "ETH-USD", Price: decimal.NewFromInt(31), Volume: decimal.NewFromInt(30)},
				{Pair: "BTC-USD", Price: decimal.NewFromInt(21), Volume: decimal.NewFromInt(20)},
				{Pair: "ETH-USD", Price: decimal.NewFromInt(41), Volume: decimal.NewFromInt(33)},
				{Pair: "ETH-USD", Price: decimal.NewFromInt(45), Volume: decimal.NewFromInt(15)},
				{Pair: "BTC-USD", Price: decimal.NewFromInt(25), Volume: decimal.NewFromInt(100)},
			},
			want: []vwap.Result{
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("10")},
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("10")},
				{Pair: "ETH-USD", VWAP: decimal.RequireFromString("31")},
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("15.5")},
				{Pair: "ETH-USD", VWAP: decimal.RequireFromString("36.2380952380952381")},
				{Pair: "ETH-USD", VWAP: decimal.RequireFromString("37.9230769230769231")},
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("24.3333333333333333")},
			},
			size: 5,
		},
		{
			name: "MixedPairsSmallWindow",
			dataPoints: []websocket.DataPoint{
				{Pair: "BTC-USD", Price: decimal.NewFromInt(10), Volume: decimal.NewFromInt(10)},
				{Pair: "BTC-USD", Price: decimal.NewFromInt(10), Volume: decimal.NewFromInt(10)},
				{Pair: "ETH-USD", Price: decimal.NewFromInt(31), Volume: decimal.NewFromInt(30)},
				{Pair: "BTC-USD", Price: decimal.NewFromInt(21), Volume: decimal.NewFromInt(20)},
				{Pair: "ETH-USD", Price: decimal.NewFromInt(41), Volume: decimal.NewFromInt(33)},
				{Pair: "ETH-USD", Price: decimal.NewFromInt(45), Volume: decimal.NewFromInt(15)},
				{Pair: "BTC-USD", Price: decimal.NewFromInt(25), Volume: decimal.NewFromInt(100)},
			},
			want: []vwap.Result{
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("10")},
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("10")},
				{Pair: "ETH-USD", VWAP: decimal.RequireFromString("31")},
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("21")},
				{Pair: "ETH-USD", VWAP: decimal.RequireFromString("41")},
				{Pair: "ETH-USD", VWAP: decimal.RequireFromString("42.25")},
				{Pair: "BTC-USD", VWAP: decimal.RequireFromString("25")},
			},
			size: 2,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			var in = make(chan websocket.DataPoint)
			var out = make(chan vwap.Result)

			w, err := vwap.NewWindow(tt.size)
			require.NoError(t, err)

			go w.Calculate(ctx, in, out)

			for i, d := range tt.dataPoints {
				in <- d

				v, ok := <-out
				if !ok {
					break
				}
				log.Print(tt.want[i].VWAP.String(), " = ", v.VWAP.String())
				require.Equal(t, tt.want[i].VWAP.String(), v.VWAP.String())
			}
		})
	}
}
