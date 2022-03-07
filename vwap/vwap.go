package vwap

import (
	"context"
	"sync"

	"github.com/act28/vwap/websocket"
	"github.com/shopspring/decimal"
)

// MaxWindowSize is the maximum number of dataPoints in the window.
const MaxWindowSize = 200

type Result struct {
	Pair string
	VWAP decimal.Decimal
}

type cumSum struct {
	sumPQ decimal.Decimal
	sumQ  decimal.Decimal
	vwap  decimal.Decimal
}

// Window is a synchronized structure of DataPoints time series, and cumulative
// totals for calculating the VWAP.
type Window struct {
	size   uint
	series []websocket.DataPoint
	cumSum map[string]cumSum

	m sync.Mutex
}

// NewWindow returns a new sliding Window of size `size`.
func NewWindow(size uint) (Window, error) {
	if size == 0 || size > MaxWindowSize {
		size = MaxWindowSize
	}

	return Window{
		series: []websocket.DataPoint{},
		size:   size,
		cumSum: map[string]cumSum{},
	}, nil
}

func (w *Window) VWAP(pair string) Result {
	tp, ok := w.cumSum[pair]
	if !ok {
		return Result{}
	}

	return Result{
		Pair: pair,
		VWAP: tp.vwap,
	}
}

// Push pushes a new datapoint into the window.
func (w *Window) Push(dp websocket.DataPoint) {
	w.m.Lock()
	defer w.m.Unlock()

	if len(w.series) >= int(w.size) {
		// Drop the oldest datapoint from the series.
		p := w.series[0]
		w.series = w.series[1:]

		// Subtract the dropped point's PQ and Q from the cumulative totals for
		// the dropped point's trading pair and recalculate the VWAP.
		pp, ok := w.cumSum[p.Pair]
		if ok {
			pp.sumPQ = pp.sumPQ.Sub(p.Price.Mul(p.Volume))
			pp.sumQ = pp.sumQ.Sub(p.Volume)
			pp.vwap = decimal.NewFromInt(0)
			if !pp.sumQ.IsZero() {
				pp.vwap = pp.sumPQ.Div(pp.sumQ)
			}
			w.cumSum[p.Pair] = pp
		}
	}

	// Get the cumulative totals for the trading pair.
	tp, ok := w.cumSum[dp.Pair]
	if !ok {
		init := decimal.NewFromInt(0)
		tp = cumSum{
			sumPQ: init,
			sumQ:  init,
			vwap:  init,
		}
	}

	// Append the next datapoint to the series, and update the cumulative totals
	// for the trading pair.
	w.series = append(w.series, dp)
	tp.sumPQ = tp.sumPQ.Add(dp.Price.Mul(dp.Volume))
	tp.sumQ = tp.sumQ.Add(dp.Volume)

	// Calculate the new VWAP.
	if !tp.sumQ.IsZero() {
		tp.vwap = tp.sumPQ.Div(tp.sumQ)
	}

	w.cumSum[dp.Pair] = tp
}

func (w *Window) Calculate(ctx context.Context, in <-chan websocket.DataPoint, out chan<- Result) error {
	for {
		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				return err
			}
		default:
			dp, ok := <-in
			if !ok {
				close(out)
				return ctx.Err()
			}

			w.Push(dp)

			tp, ok := w.cumSum[dp.Pair]
			if !ok {
				out <- Result{
					Pair: dp.Pair,
					VWAP: decimal.NewFromInt(0),
				}
				continue
			}

			out <- Result{
				Pair: dp.Pair,
				VWAP: tp.vwap,
			}
		}
	}
}
