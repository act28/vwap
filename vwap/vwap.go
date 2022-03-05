package vwap

import (
	"sync"

	"github.com/act28/vwap/websocket"
	"github.com/shopspring/decimal"
)

const maxWindowSize = 200

// Window is a synchronized structure of DataPoints time series, and cumulative
// totals for calculating the VWAP.
type Window struct {
	size   uint
	series []websocket.DataPoint
	sumPQ  map[string]decimal.Decimal
	sumQ   map[string]decimal.Decimal
	vwap   map[string]decimal.Decimal

	mu sync.Mutex
}

// NewWindow returns a new sliding Window of size `size`.
func NewWindow(size uint) (Window, error) {
	if size == 0 || size > maxWindowSize {
		size = maxWindowSize
	}

	return Window{
		series: []websocket.DataPoint{},
		size:   size,
		sumPQ:  map[string]decimal.Decimal{},
		sumQ:   map[string]decimal.Decimal{},
		vwap:   map[string]decimal.Decimal{},
	}, nil
}

// Len returns the length of the time series.
func (w *Window) Len() int {
	return len(w.series)
}

func (w *Window) VWAP(pair string) decimal.Decimal {
	return w.vwap[pair]
}

// Push pushes a new datapoint into the window.
func (w *Window) Push(dp websocket.DataPoint) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, ok := w.vwap[dp.Pair]; !ok {
		init := decimal.NewFromInt(0)
		w.sumPQ[dp.Pair] = init
		w.sumQ[dp.Pair] = init
	}

	if w.Len() >= int(w.size) {
		// Drop the oldest point from the series.
		p := w.series[0]
		w.series = w.series[1:]

		// Subtract the dropped point's PQ and Q from the cumulative totals.
		w.sumPQ[p.Pair] = w.sumPQ[p.Pair].Sub(p.Price.Mul(p.Volume))
		w.sumQ[p.Pair] = w.sumQ[p.Pair].Sub(p.Volume)
	}

	// Append the next datapoint to the series, and update the cumulative totals.
	w.series = append(w.series, dp)
	w.sumPQ[dp.Pair] = w.sumPQ[dp.Pair].Add(dp.Price.Mul(dp.Volume))
	w.sumQ[dp.Pair] = w.sumQ[dp.Pair].Add(dp.Volume)

	// Calculate the new VWAP.
	if !w.sumQ[dp.Pair].IsZero() {
		w.vwap[dp.Pair] = w.sumPQ[dp.Pair].Div(w.sumQ[dp.Pair])
	}
}
