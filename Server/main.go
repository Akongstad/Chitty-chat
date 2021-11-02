package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"strconv"
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
	Name   string
	active bool
	error  chan error
}

type Server struct {
	Connections []*Connection
	chat.UnimplementedChatServiceServer
	ServerTimestamp int32
	lock sync.Mutex
}

func (s *Server) Publish(ctx context.Context, msg *chat.Message) (*chat.Message, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.ServerTimestamp = Max(s.ServerTimestamp, msg.Timestamp)

	log.Print("__________________________________")
	log.Printf("Server received published message: '%s', (%d) from client: %s",
		msg.Body,
		msg.Timestamp,
		msg.User.Name)

	s.Broadcast(ctx, msg)
	return msg, nil
}

func (s *Server) Broadcast(ctx context.Context, msg *chat.Message) (*chat.Close, error) {
	wait := sync.WaitGroup{}
	done := make(chan int)

	for _, conn := range s.Connections {
		wait.Add(1)

		go func(msg *chat.Message, conn *Connection) {
			defer wait.Done()

			if conn.active {

				err := conn.stream.Send(msg)
				log.Printf("Broadcasting message to: %s", conn.Name)
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
		id:     connect.User.Name + strconv.Itoa(int(connect.User.Id)),
		Name:   connect.User.Name,
		error:  make(chan error),
	}

	s.Connections = append(s.Connections, conn)
	s.ServerTimestamp++
	s.ServerTimestamp = Max(s.ServerTimestamp, connect.User.Timestamp)

	joinMessage := chat.Message{
		Body:      connect.GetUser().GetName() + " has succesfully connected",
		User:      &chat.User{Name: "Server"},
		Timestamp: s.ServerTimestamp,
	}
	log.Print("__________________________________")
	log.Printf("Server: %s,(%d)", joinMessage.GetBody(), joinMessage.GetTimestamp())
	s.Broadcast(context.Background(), &joinMessage)

	return <-conn.error
}

func (s *Server) CloseConnection(ctx context.Context, user *chat.User) (*chat.Close, error) {

	var deleted *Connection

	for index, conn := range s.Connections {
		if conn.id == user.Name+strconv.Itoa(int(user.Id)) {
			s.Connections = remove(s.Connections, index)
			deleted = conn
		}
	}
	if deleted == nil {
		log.Print("No such connection to close")
		return &chat.Close{}, nil
	}

	s.ServerTimestamp = Max(s.ServerTimestamp, user.Timestamp)
	leaveMessage := chat.Message{
		Body:      user.GetName() + " has disconnected",
		User:      &chat.User{Name: "Server"},
		Timestamp: s.ServerTimestamp,
	}

	log.Print("__________________________________")
	log.Printf("Server: %s,(%d)", leaveMessage.GetBody(), leaveMessage.GetTimestamp())
	s.Broadcast(context.Background(), &leaveMessage)

	return &chat.Close{}, nil
}

func remove(slice []*Connection, i int) []*Connection {
	slice[i] = slice[len(slice)-1]
	return slice[:len(slice)-1]
}

func main() {
	var connections []*Connection
	s := &Server{
		Connections:                    connections,
		UnimplementedChatServiceServer: chat.UnimplementedChatServiceServer{},
		ServerTimestamp:                0,
		lock: sync.Mutex{},
	}
	// If the file doesn't exist, create it or append to the file. For append functionality : os.O_APPEND
	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	//Create multiwriter
	multiWriter := io.MultiWriter(os.Stdout, file)
	log.SetOutput(multiWriter)

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

func Max(a int32, b int32) int32 {
	if b > a {
		return b
	}
	return a
}
