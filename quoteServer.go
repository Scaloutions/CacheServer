package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

const (
	QUOTE_SERVER_API = "quoteserve.seng"
	PORT             = "4444"
	CONNECTION_TYPE  = "tcp"
)

type Quote struct {
	Price float64
	Stock string
	// UserId    string
	Timestamp int64
	CryptoKey string
}

func getQuoteFromQS(userid string, stock string) (Quote, error) {

	// Mock QuoteServer hit for local testing
	testMode, _ := strconv.ParseBool(os.Getenv("DEV_ENVIRONMENT"))
	if testMode {
		r := rand.New(rand.NewSource(getCurrentTs()))

		return Quote{
			Price: r.Float64(),
			Stock: stock,
			// UserId:    userid,
			Timestamp: getCurrentTs(),
			CryptoKey: "PXdxruf7H5p9Br19Si5hq",
		}, nil
	}

	quote := Quote{}

	conn, err := getConnection()
	if err != nil {
		return quote, err
	}

	cstr := stock + "," + userid + "\n"
	_, err = conn.Write([]byte(cstr))

	if err != nil {
		return quote, err
	}

	buff := make([]byte, 1024)
	len, err := conn.Read(buff)

	if err != nil {
		glog.Error("Error reading data from the Quote Server")
		return quote, errors.New("Error reading the Quote.")
	}

	response := string(buff[:len])
	glog.Info("Got back from Quote server: ", response)

	quoteArgs := strings.Split(response, ",")

	// Returns: quote,sym,userid,timestamp,cryptokey
	price, err := strconv.ParseFloat(quoteArgs[0], 64)
	if err != nil {
		glog.Error("Cannot parse QS stock price into float64 ", quoteArgs[0])
		return quote, errors.New("Error parsing the Quote.")
	}

	timestamp, err := strconv.ParseInt(quoteArgs[3], 10, 64)
	if err != nil {
		glog.Error("Cannot parse QS timestamp into int64 ", quoteArgs[3])
		return quote, errors.New("Error parsing the Quote.")
	}
	conn.Close()

	return Quote{
		Price: price,
		Stock: quoteArgs[1],
		// UserId:    quoteArgs[2],
		Timestamp: timestamp,
		CryptoKey: strings.TrimSpace(quoteArgs[4]),
	}, nil

}

func getConnection() (net.Conn, error) {
	glog.Info("Connecting to the quote server... ")
	url := QUOTE_SERVER_API + ":" + PORT
	conn, err := net.Dial(CONNECTION_TYPE, url)

	if err != nil {
		fmt.Print("Error connecting to the Quote Server: somthing went wrong :(")
		return nil, errors.New("Cannot establish connection with the Quote Server")
	}
	return conn, nil
}
