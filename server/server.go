package server

import (
	"autoSell/model"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	host     = "127.0.0.1"
	database = "vending"
	root     = "root"
	root_pwd = "s850429s"
	port     = 3306
)

type Product struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Inventory   int64   `json:"inventory"`
	Status      int64   `json:"status"`
	Update_time time.Time
}

type Customer struct {
	ID           int64   `json:"id"`
	Name         string  `json:"name"`
	Total_price  float64 `json:"price"`
	Buy_count    int64   `json:"buy_count"`
	Updated_time time.Time
}

func (Product) TableName() string {
	return "product_info"
}

func (Customer) TableName() string {
	return "customer_info"
}

func GetProductInfo(productName string) (f0 Product, err error) {
	status, err := Search(productName)
	if status != 1 {
		log.Printf("Product not found")
		return f0, err
	}
	var product Product
	addr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True", root, root_pwd, host, port, database)
	conn, err := gorm.Open(mysql.Open(addr), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
		return f0, err
	}
	if err2 := conn.Where("name=?", productName).Find(&product).Error; err2 != nil {
		log.Fatalln(err)
		return f0, err
	}
	return product, nil
}

func GetCustomerInfo(CustomerName string) (f0 Customer, err error) {

	var customer Customer
	addr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True", root, root_pwd, host, port, database)
	conn, err := gorm.Open(mysql.Open(addr), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
		return f0, err
	}
	if err2 := conn.Where("name=?", CustomerName).Find(&customer).Error; err2 != nil {
		log.Fatalln(err)
		return f0, err
	}
	return customer, nil
}

func Search(productName string) (int, error) {
	var product Product
	addr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True", root, root_pwd, host, port, database)
	conn, err := gorm.Open(mysql.Open(addr), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
		return 0, err
	}
	if err2 := conn.Where("name=?", productName).Find(&product).Error; err2 != nil {
		log.Fatal(err)
	}

	if product.Name == "" {
		return -1, nil
	}

	return 1, nil
}

func InsertCustomerInfo(customer_info Customer) (int, error) {
	addr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True", root, root_pwd, host, port, database)
	conn, err := gorm.Open(mysql.Open(addr), &gorm.Config{})
	if err != nil {
		return -1, err
	}
	if errCreate := conn.Debug().Create(&customer_info).Error; errCreate != nil {
		return -1, errCreate
	}

	return 1, nil
}

func UpdateCustomerInfo(customer_info Customer, product_price float64) (int, error) {
	addr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True", root, root_pwd, host, port, database)
	conn, err := gorm.Open(mysql.Open(addr), &gorm.Config{})
	if err != nil {
		return -1, err
	}
	log.Println(customer_info.Total_price, product_price)
	new_info := Customer{
		Name:         customer_info.Name,
		Total_price:  customer_info.Total_price + product_price,
		Buy_count:    customer_info.Buy_count + 1,
		Updated_time: time.Now(),
	}

	if errModify := conn.Model(&Customer{}).Where("id = ?", customer_info.ID).Updates(new_info).Error; errModify != nil {
		return -1, errModify
	}
	return 1, nil
}

func RedisUserBuyCount(in model.CustomerInfo) (int, error) {

	client := NewClient()
	val, err := client.LLen(in.Customer).Result()
	if err != nil {
		log.Println(err)
	}
	if val >= 10 {
		return 0, nil
	}
	return 1, nil
}

func RedisProductionBuyCount(in model.CustomerInfo) (int, error) {

	client := NewClient()
	val, err := client.Get(in.Production).Result()
	if err != nil {
		log.Println(err)
	}

	result, err := strconv.ParseInt(val, 10, 64)
	if result >= 3 {
		return 0, nil
	}
	return 1, nil
}

func NewClient() *redis.Client { // 實體化redis.Client 並返回實體的位址
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return client
}
