package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
)

func usage() {
	fmt.Println("usage: example -logtostderr=true -stderrthreshold=[INFO|WARN|FATAL|ERROR] -log_dir=[string]\n")
	flag.PrintDefaults()
}

func echoString(c *gin.Context) {
	c.String(http.StatusOK, "Cache server is UP!")
}

func getQuoteReq(c *gin.Context) {
	//check cache

	// get from ns

}

func getQuote(stock string) Quote {
	//check cache
	cacheq, err := GetFromCache(stock)

	// found stock in the cache
	if err == nil {
		glog.Info("Got QUOTE from Redis: ", cacheq)
		// log system event
		// log := getSystemEvent(transactionNum, QUOTE, userId, stock, cacheq.Price)
		// go logEvent(log)
		// glog.Info("LOGGING ######## ", log)

		// return cacheq.Price, nil
		return cacheq
	}

	quoteObj, err := getQuoteFromQS("CacheServer", stock)

	// put it in CACHE
	glog.Info("Putting new Stock Quote into Redis Cache ", quoteObj)
	err = SetToCache(quoteObj)
	if err != nil {
		glog.Error("Error putting QUOTE into Redist cache ", quoteObj)
	}

	return Quote{
		Price:     quoteObj.Price,
		Stock:     quoteObj.Stock,
		CryptoKey: quoteObj.CryptoKey,
		Timestamp: quoteObj.Timestamp,
	}
}

func main() {
	router := gin.Default()

	//glog initialization flags
	flag.Usage = usage
	flag.Parse()

	api := router.Group("/api")
	{
		api.GET("/test", echoString)
		// api.GET("/get_quote", getQuoteReq)
		api.POST("/quote", echoString)
	}

	log.Fatal(router.Run(":9092"))

}
