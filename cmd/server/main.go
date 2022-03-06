package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/act28/vwap/vwap"
	"github.com/act28/vwap/websocket"
	"github.com/act28/vwap/websocket/coinbase"
	"github.com/dogmatiq/dodeca/config"
	"go.uber.org/dig"
)

func init() {
	rand.Seed(time.Now().UnixNano())

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
		recv chan websocket.DataPoint,
	) error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		go func() {
			<-ctx.Done()
		}()

		pairs := strings.Split(
			strings.TrimSpace(
				config.AsStringDefault(env, "TRADING_PAIRS", "BTC-USD ETH-USD ETH-BTC"),
			),
			" ",
		)

		// Start the websocket feed.
		if err := ws.Subscribe(
			ctx,
			pairs,
			recv,
		); err != nil {
			return err
		}

		// Calculate the VWAP and dump to tty.
		w, err := vwap.NewWindow(config.AsUintDefault(env, "WINDOW_SIZE", vwap.MaxWindowSize))
		if err != nil {
			return err
		}

		for m := range recv {
			select {
			case <-time.After(3 * time.Second):
				cancel()
				return err
			case <-ctx.Done():
				break
			default:
				w.Push(m)
				log.Print(w.VWAP(m.Pair))
			}
		}

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
			return nil
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
