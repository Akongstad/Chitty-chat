package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

const (
	port = ":9000"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(port, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	//c := course_proto.NewCourseServiceClient(conn)
}