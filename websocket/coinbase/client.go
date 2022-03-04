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

// RequestType is the type of websocket request.
type RequestType string

const (
	// RequestTypeSubscribe indicates that this is a channel `subscribe`
	// request.
	RequestTypeSubscribe RequestType = "subscribe"
)

// ChannelType is the type of channel subscription.
type ChannelType string

const (
	// ChannelTypeMatches indicates a `matches` channel subscription type.
	ChannelTypeMatches ChannelType = "matches"
)

// Request is a Coinbase websocket request.
//
// This uses the alternate form of specifying product IDs in the root object, to
// add all product IDs to the subscribed channels.
type Request struct {
	Type       RequestType   `json:"type"`
	ProductIDs []string      `json:"product_ids"`
	Channels   []ChannelType `json:"channels"`
}

// ResponseType is the type of websocket response.
type ResponseType string

const (
	// ResponseTypeError indicates an `error` response from the websocket.
	ResponseTypeError ResponseType = "error"
	// ResponseTypeLastMatch indicates a `last_match` response from the websocket.
	ResponseTypeLastMatch ResponseType = "last_match"
	// ResponseTypeMatch indicates a `match` response from the websocket.
	ResponseTypeMatch ResponseType = "match"
	//ResponseTypeSubscriptions indicates a `subscription` response from the
	//websocket.
	ResponseTypeSubscriptions ResponseType = "subscriptions"
)

// Channel is a data struct that forms part of the channel subscription response.
type Channel struct {
	Name       ChannelType
	ProductIDs []string
}

// SubscriptionResponse is a Coinbase websocket channel subscription response.
type SubscriptionResponse struct {
	Type      ResponseType    `json:"type"`
	Channels  []Channel       `json:"channels"`
	Message   string          `json:"message,omitempty"`
	Size      decimal.Decimal `json:"size"`
	Price     decimal.Decimal `json:"price"`
	ProductID string          `json:"product_id"`
}

// MatchResponse is a Coinbase websocket `match` channel response.
type MatchResponse struct {
	Type      ResponseType    `json:"type"`
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
func (c *client) Subscribe(ctx context.Context, tradingPairs []string, receiver chan websocket.DataPoint) error {
	if len(tradingPairs) == 0 {
		return errors.New(`subscription error: tradingPairs must be provided`)
	}

	if err := wsjson.Write(ctx, c.conn, Request{
		Type:       RequestTypeSubscribe,
		ProductIDs: tradingPairs,
		Channels: []ChannelType{
			ChannelTypeMatches,
		},
	}); err != nil {
		return fmt.Errorf(`subscription error: "%w"`, err)
	}

	var resp SubscriptionResponse
	if err := wsjson.Read(ctx, c.conn, &resp); err != nil {
		return fmt.Errorf(`subscription response error:: "%w"`, err)
	}

	if resp.Type == ResponseTypeError {
		return fmt.Errorf(`subscription error: "%v"`, resp.Message)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				if err := c.conn.Close(ws.StatusInternalError, `failed to close websocket`); err != nil {
					log.Printf(`websocket error: "%s"`, err)
				}
				log.Printf("context done: %s", ctx.Err())
				return

			default:
				m, buf, err := c.conn.Reader(ctx)
				if err != nil {
					log.Printf(`channel error: "%s"`, err)
					return
				}

				if m != ws.MessageText {
					// Ignore non-text messages.
					continue
				}
				var match MatchResponse
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
	}()

	return nil
}

var sequencer sync.Map

func makeDataPoint(m MatchResponse) (websocket.DataPoint, bool) {
	if m.Type != ResponseTypeMatch && m.Type != ResponseTypeLastMatch {
		// Ignore anything that is not a `match` or`last_match`.
		return websocket.DataPoint{}, false
	}

	seq, _ := sequencer.LoadOrStore(m.ProductID, m.Sequence)
	if (m.Sequence).Cmp(seq.(*big.Int)) == -1 {
		// Ignore out of order messages.
		return websocket.DataPoint{}, false
	}

	sequencer.Store(m.ProductID, m.Sequence)

	return websocket.DataPoint{
		Pair:   m.ProductID,
		Volume: m.Size,
		Price:  m.Price,
	}, true
}
