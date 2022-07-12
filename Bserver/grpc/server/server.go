package main

import (
	pb "autoSell/Bserver/grpc/pb"
	"log"
	"net"
	"strconv"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedServiceServerServer
}

func (s *server) RedisUserBuyCount(ctx context.Context, in *pb.BuyRequest) (*pb.StatusReply, error) {
	client := NewClient()
	val, err := client.LLen(in.Customer).Result()
	if err != nil {
		log.Println(err)
	}
	log.Println(val)
	if val >= 10 {
		return &pb.StatusReply{Code: 0, Status: "BuyCount > 10"}, nil
	}
	return &pb.StatusReply{Code: 1, Status: "BuyCount < 10"}, nil
}

func (s *server) RedisProductionBuyCount(ctx context.Context, in *pb.BuyRequest) (*pb.StatusReply, error) {

	client := NewClient()
	val, err := client.Get(in.Production).Result()
	if err != nil {
		log.Println(err)
	}

	result, err := strconv.ParseInt(val, 10, 64)
	log.Println(result)
	if result >= 3 {
		return &pb.StatusReply{Code: 0, Status: "Product_Buy_count > 3"}, nil
	}
	return &pb.StatusReply{Code: 1, Status: "Product_Buy_count < 3"}, nil
}

func NewClient() *redis.Client { // 實體化redis.Client 並返回實體的位址
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return client
}

func initServer() {
	log.Println("Master node start")
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Can't listen on port %v", err.Error())
	}
	s := grpc.NewServer()
	pb.RegisterServiceServerServer(s, &server{})
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Can't init gRPC server：%v", err.Error())
	}
}

func main() {
	initServer()
}
