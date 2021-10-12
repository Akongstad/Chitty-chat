package main

import (
	"log"
	"net"
	"github.com/Akongstad/Chitty-chat/chat"

	"google.golang.org/grpc"
)

const (
	port = ":9000"
)
type Server struct{
	chat.UnimplementedChatServiceServer
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen port 9000: %v", err)
	}
	s := Server{}
	grpcServer := grpc.NewServer()

	

	chat.RegisterChatServiceServer(grpcServer, &s)

	log.Printf("server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed serve server: %v", err)
	}
}