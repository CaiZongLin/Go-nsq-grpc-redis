package model

import "time"

type Product struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Inventory   string  `json:"inventory"`
	Status      int64   `json:"status"`
	Update_time time.Time
}

type CustomerInfo struct {
	Customer   string  `json:"customer"`
	Production string  `json:"production"`
	Price      float64 `json:"price"`
	Status     int64   `json:"status"`
	Comment    string  `json:"comment"`
}

type Customer struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Total_price  float64 `json:"price"`
	Buy_count    int64   `json:"buy_count"`
	Updated_time time.Time
}

type Order struct {
	Customer   string
	Production string
	Inventory  int64
	Price      float64
}
