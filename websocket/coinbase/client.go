package coinbase

import (
	"log"

	"github.com/act28/vwap/websocket"
	ws "golang.org/x/net/websocket"
)

type client struct {
	conn *ws.Conn
}

// NewClient returns a new websocket client, or an error.
func NewClient(url string) (websocket.Client, error) {
	conn, err := ws.Dial(url, "", "http://localhost/")
	if err != nil {
		return nil, err
	}

	log.Printf("websocket connected to: %s", url)

	return &client{
		conn: conn,
	}, nil
}
