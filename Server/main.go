package main

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/Akongstad/Chitty-chat/chat"

	"google.golang.org/grpc"
)

const (
	port = ":9000"
)

type Connection struct {
	stream chat.ChatService_OpenConnectionServer
	id     string
	active bool
	err    chan error
}

type Server struct {
	Connection []*Connection
	chat.UnimplementedChatServiceServer
}

func (s *Server) Publish(ctx context.Context, msg *chat.Message) (*chat.Message, error) {
	log.Printf("Server received published message %s", msg.GetBody())
	log.Printf("Published at: %s", msg.GetTimestamp())
	log.Printf("Client:", msg.GetUser())
	var reply = msg
	//TODO
	reply.Timestamp += 1
	return reply, nil
}

func (s *Server) OpenConnection(connect *chat.Connect, stream chat.ChatService_OpenConnectionServer) error {
	conn := &Connection{
		active: true,
		stream: stream,
		id:     connect.User.Name,
		err:    make(chan error),
	}
	s.Connection = append(s.Connection, conn)

	return <-conn.err
}

func (s *Server) Broadcast(ctx context.Context, msg *chat.Message) (*chat.Close, error) {
	wait := sync.WaitGroup{}
	done := make(chan int)

	for _, conn := range s.Connection {
		wait.Add(1)

		go func(msg *chat.Message, conn *Connection) {
			defer wait.Done()
			if conn.active {
				//passes message back to the client
				err := conn.stream.Send((msg))
				//sæt ind i log her

				if err != nil {
					conn.active = false
					conn.err <- err
				}

			}
		}(msg, conn)
	}

	go func() {
		//makes sure that the wait group will wait for other goroutines to exit
		wait.Wait()
		close(done)
	}()
	<-done
	return &chat.Close{}, nil

}

func main() {
	var connections []*Connection
	s := &Server{
		Connection:                     connections,
		UnimplementedChatServiceServer: chat.UnimplementedChatServiceServer{},
	}
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen port 9000: %v", err)
	}

	grpcServer := grpc.NewServer()

	chat.RegisterChatServiceServer(grpcServer, s)

	log.Printf("server listening at %v", lis.Addr())

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed serve server: %v", err)
	}
}
