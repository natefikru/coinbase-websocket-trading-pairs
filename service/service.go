package service

import (
	"coinbase-websocket-trading-pairs/util"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type Service struct {
	WebsocketClient IWebSocketClient
	Config          *util.Config
	TotalValues     map[string]PairTotalValue
}

func NewService(websocketClient IWebSocketClient, config *util.Config) *Service {
	totalValues := make(map[string]PairTotalValue)
	return &Service{
		WebsocketClient: websocketClient,
		Config:          config,
		TotalValues:     totalValues,
	}
}

func (s *Service) Run() error {
	// Establish WebSocket Connection
	conn, err := s.WebsocketClient.EstablishConnection(s.Config.CoinbaseSocketURL)
	if err != nil {
		err = errors.Wrap(err, "Error Establishing Connection")
		return err
	}

	// Start Websocket Connection Listener
	go s.socketListener(conn)

	// set up the coinbase subscription message object
	request := s.setUpRequest()

	// convert request to byte slice
	requestBytes, err := json.Marshal(request)
	if err != nil {
		err = errors.Wrap(err, "Error marshalling request")
		return err
	}

	// write subscribe message to the socket connection
	err = s.WebsocketClient.WriteMessageToSocketConnInterval(conn, requestBytes, time.Second*subscribeIntervalSeconds)
	if err != nil {
		err = errors.Wrap(err, "Error Executing Connection")
		return err
	}
	return nil
}

// socketListener: infinite loop the reads all incoming messages from the socket connection
func (s *Service) socketListener(conn *websocket.Conn) {
	for {
		// reads incoming messages
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error in reading messags from connection: ", err)
			return
		}

		// convert bytes to response struct
		var response Response
		err = json.Unmarshal(msgBytes, &response)
		if err != nil {
			log.Println("Error unmarshalling match response:", err)
			return
		}

		s.evaluate(&response)
	}
}

// setUpRequest: initializes request with pre-defined channels and trading pair product id's
func (s *Service) setUpRequest() *SubscribeMessage {
	var channels []interface{}
	channels = append(channels, matches)
	productIDs := s.getSupportedProductIDs()

	// Initialize trading pairs dict
	for _, key := range productIDs {
		val := PairTotalValue{}
		s.TotalValues[key] = val
	}

	return &SubscribeMessage{
		Type:       subscribe,
		ProductIDs: productIDs,
		Channels:   channels,
	}
}

func (s *Service) evaluate(response *Response) error {
	switch response.Type {
	case match:
		err := s.evaluateMatch(response)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) evaluateMatch(response *Response) error {
	pairValues := s.TotalValues[response.ProductID]
	if pairValues.TotalCount < 200 {
		// Add resp to end of queue
		pairValues.MatchQueue = append(pairValues.MatchQueue, *response)

		// Add new price to total sum
		price, err := strconv.ParseFloat(response.Price, 64)
		if err != nil {
			err = errors.Wrap(err, "error parsing sub 200 response.Price")
		}
		pairValues.TotalSum += price

		// add 1 to count
		pairValues.TotalCount += 1
	} else {
		// Save oldest Response
		oldResp := pairValues.MatchQueue[0]

		// Remove oldest response from front of queue, add new response to end of queue
		pairValues.MatchQueue = append(pairValues.MatchQueue[1:], *response)

		// subtract oldest resp price from total sum add new resp price to total sum
		oldPrice, err := strconv.ParseFloat(oldResp.Price, 64)
		if err != nil {
			err = errors.Wrap(err, "error parsing plus 200 oldResp.Price")
		}
		newPrice, err := strconv.ParseFloat(response.Price, 64)
		if err != nil {
			err = errors.Wrap(err, "error parsing plus 200 response.Price")
		}
		pairValues.TotalSum = pairValues.TotalSum - oldPrice + newPrice
	}

	// Evaluate and save new average by dividing totalSum by totalCount
	pairValues.Average = pairValues.TotalSum / float64(pairValues.TotalCount)
	s.TotalValues[response.ProductID] = pairValues

	//TODO: save new line file
	fmt.Println(response.ProductID, pairValues.TotalCount, pairValues.Average)

	return nil
}
