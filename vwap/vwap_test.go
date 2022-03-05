package vwap_test

import (
	"testing"

	"github.com/act28/vwap/vwap"
	"github.com/act28/vwap/websocket"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestVWAPCalculation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		dataPoints []websocket.DataPoint
		want       map[string]decimal.Decimal
		size       uint
	}{
		{
			name:       "EmptyDataPoints",
			dataPoints: []websocket.DataPoint{},
			want:       map[string]decimal.Decimal{},
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
			want: map[string]decimal.Decimal{
				"BTC-USD": decimal.RequireFromString("11.705607476635514"),
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
			want: map[string]decimal.Decimal{
				"BTC-USD": decimal.RequireFromString("11.705607476635514"),
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
			want: map[string]decimal.Decimal{
				"BTC-USD": decimal.RequireFromString("17.7415254237288136"),
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
			want: map[string]decimal.Decimal{
				"ETH-USD": decimal.RequireFromString("37.9230769230769231"),
				"BTC-USD": decimal.RequireFromString("24.3333333333333333"),
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
			want: map[string]decimal.Decimal{
				"ETH-USD": decimal.RequireFromString("42.25"),
				"BTC-USD": decimal.RequireFromString("25"),
			},
			size: 2,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			list, err := vwap.NewWindow(tt.size)
			require.NoError(t, err)

			for _, d := range tt.dataPoints {
				list.Push(d)
			}

			for k := range tt.want {
				require.Equal(t, tt.want[k].String(), list.VWAP(k).String())
			}
		})
	}
}
