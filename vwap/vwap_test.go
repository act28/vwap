package vwap_test

import (
	"testing"

	"github.com/act28/vwap/vwap"
	"github.com/act28/vwap/websocket"
	"github.com/dogmatiq/dapper"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestVWAPCalculation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name       string
		DataPoints []websocket.DataPoint
		WantVwap   map[string]decimal.Decimal
		MaxSize    int
	}{
		// {
		// 	Name:       "EmptyDataPoints",
		// 	DataPoints: []websocket.DataPoint{},
		// 	WantVwap:   map[string]decimal.Decimal{},
		// 	MaxSize:    4,
		// },
		// {
		// 	Name: "InvalidWindowSize",
		// 	DataPoints: []websocket.DataPoint{
		// 		{Price: decimal.NewFromInt(10), Volume: decimal.NewFromInt(11), Pair: "BTC-USD"},
		// 		{Price: decimal.NewFromInt(20), Volume: decimal.NewFromInt(23), Pair: "BTC-USD"},
		// 		{Price: decimal.NewFromInt(8), Volume: decimal.NewFromInt(15), Pair: "BTC-USD"},
		// 		{Price: decimal.NewFromInt(11), Volume: decimal.NewFromInt(165), Pair: "BTC-USD"},
		// 	},
		// 	WantVwap: map[string]decimal.Decimal{
		// 		{Pair: "BTC-USD", VWAP: decimal.RequireFromString("11.705607476635514")},
		// 	},
		// 	MaxSize: -1,
		// },
		// {
		// 	Name: "SinglePairBigWindow",
		// 	DataPoints: []websocket.DataPoint{
		// 		{Price: decimal.NewFromInt(10), Volume: decimal.NewFromInt(11), Pair: "BTC-USD"},
		// 		{Price: decimal.NewFromInt(20), Volume: decimal.NewFromInt(23), Pair: "BTC-USD"},
		// 		{Price: decimal.NewFromInt(8), Volume: decimal.NewFromInt(15), Pair: "BTC-USD"},
		// 		{Price: decimal.NewFromInt(11), Volume: decimal.NewFromInt(165), Pair: "BTC-USD"},
		// 	},
		// 	WantVwap: map[string]decimal.Decimal{
		// 		{Pair: "BTC-USD", VWAP: decimal.RequireFromString("11.705607476635514")},
		// 	},
		// 	MaxSize: 5,
		// },
		{
			Name: "SinglePairSmallWindow",
			DataPoints: []websocket.DataPoint{
				{Price: decimal.NewFromInt(10), Volume: decimal.NewFromInt(11), Pair: "BTC-USD"},
				{Price: decimal.NewFromInt(20), Volume: decimal.NewFromInt(23), Pair: "BTC-USD"},
				{Price: decimal.NewFromInt(8), Volume: decimal.NewFromInt(15), Pair: "BTC-USD"},
				{Price: decimal.NewFromInt(11), Volume: decimal.NewFromInt(165), Pair: "BTC-USD"},
				{Price: decimal.NewFromInt(19), Volume: decimal.NewFromInt(1000), Pair: "BTC-USD"},
			},
			WantVwap: map[string]decimal.Decimal{
				"BTC-USD": decimal.RequireFromString("17.7415254237288136"),
			},
			MaxSize: 3,
		},
		// {
		// 	Name: "MixedPairsBigWindow",
		// 	DataPoints: []websocket.DataPoint{
		// 		{Pair: "BTC-USD", Price: decimal.NewFromInt(10), Volume: decimal.NewFromInt(10)},
		// 		{Pair: "BTC-USD", Price: decimal.NewFromInt(10), Volume: decimal.NewFromInt(10)},
		// 		{Pair: "ETH-USD", Price: decimal.NewFromInt(31), Volume: decimal.NewFromInt(30)},
		// 		{Pair: "BTC-USD", Price: decimal.NewFromInt(21), Volume: decimal.NewFromInt(20)},
		// 		{Pair: "ETH-USD", Price: decimal.NewFromInt(41), Volume: decimal.NewFromInt(33)},
		// 		{Pair: "ETH-USD", Price: decimal.NewFromInt(45), Volume: decimal.NewFromInt(15)},
		// 		{Pair: "BTC-USD", Price: decimal.NewFromInt(25), Volume: decimal.NewFromInt(100)},
		// 	},
		// 	WantVwap: map[string]decimal.Decimal{
		// 		"ETH-USD": decimal.RequireFromString("37.9230769230769231").
		// 		"BTC-USD": decimal.RequireFromString("22.2857142857142857").
		// 	},
		// 	MaxSize: 5,
		// },
		// {
		// 	Name: "MixedPairsSmallWindow",
		// 	DataPoints: []websocket.DataPoint{
		// 		{Pair: "BTC-USD", Price: decimal.NewFromInt(10), Volume: decimal.NewFromInt(10)},
		// 		{Pair: "BTC-USD", Price: decimal.NewFromInt(10), Volume: decimal.NewFromInt(10)},
		// 		{Pair: "ETH-USD", Price: decimal.NewFromInt(31), Volume: decimal.NewFromInt(30)},
		// 		{Pair: "BTC-USD", Price: decimal.NewFromInt(21), Volume: decimal.NewFromInt(20)},
		// 		{Pair: "ETH-USD", Price: decimal.NewFromInt(41), Volume: decimal.NewFromInt(33)},
		// 		{Pair: "ETH-USD", Price: decimal.NewFromInt(45), Volume: decimal.NewFromInt(15)},
		// 		{Pair: "BTC-USD", Price: decimal.NewFromInt(25), Volume: decimal.NewFromInt(100)},
		// 	},
		// 	WantVwap: map[string]decimal.Decimal{
		// 		"ETH-USD": decimal.RequireFromString("37.9230769230769231"),
		// 		"BTC-USD": decimal.RequireFromString("42.25"),
		// 	},
		// 	MaxSize: 2,
		// },
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			list, err := vwap.NewWindow(tt.DataPoints, uint(tt.MaxSize))
			require.NoError(t, err)

			for _, d := range tt.DataPoints {
				list.Push(d)
			}

			dapper.Print(list)
			for k := range tt.WantVwap {
				require.Equal(t, tt.WantVwap[k].String(), list.VWAP[k].String())
			}
		})
	}
}
