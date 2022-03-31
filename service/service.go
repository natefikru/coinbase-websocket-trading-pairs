package service

import (
	"coinbase-websocket-trading-pairs/util"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

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

func (s *Service) getProductIDs() []string {
	return []string{
		"BTC-USD",
		"ETH-USD",
		"ETH-BTC",
	}
}

func (s *Service) Run() error {
	conn, err := s.WebsocketClient.EstablishConnection(s.Config.CoinbaseSocketURL)
	if err != nil {
		err = errors.Wrap(err, "Error Establishing Connection")
		return err
	}

	go s.socketMatchListener(conn)

	request := s.setUpRequest()
	newRequestByes, err := json.Marshal(request)
	if err != nil {
		err = errors.Wrap(err, "Error marshalling request")
		return err
	}

	err = s.WebsocketClient.WriteMessageToSocketConn(conn, newRequestByes)
	if err != nil {
		err = errors.Wrap(err, "Error Executing Connection")
		return err
	}
	return nil
}

func (s *Service) socketMatchListener(conn *websocket.Conn) {
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error in reading messags from connection: ", err)
			return
		}
		var response Response
		err = json.Unmarshal(msgBytes, &response)
		if err != nil {
			log.Println("Error unmarshalling match response:", err)
			return
		}

		go s.evaluate(&response)
	}
}

func (s *Service) setUpRequest() *SubscribeMessage {
	var channels []interface{}
	channels = append(channels, Matches)
	productIDs := s.getProductIDs()

	// Initialize trading pairs dict
	for _, key := range productIDs {
		val := PairTotalValue{}
		s.TotalValues[key] = val
	}

	return &SubscribeMessage{
		Type:       Subscribe,
		ProductIDs: productIDs,
		Channels:   channels,
	}
}

func (s *Service) evaluate(response *Response) error {
	switch response.Type {
	case Match:
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
		// add resp to end of queue
		pairValues.MatchQueue = append(pairValues.MatchQueue, *response)

		// add response.Price to pairValues.TotalSum
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
		// remove oldest response off front of queue, add new response to end of queue
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
	// save new to Average by dividing totalSum by totalCount(200)
	pairValues.Average = pairValues.TotalSum / float64(pairValues.TotalCount)
	s.TotalValues[response.ProductID] = pairValues
	fmt.Println(response.ProductID, pairValues.TotalCount, pairValues.Average)

	return nil
}
