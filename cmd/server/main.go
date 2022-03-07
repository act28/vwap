package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/act28/vwap/vwap"
	"github.com/act28/vwap/websocket"
	"github.com/act28/vwap/websocket/coinbase"
	"github.com/dogmatiq/dodeca/config"
	"go.uber.org/dig"
	"golang.org/x/sync/errgroup"
)

func init() {
	provide(func() context.Context {
		return context.Background()
	})

	// Setup a config bucket for reading environment variables.
	provide(func() config.Bucket {
		return config.Environment()
	})

	// Setup a websocket client.
	provide(func(env config.Bucket, ctx context.Context) websocket.Client {
		// @TODO use a WebsocketProvider to return a client implementation based
		// on env config.
		ws, err := coinbase.NewClient(
			ctx,
			config.AsStringDefault(env, "WEBSOCKET_ENDPOINT", coinbase.ProdWSURL),
		)
		if err != nil {
			panic(err)
		}
		return ws
	})

	// Setup the data channel.
	provide(func() chan websocket.DataPoint {
		return make(chan websocket.DataPoint)
	})

	// Setup the output channel.
	provide(func() chan vwap.Result {
		return make(chan vwap.Result)
	})
}

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	invoke(func(
		env config.Bucket,
		ctx context.Context,
		ws websocket.Client,
		in chan websocket.DataPoint,
		out chan vwap.Result,
	) error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		g, ctx := errgroup.WithContext(ctx)

		g.Go(func() error {
			pairs := strings.Split(
				strings.TrimSpace(
					config.AsStringDefault(env, "TRADING_PAIRS", "BTC-USD ETH-USD ETH-BTC"),
				),
				" ",
			)

			// Start the websocket feed.
			if err := ws.Subscribe(ctx, pairs); err != nil {
				return err
			}
			ws.Receive(ctx, in)

			return nil
		})

		g.Go(func() error {
			// Calculate the VWAP.
			w, err := vwap.NewWindow(config.AsUintDefault(env, "WINDOW_SIZE", vwap.MaxWindowSize))
			if err != nil {
				return err
			}

			go w.Calculate(ctx, in, out)

			return nil
		})

		g.Go(func() error {
			for v := range out {
				select {
				case <-ctx.Done():
				default:
					log.Print(v)
				}
			}
			return nil
		})

		// Wait for a signal or an error to occur.
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		select {
		case <-sig:
			// We have been asked to shutdown, cancel the context and wait for
			// the goroutines to finish but do not report any errors.
			cancel()
			return nil
		case <-ctx.Done():
			// The context was canceled before we received a signal, something
			// has gone wrong. Wait for the goroutines to finish and report the
			// causal error.
			return g.Wait()
		}
	})

	return nil
}

var container = dig.New()

func provide(fn interface{}) {
	if err := container.Provide(fn); err != nil {
		panic(err)
	}
}

func invoke(fn interface{}) {
	if err := container.Invoke(fn); err != nil {
		panic(err)
	}
}
