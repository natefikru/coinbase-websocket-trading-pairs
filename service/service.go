package service

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type Service struct {
	WebsocketClient IWebSocketClient
	FileClient      IFileClient
	SocketUrl       string
	TotalValues     map[string]PairTotalValue
	runListener     bool
}

func NewService(websocketClient IWebSocketClient, fileClient IFileClient, socketURL string) *Service {
	return &Service{
		WebsocketClient: websocketClient,
		FileClient:      fileClient,
		SocketUrl:       socketURL,
		TotalValues:     make(map[string]PairTotalValue),
		runListener:     true,
	}
}

func (s *Service) Run() error {
	// Initialize standard out file
	err := s.FileClient.InitFileConn()
	if err != nil {
		err = errors.Wrap(err, "issue initializing file")
		return err
	}

	s.FileClient.WriteToFile(fmt.Sprintf("\nStarted new client connection - %v", time.Now().Format("2017-09-07 17:06:06")))

	// Establish WebSocket Connection
	conn, err := s.WebsocketClient.EstablishConnection(s.SocketUrl)
	if err != nil {
		err = errors.Wrap(err, "Error Establishing Connection")
		return err
	}

	// Start Websocket Connection Listener
	go s.socketListener(conn)

	// set up the subscription message object
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
		if !s.runListener {
			break
		}
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

// evaluate: starts evaluator process that parses socket response by type.
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

// evaluateMatch: for response of 'match' type, updates the Volume Weighted Average Price
func (s *Service) evaluateMatch(response *Response) error {
	pairValues := s.TotalValues[response.ProductID]
	if len(pairValues.MatchQueue) < 200 {
		// Add resp to end of queue
		pairValues.MatchQueue = append(pairValues.MatchQueue, *response)

		// Add new price to total sum
		price, err := strconv.ParseFloat(response.Price, 64)
		if err != nil {
			err = errors.Wrap(err, "error parsing sub 200 response.Price")
		}
		pairValues.TotalSum += price
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

	// Evaluate and save new average by dividing totalSum by count of total responses
	pairValues.VolumeWeightedMovingAverage = pairValues.TotalSum / float64(len(pairValues.MatchQueue))

	// Replace the old product ID dictionary with the newly evaluated one
	s.TotalValues[response.ProductID] = pairValues

	// Write new Volume Weighted Average Price to file
	output := fmt.Sprintf("Product ID: %v, Total Count: %v, New Price: %v, VWAP: %v", response.ProductID, len(pairValues.MatchQueue), response.Price, math.Floor(pairValues.VolumeWeightedMovingAverage*100)/100)
	s.FileClient.WriteToFile(output)
	return nil
}
