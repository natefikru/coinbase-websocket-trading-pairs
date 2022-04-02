# ZeroHash Programming Exercise
#### Nathnael Fikru 

This is a backend process that initiates a web socket listener on the Coinbase Matching engine and maintains the volume weighted moving average of the last 200 responses per trading pair.

## Project Overview
This API uses the gorilla/websocket framework to run its websocket related commands. This process primarily uses the public [coinbase websocket api](https://docs.cloud.coinbase.com/exchange/docs/websocket-overview) as its point of web integration. 

This project is split into 4 different sections:
- the Application Service 
- the Config Utility
- the File writer Client
- the Websocket Client

The config utility is a simple utility that loads the configuration values for the app into a config struct to be referenced outside of the Utility. The FileClient allows for creation of files for simple writing purposes. The websocket client allows the user to create a connection to a socket, send/read messages to/from that client.

The Application Service is what executes the core functions for this application. Once the clients are initialized in the `main.go` file, they are passed in as dependencies of the service application. the main service creates the file that it will write to, then establishes a connection to the web socket. Once the Connection is set up, it will initialize the socket listener, then send a request message to the socket channel every couple of seconds to maintain the subscription. 

As messages come in, the listener loads data into a response struct, checks its type and if it has a type of `match`, it will evaluate the match response. To evaluate the Volume Weighted Average Price, we need to maintain a queue of responses with the earliest responses at the front. Because we want to maintain a maximum queue of length 200, when the queue reaches that number, instead of adding more response values to the queue, the function will pop the earliest queue response, subtract that response's price from the maintained total sum, then add the new response to the end of the queue as well as add its price to the maintained total sum.

the Volume weighted average price is evaluated by taking the maintained total sum of the queue and dividing it by the length of the queue. Each response message outcome is then printed to the specified file.
## Configuration
The `CoinbaseSocketURL` and log `FileName` can be configured within the `util/config.go` file
## How to Run
To run this process, first install the required go modules 
```
go mod download 
```

After the modules have been downloaded, run
```
go run main.go
```

## How to Run Tests
All of the tests for this project have been written within the `/service` directory. First install the required go modules if they weren't already downloaded via 
```
go mod download 
``` 

To run the tests, navigate to the `/service` folder and run 
```
go test
```
This will execute the test suite.

## Things I would do differently
### Secrets
The Coinbase websocket api optionally supports a 2 layer authentication that supports not just the projectID as a key but also a secret key as well. I opted in to not use the secret key for authentication because it seemed a bit of overkill in this specific situation. In a prod environment, I would likely be deploying this app into the cloud via a docker container or some sort of pre-determined environment. With that being said, I would write out a mechanism for fetching a hidden secret key that is registered to the os.Environment. This is very easily done within the dockerfile definition for example.