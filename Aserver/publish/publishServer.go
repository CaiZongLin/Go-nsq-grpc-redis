package main

import (
	pb "autoSell/Bserver/grpc/pb"
	"autoSell/model"
	useful "autoSell/server"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nsqio/go-nsq"
	"google.golang.org/grpc"
)

func initNsqServerAndPublishToConsumerC(f1 model.CustomerInfo) {

	config := nsq.NewConfig()
	producer, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("TO 1")
	log.Println(f1)

	result, err := json.Marshal(&f1)

	topicName := "GONSQ_TOPIC"
	//發佈到定義好的topic
	err = producer.Publish(topicName, result)

	if err != nil {
		log.Fatal(err)
	}

	producer.Stop()
}

func initNsqServerAndPublishToConsumerB(f1 model.CustomerInfo) {

	config := nsq.NewConfig()

	producer, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		log.Fatal(err)
	}

	result, err := json.Marshal(&f1)

	topicName := "GONSQ_TOPIC2"
	//發佈到定義好的topic
	err = producer.Publish(topicName, result)

	if err != nil {
		log.Fatal(err)
	}

	producer.Stop()
}

func grpcConnect() pb.ServiceServerClient {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalln(err)
	}

	client := pb.NewServiceServerClient(conn)
	return client
}

func main() {
	conn()
}

func conn() {
	ln, err := net.Listen("tcp", ":1450")

	defer ln.Close()

	if err != nil {
		log.Fatalln("listen error: ", err)
	}

	log.Println("Server ")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalln("accept error: ", err)
			break
		}

		go func(conn net.Conn) {
			defer conn.Close()

			req := make([]byte, 1024)
			total, err := conn.Read(req)
			fmt.Println("已接收到", string(req))

			var data model.CustomerInfo

			err = json.Unmarshal(req[:total], &data)
			if err != nil {
				log.Fatalln(err)
			}

			grpc := grpcConnect()

			search := &pb.BuyRequest{
				Customer:   data.Customer,
				Production: data.Production,
			}

			UBuyCount, err := grpc.RedisUserBuyCount(context.Background(), search)
			if err != nil {
				return
			}

			PBuyCount, err := grpc.RedisProductionBuyCount(context.Background(), search)
			if err != nil {
				return
			}

			if UBuyCount.Code != 1 {
				data.Comment = fmt.Sprintf("客人 %s購買次數在十分鐘內超過10次，請稍後再嘗試購買", data.Customer)
				data.Status = -1
				sendClientData, err := json.Marshal(data)
				if err != nil {
					log.Printf("sendClientData failed: %v", err)
				}
				conn.Write(sendClientData)
				return
			}

			if PBuyCount.Code != 1 {
				data.Comment = fmt.Sprintf("商品 %s在十分鐘內購買已達三次，請選擇其他商品", data.Production)
				data.Status = -1

				sendClientData, err := json.Marshal(data)
				if err != nil {
					log.Printf("sendClientData failed: %v", err)
				}

				conn.Write(sendClientData)
				return
			}

			var product_info useful.Product
			product_info, err = useful.GetProductInfo(data.Production) //get product price from DB.product_info
			if err != nil {
				log.Println(err)
				return
			}

			data.Price = product_info.Price
			data.Comment = "購買成功"
			data.Status = 0

			searchCustomer, err := useful.GetCustomerInfo(data.Customer) // get customer's total price from DB.customer_info
			if err != nil {
				log.Println(err)
				return
			}

			if searchCustomer.Total_price > 10000 {
				data.Price *= 0.9
			}

			initNsqServerAndPublishToConsumerC(data) //寫入DB
			initNsqServerAndPublishToConsumerB(data) //寫入redis

			//回傳給client端
			sendClientData, err := json.Marshal(data)
			if err != nil {
				log.Printf("sendClientData failed: %v", err)
			}

			conn.Write(sendClientData)

		}(conn)

	}

}
