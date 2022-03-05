package vwap

import (
	"sync"

	"github.com/act28/vwap/websocket"
	"github.com/dogmatiq/dapper"
	"github.com/shopspring/decimal"
)

const maxWindowSize = 200

// Result is the VWAP for the trading Pair at a particular point in time.
type Result struct {
	Pair string
	VWAP decimal.Decimal
}

// Window is a window of DataPoints.
type Window struct {
	Size   uint
	Series []websocket.DataPoint
	SumPQ  map[string]decimal.Decimal
	SumQ   map[string]decimal.Decimal
	VWAP   map[string]decimal.Decimal

	mu sync.Mutex
}

func NewWindow(series []websocket.DataPoint, size uint) (Window, error) {
	if size == 0 || size > maxWindowSize {
		size = maxWindowSize
	}

	if len(series) > int(size) {
		// Discard the first k items.
		k := len(series) - int(size)
		series = series[k:]
		dapper.Print(series)
	}

	return Window{
		Series: series,
		Size:   size,
		SumPQ:  make(map[string]decimal.Decimal),
		SumQ:   make(map[string]decimal.Decimal),
		VWAP:   make(map[string]decimal.Decimal),
	}, nil
}

// Len returns the length of the time series.
func (w *Window) Len() int {
	return len(w.Series)
}

// Push pushes a new datapoint into the window.
func (w *Window) Push(dp websocket.DataPoint) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.Len() >= int(w.Size) {
		// Drop the oldest point from the series.
		p := w.Series[0]
		w.Series = w.Series[1:]

		// Subtract the dropped point's PW.and W.from the cumulative totals.
		w.SumPQ[p.Pair] = w.SumPQ[p.Pair].Sub(p.Price.Mul(p.Volume))
		w.SumQ[p.Pair] = w.SumQ[p.Pair].Sub(p.Volume)
	}

	if _, ok := w.VWAP[dp.Pair]; !ok {
		w.SumPQ[dp.Pair] = dp.Price.Mul(dp.Volume)
		w.SumQ[dp.Pair] = dp.Volume
		w.VWAP[dp.Pair] = w.SumPQ[dp.Pair]
	}

	// Append the next datapoint to the series, and update the cumulative totals.
	w.Series = append(w.Series, dp)
	w.SumPQ[dp.Pair] = w.SumPQ[dp.Pair].Add(dp.Price.Mul(dp.Volume))
	w.SumQ[dp.Pair] = w.SumQ[dp.Pair].Add(dp.Volume)

	// Calculate the new VWAP.
	if !w.SumQ[dp.Pair].IsZero() {
		w.VWAP[dp.Pair] = w.SumPQ[dp.Pair].Div(w.SumQ[dp.Pair])
	}
}
