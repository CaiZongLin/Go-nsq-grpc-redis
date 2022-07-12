package main

import (
	"autoSell/model"
	"encoding/json"
	"fmt"
	"log"
	"net"
)

func main() {

	conn, err := net.Dial("tcp", "localhost:1450")
	defer conn.Close()

	if err != nil {
		log.Fatalln(err)
		return
	}

	client := model.CustomerInfo{Customer: "Lex", Production: "肉燥飯"}

	fmt.Printf("購買資料: 購買人: %s,商品: %s\n", client.Customer, client.Production)

	sendClientData, err := json.Marshal(client)
	//發送訊息給Server
	conn.Write(sendClientData)

	//接收伺服器回傳的訊息
	res := make([]byte, 1024)

	total, err := conn.Read(res)
	if err != nil {
		log.Fatalln(err)
		return
	}

	var result model.CustomerInfo

	err = json.Unmarshal(res[:total], &result)
	if err != nil {
		log.Fatalln(err)
		return
	}

	if result.Status == -1 {
		fmt.Printf("購買失敗，原因: %s", result.Comment)
		return
	}

	fmt.Printf("購買成功，最終購買資料: 購買人: %s, 商品: %s, 價格:%.2f ", result.Customer, result.Production, result.Price)

}
