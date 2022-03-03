package coinbase_test

import (
	"testing"

	"github.com/act28/vwap/websocket/coinbase"
	"github.com/stretchr/testify/require"
)

func TestNewWebsocketClient(t *testing.T) {
	t.Parallel()

	_, err := coinbase.NewClient(coinbase.SandboxWSURL)
	require.NoError(t, err)
}
