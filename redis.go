package main

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/glog"
)

var (
	Pool         *redis.Pool
	CACHE_SERVER string
)

func InitializeRedisCache() {
	// Get Cache server address
	testMode, _ := strconv.ParseBool(os.Getenv("DEV_ENVIRONMENT"))
	if testMode {
		CACHE_SERVER = os.Getenv("REDIS_CACHE_LOCAL")
	} else {
		CACHE_SERVER = os.Getenv("REDIS_CACHE_PROD")
	}

	host := CACHE_SERVER + ":6379"
	glog.Info(">>>>>>> Connecting to Redis through ", host)
	Pool = newRedisPool(host)
}

type RedisQuote struct {
	Price float64
	Stock string
	// UserId    string
	// Timestamp int64
	CryptoKey string
}

func newRedisPool(server string) *redis.Pool {

	return &redis.Pool{
		MaxIdle:     80,
		MaxActive:   10000,
		IdleTimeout: 30 * time.Second,

		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", server)
			if err != nil {
				glog.Error(err)
				return nil, err
			}

			return conn, err
		},
	}
}

func SetToCache(qt Quote) error {

	c := Pool.Get()
	defer c.Close()

	q := RedisQuote{
		Price:     qt.Price,
		Stock:     qt.Stock,
		CryptoKey: qt.CryptoKey,
	}

	quote, err := json.Marshal(q)

	if err != nil {
		glog.Error("Can't marshal the Quote object ", q)
		return errors.New("Can't marshal the Quote object.")
	}

	c.Send("MULTI")
	c.Send("SET", q.Stock, string(quote))
	c.Send("EXPIRE", q.Stock, "60")
	_, err = c.Do("EXEC")

	if err != nil {
		//couldnt set to redis
		glog.Error("Could not SET REDIS value ", err)
		return err
	}

	return nil
}

func GetFromCache(stock string) (Quote, error) {
	c := Pool.Get()
	defer c.Close()

	q := Quote{}

	val, err := redis.String(c.Do("GET", stock))

	bytes := ([]byte)(val)

	err = json.Unmarshal(bytes, &q)

	if err != nil {
		glog.Error("Error unmarshling Redis quote ", err)
		return Quote{}, err
	}

	q.Timestamp = getCurrentTs()

	glog.Info("Returnig quote from Redis Cache: ", q)

	return q, nil
}
