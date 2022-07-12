package main

import (
	"autoSell/model"
	useful "autoSell/server"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nsqio/go-nsq"
)

type myMessageHandler struct{} //consumer B

func initNsqClient() {
	//建立Consumer
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer("GONSQ_TOPIC", "channel", config)
	if err != nil {
		log.Fatalf("Failed to create consumer, %v", err)
	}

	//新增handler處理收到訊息時動作
	consumer.AddHandler(&myMessageHandler{})

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

func (h *myMessageHandler) HandleMessage(m *nsq.Message) error {
	var cOrder model.Order

	err := json.Unmarshal(m.Body, &cOrder)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	log.Println(cOrder.Customer, cOrder.Price, cOrder.Production)

	customer_info, err := useful.GetCustomerInfo(cOrder.Customer) //get customer_info from DB.customer_info
	if customer_info.Name == "" {                                 //空代表沒有資料，所以insert
		new := useful.Customer{
			Name:         cOrder.Customer,
			Total_price:  cOrder.Price,
			Buy_count:    1,
			Updated_time: time.Now(),
		}

		status, err := useful.InsertCustomerInfo(new)
		if err != nil {
			log.Println(status, err)
		}
	}

	status, err := useful.UpdateCustomerInfo(customer_info, cOrder.Price) //有資料就update
	if err != nil {
		log.Println(status, err)
	}

	return nil
}

func main() {
	initNsqClient()
}
