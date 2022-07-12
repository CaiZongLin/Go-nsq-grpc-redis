package main

import (
	"autoSell/model"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	"github.com/nsqio/go-nsq"
)

type myMessageHandler2 struct{} //consumer C

func initNsqClient() {
	//建立Consumer
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer("GONSQ_TOPIC2", "channel", config)
	if err != nil {
		log.Fatalf("Failed to create consumer, %v", err)
	}

	//新增handler處理收到訊息時動作
	consumer.AddHandler(&myMessageHandler2{})

	//連線NSQD
	err = consumer.ConnectToNSQD("127.0.0.1:4150")
	if err != nil {
		log.Fatal(err)
	}

	//卡住，不要讓main.go執行完就結束
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	consumer.Stop()
}

func NewClient() *redis.Client { // 實體化redis.Client 並返回實體的位址
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return client
}

func (h *myMessageHandler2) HandleMessage(m *nsq.Message) error {

	var cOrder model.Order

	err := json.Unmarshal(m.Body, &cOrder)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	client := NewClient() //redis

	err = client.RPush(cOrder.Customer, cOrder.Production).Err()
	if err != nil {
		log.Println(err)
	}

	ttl := time.Minute * time.Duration(10)

	client.Expire(cOrder.Customer, ttl)

	client.IncrBy(cOrder.Production, 1).Result()
	if err != nil {
		log.Println(err)
	}
	client.Expire(cOrder.Production, ttl)

	return nil
}

func main() {
	initNsqClient()
}
