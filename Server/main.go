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
	error  chan error
}

type Server struct {
	Connection []*Connection
	chat.UnimplementedChatServiceServer
}

func (s *Server) Publish(ctx context.Context, msg *chat.Message) (*chat.Message, error) {
	log.Print("__________________________________")
	log.Printf("Server received published message: '%s', (%d) from client: %s", 
		msg.GetBody(), 
		msg.GetUser().GetTimestamp(), 
		msg.GetUser().GetName())
		
	var reply = msg
	//TODO
	reply.GetUser().Timestamp += 1
	s.Broadcast(ctx, msg)
	return reply, nil
}

func (s *Server) Broadcast(ctx context.Context, msg *chat.Message) (*chat.Close, error) {
	wait := sync.WaitGroup{}
	done := make(chan int)

	for _, conn := range s.Connection {
		wait.Add(1)

		go func(msg *chat.Message, conn *Connection) {
			defer wait.Done()

			if conn.active {

				err := conn.stream.Send(msg)
				log.Printf("Broadcasting message to: %s", conn.id)
				if err != nil {
					conn.active = false
					conn.error <- err
				}
			}
		}(msg, conn)
	}

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
	return &chat.Close{}, nil
}

func (s *Server) OpenConnection(connect *chat.Connect, stream chat.ChatService_OpenConnectionServer) error {
	conn := &Connection{
		stream: stream,
		active: true,
		id:     connect.User.Name,
		error:  make(chan error),
	}

	s.Connection = append(s.Connection, conn)

	joinMessage := chat.Message{
		Body: connect.GetUser().GetName() + " has succesfully connected",
		User: &chat.User{Id: 1000, Name: "Server", Timestamp: connect.GetUser().GetTimestamp()+1},
	}
	log.Print("__________________________________")
	log.Printf("%v: %s,(%d)", joinMessage.GetUser().GetName(), joinMessage.GetBody(), joinMessage.GetUser().GetTimestamp())
	s.Broadcast(context.Background(), &joinMessage)

	return <-conn.error
}

func main() {
	var connections []*Connection
	s := &Server{
		Connection:                     connections,
		UnimplementedChatServiceServer: chat.UnimplementedChatServiceServer{},
	}

	grpcServer := grpc.NewServer()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen port: %v", err)
	}

	log.Printf("server listening at %v", lis.Addr())

	chat.RegisterChatServiceServer(grpcServer, s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed serve server: %v", err)
	}
}
