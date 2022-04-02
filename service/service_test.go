package service

import (
	"testing"

	"github.com/golang/mock/gomock"
	websocket "github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

func TestRun(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockWebSocketClient := NewMockIWebSocketClient(ctrl)
	mockFileClient := NewMockIFileClient(ctrl)

	service := NewService(mockWebSocketClient, mockFileClient, "")
	service.runListener = false

	tests := map[string]struct {
		fileConnError      bool
		establishConnError bool
		writeMessageError  bool
		resp               error
	}{
		"Valid Run() call": {
			fileConnError:      false,
			establishConnError: false,
			writeMessageError:  false,
			resp:               nil,
		},
		"InitFileConn returns error": {
			fileConnError:      true,
			establishConnError: false,
			writeMessageError:  false,
			resp:               errors.Wrap(errors.New(""), "issue initializing file"),
		},
		"EstablishConnection returns error": {
			fileConnError:      false,
			establishConnError: true,
			writeMessageError:  false,
			resp:               errors.Wrap(errors.New(""), "Error Establishing Connection"),
		},
		"WriteMessageToSocketConnInterval returns error": {
			fileConnError:      false,
			establishConnError: false,
			writeMessageError:  true,
			resp:               errors.Wrap(errors.New(""), "Error Executing Connection"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			testConn := websocket.Conn{}
			if test.fileConnError {
				mockFileClient.EXPECT().InitFileConn().Return(errors.New("")).Times(1)
			} else {
				mockFileClient.EXPECT().InitFileConn().Return(nil).Times(1)
				mockFileClient.EXPECT().WriteToFile(gomock.Any()).Times(1)
				if test.establishConnError {
					mockWebSocketClient.EXPECT().EstablishConnection(gomock.Any()).Return(&testConn, errors.New("")).Times(1)
				} else {
					mockWebSocketClient.EXPECT().EstablishConnection(gomock.Any()).Return(&testConn, nil).Times(1)
					if test.writeMessageError {
						mockWebSocketClient.EXPECT().WriteMessageToSocketConnInterval(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("")).Times(1)

					} else {
						mockWebSocketClient.EXPECT().WriteMessageToSocketConnInterval(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
					}
				}
			}

			resp := service.Run()
			if test.resp != nil && test.resp.Error() != resp.Error() {
				t.Errorf("Expected response %v, received %v for test %s", test.resp, resp, name)
			}
		})
	}
}

func TestEvaulateMatch(t *testing.T) {
	testProductID := "testProductID"
	ctrl := gomock.NewController(t)
	mockFileClient := NewMockIFileClient(ctrl)
	service := NewService(nil, mockFileClient, "")

	tests := map[string]struct {
		queueCount int
		response   *Response
		newVWMA    float64
		resp       error
	}{
		"no initial values in Queue": {
			queueCount: 0,
			response: &Response{
				ProductID: testProductID,
				Price:     "11",
			},
			newVWMA: float64(11),
			resp:    nil,
		},
		"10 values in Queue": {
			queueCount: 10,
			response: &Response{
				ProductID: testProductID,
				Price:     "21",
			},
			newVWMA: float64(11),
			resp:    nil,
		},
		"100 values in Queue": {
			queueCount: 100,
			response: &Response{
				ProductID: testProductID,
				Price:     "111",
			},
			newVWMA: float64(11),
			resp:    nil,
		},
		"200 values in Queue": {
			queueCount: 200,
			response: &Response{
				ProductID: testProductID,
				Price:     "210",
			},
			newVWMA: float64(11),
			resp:    nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			service.TotalValues[testProductID] = PairTotalValue{}
			mockFileClient.EXPECT().WriteToFile(gomock.Any()).Times(1)

			pairValues := service.TotalValues[testProductID]
			for i := 0; i < test.queueCount; i++ {
				pairValues.MatchQueue = append(pairValues.MatchQueue, Response{Price: "10"})
				pairValues.TotalSum += float64(10)
			}
			service.TotalValues[testProductID] = pairValues

			resp := service.evaluateMatch(test.response)
			if test.resp != nil && test.resp.Error() != resp.Error() {
				t.Errorf("Expected response %v, received %v for test %s", test.resp, resp, name)
			}

			if service.TotalValues[testProductID].VolumeWeightedMovingAverage != test.newVWMA {
				t.Errorf("Expected vwma %v, received %v for test %s", test.newVWMA, service.TotalValues[testProductID].VolumeWeightedMovingAverage, name)
			}
		})
	}
}
