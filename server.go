package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/joho/godotenv"
)

type Request struct {
	UserId        string
	CommandNumber int
	Stock         string
}

type Quote struct {
	Price float64
	Stock string
	// UserId    string
	Timestamp int64
	CryptoKey string
}

func usage() {
	fmt.Println("usage: example -logtostderr=true -stderrthreshold=[INFO|WARN|FATAL|ERROR] -log_dir=[string]\n")
	flag.PrintDefaults()
}

func echoString(c *gin.Context) {
	c.String(http.StatusOK, "Cache server is UP!")
}

func getParams(c *gin.Context) Request {
	request := Request{}
	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		glog.Error("Error processing request: %s", err)
	}

	err = json.Unmarshal(body, &request)
	if err != nil {
		glog.Error("Error parsing JSON: %s", err)
	}

	return request
}

func getQuoteReq(c *gin.Context) {
	req := getParams(c)
	//check cache
	glog.Info("Processing QUOTE get request.... for: ", req)
	quote, err := getQuote(req.Stock)

	if err != nil {
		glog.Info("Responding with error: ", err)
		c.JSON(500, gin.H{
			"transaction_num": req.CommandNumber,
			"command":         "CACHE SERVER QUOTE",
			"error":           err.Error(),
		})
	} else {
		c.BindJSON(&quote)
		c.IndentedJSON(http.StatusOK, quote)
	}
}

func getQuote(stock string) (Quote, error) {
	//check cache
	glog.Info("Getting quote for ", stock)
	cacheq, err := GetFromCache(stock)

	// found stock in the cache
	if err == nil {
		glog.Info("Got QUOTE from Redis: ", cacheq)
		// log system event
		// log := getSystemEvent(transactionNum, QUOTE, userId, stock, cacheq.Price)
		// go logEvent(log)
		// glog.Info("LOGGING ######## ", log)

		// return cacheq.Price, nil
		return Quote{
			Price:     cacheq.Price,
			Stock:     cacheq.Stock,
			CryptoKey: cacheq.CryptoKey,
			Timestamp: getCurrentTs(),
		}, nil
	}

	glog.Info("Getting Quote from the QS")
	quoteObj, err := getQuoteFromQS("CacheServer", stock)
	glog.Info("Got back from the QS: >>>>>>>> ", quoteObj)

	// put it in CACHE
	glog.Info("Putting new Stock Quote into Redis Cache ", quoteObj)
	err = SetToCache(quoteObj)
	if err != nil {
		glog.Error("Error putting QUOTE into Redist cache ", quoteObj)
		return Quote{}, err
	}

	return Quote{
		Price:     quoteObj.Price,
		Stock:     quoteObj.Stock,
		CryptoKey: quoteObj.CryptoKey,
		Timestamp: quoteObj.Timestamp,
	}, nil
}

func main() {
	router := gin.Default()

	//glog initialization flags
	flag.Usage = usage
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		glog.Error("Error loading .env file")
	}

	InitializeRedisCache()

	api := router.Group("/api")
	{
		api.GET("/test", echoString)
		// api.GET("/get_quote", getQuoteReq)
		api.POST("/quote", getQuoteReq)
	}

	log.Fatal(router.Run(":9092"))

}
