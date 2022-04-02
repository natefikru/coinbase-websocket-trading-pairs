package websocketClient

import (
	"coinbase-websocket-trading-pairs/util"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type WebSocketClient struct {
}

func NewWebSocketClient(config *util.Config) *WebSocketClient {
	return &WebSocketClient{}
}

func (ws *WebSocketClient) EstablishConnection(url string) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		err = errors.Wrap(err, "Error connecting to Websocket Server")
		return nil, err
	}
	log.Printf("Established Socket Connection to: %v", url)
	return conn, nil
}

func (ws *WebSocketClient) WriteMessageToSocketConnInterval(conn *websocket.Conn, data []byte, seconds time.Duration) error {
	for {
		select {
		case <-time.After(seconds):
			// Send a packet every 4 seconds
			err := conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				err = errors.Wrap(err, "Error during writing to websocket")
				return err
			}
		}
	}
}

// ReadMessageFromSockenConn: reads incoming messages from socket connection
func (ws *WebSocketClient) ReadMessageFromSockenConn(conn *websocket.Conn) ([]byte, error) {
	_, msgBytes, err := conn.ReadMessage()
	if err != nil {
		err = errors.Wrap(err, "error ReadMessageFromSockenConn")
		return nil, err
	}
	return msgBytes, nil
}
