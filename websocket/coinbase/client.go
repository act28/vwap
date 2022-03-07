package coinbase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"sync"

	"github.com/act28/vwap/websocket"
	"github.com/shopspring/decimal"
	ws "nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	// ProdWSURL is the production websocket URL.
	ProdWSURL = "wss://ws-feed.exchange.coinbase.com"
	// SandboxWSURL is the sandbox websocket URL.
	SandboxWSURL = "wss://ws-feed-public.sandbox.exchange.coinbase.com"
)

// requestType is the type of websocket request.
type requestType string

const (
	// requestTypeSubscribe indicates that this is a channel `subscribe`
	// request.
	requestTypeSubscribe requestType = "subscribe"
)

// channelType is the type of channel subscription.
type channelType string

const (
	// channelTypeMatches indicates a `matches` channel subscription type.
	channelTypeMatches channelType = "matches"
)

// request is a Coinbase websocket request.
//
// This uses the alternate form of specifying product IDs in the root object, to
// add all product IDs to the subscribed channels.
type request struct {
	Type       requestType   `json:"type"`
	ProductIDs []string      `json:"product_ids"`
	Channels   []channelType `json:"channels"`
}

// responseType is the type of websocket response.
type responseType string

const (
	// responseTypeError indicates an `error` response from the websocket.
	responseTypeError responseType = "error"
	// responseTypeLastMatch indicates a `last_match` response from the websocket.
	responseTypeLastMatch responseType = "last_match"
	// responseTypeMatch indicates a `match` response from the websocket.
	responseTypeMatch responseType = "match"
)

// channel is a data struct that forms part of the channel subscription response.
type channel struct {
	Name       channelType `json:"name"`
	ProductIDs []string    `json:"product_id"`
}

// subscriptionResponse is a Coinbase websocket channel subscription response.
type subscriptionResponse struct {
	Type      responseType    `json:"type"`
	Channels  []channel       `json:"channels"`
	Message   string          `json:"message,omitempty"`
	Size      decimal.Decimal `json:"size"`
	Price     decimal.Decimal `json:"price"`
	ProductID string          `json:"product_id"`
}

// matchResponse is a Coinbase websocket `match` channel response.
type matchResponse struct {
	Type      responseType    `json:"type"`
	Sequence  *big.Int        `json:"sequence"`
	ProductID string          `json:"product_id"`
	Size      decimal.Decimal `json:"size"`
	Price     decimal.Decimal `json:"price"`
}

type client struct {
	conn *ws.Conn
}

// NewClient returns a new websocket client, or an error.
func NewClient(ctx context.Context, url string) (websocket.Client, error) {
	conn, _, err := ws.Dial(ctx, url, &ws.DialOptions{
		CompressionMode: ws.CompressionContextTakeover,
	})
	if err != nil {
		return nil, err
	}

	log.Printf("websocket connected to: %s", url)

	return &client{
		conn: conn,
	}, nil
}

// Subscribe subscribes to the `matches` channel on the websocket.
func (c *client) Subscribe(ctx context.Context, tradingPairs []string) error {
	if len(tradingPairs) == 0 {
		return errors.New(`subscription error: tradingPairs must be provided`)
	}

	if err := wsjson.Write(ctx, c.conn, request{
		Type:       requestTypeSubscribe,
		ProductIDs: tradingPairs,
		Channels: []channelType{
			channelTypeMatches,
		},
	}); err != nil {
		return fmt.Errorf(`subscription error: "%w"`, err)
	}

	var resp subscriptionResponse
	if err := wsjson.Read(ctx, c.conn, &resp); err != nil {
		return fmt.Errorf(`subscription response error:: "%w"`, err)
	}

	if resp.Type == responseTypeError {
		return fmt.Errorf(`subscription error: "%v"`, resp.Message)
	}

	return nil
}

// Receive listens on the data channel, and sends datapoints to the
// receiver.
func (c *client) Receive(ctx context.Context, receiver chan<- websocket.DataPoint) {
	for {
		select {
		case <-ctx.Done():
			err := c.conn.Close(ws.StatusNormalClosure, "")
			log.Printf(`websocket closed: "%s"`, err)
			log.Printf("context done: %s", ctx.Err())
			close(receiver)
			return

		default:
			m, buf, err := c.conn.Reader(ctx)
			if err != nil {
				log.Printf(`channel error: "%s"`, err)
				_ = c.conn.CloseRead(ctx)
				return
			}

			if m != ws.MessageText {
				// Ignore non-text messages.
				continue
			}

			var match matchResponse
			dec := json.NewDecoder(buf)
			if err := dec.Decode(&match); err != nil {
				if err == io.EOF {
					ctx = c.conn.CloseRead(ctx)
					continue
				}
				log.Printf(`buffer read error: "%s"`, err)
				return
			}

			if data, ok := makeDataPoint(match); ok {
				receiver <- data
			}
		}
	}
}

var sequencer struct {
	pair map[string]*big.Int

	m sync.Mutex
}

func makeDataPoint(m matchResponse) (websocket.DataPoint, bool) {
	sequencer.m.Lock()
	defer sequencer.m.Unlock()

	if m.Type != responseTypeMatch && m.Type != responseTypeLastMatch {
		// Ignore anything that is not a `match` or`last_match`.
		return websocket.DataPoint{}, false
	}

	seq, ok := sequencer.pair[m.ProductID]
	if !ok {
		seq = big.NewInt(0)
		sequencer.pair = map[string]*big.Int{
			m.ProductID: seq,
		}
	}

	if (m.Sequence).Cmp(seq) == -1 {
		// Ignore out of order messages.
		return websocket.DataPoint{}, false
	}

	sequencer.pair[m.ProductID] = m.Sequence

	return websocket.DataPoint{
		Pair:   m.ProductID,
		Volume: m.Size,
		Price:  m.Price,
	}, true
}
