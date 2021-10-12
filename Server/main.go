package main

import (
	"log"
	"net"
	"github.com/Akongstad/Chitty-chat"

	"google.golang.org/grpc"
)

const (
	port = ":9000"
)
type Server struct{

}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen port 9000: %v", err)
	}
	s := server{}
	grpcServer := grpc.NewServer()

	

	//Message .RegisterCourseServiceServer(grpcServer, &s)

	log.Printf("server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed serve server: %v", err)
	}
}