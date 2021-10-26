package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"math/rand"
	"os"
	"sync"

	"github.com/Akongstad/Chitty-chat/chat"
	"google.golang.org/grpc"
)

const (
	port = ":9000"
)

var client chat.ChatServiceClient
var wait sync.WaitGroup

func connect(user *chat.User) (error){
	var streamError error
	log.Println(*user)
	stream, err := client.OpenConnection(context.Background(), &chat.Connect{
		User: user,
		Active: true,
	})

	if err != nil {
		log.Fatalf("Connect failed: %v", err)
		return err
	}

	wait.Add(1)

	go func(str chat.ChatService_OpenConnectionClient) {
		defer wait.Done()

		for  {
			msg, err := str.Recv()

			if err != nil {
				streamError = err
				log.Fatalf("Error reading message: %v", err)
				break
			}

			log.Printf("%v : %s\n", msg.User.GetName(), msg.GetBody())
		}
	}(stream)
	
	wait.Add(1)
	go func() {
   		defer wait.Done()
   		scanner := bufio.NewScanner(os.Stdin)
   		msgID := rand.Intn(99)
   		for scanner.Scan() {
			msg := &chat.Message{
				Id: int32(msgID),
				User: user,
				Body: scanner.Text(),
				Timestamp: 1,
			}

			_, err := client.Broadcast(context.Background(), msg)
			if err != nil {
				log.Fatalf("Error sending message: %v", err)
				break
			}
		}
	 }()
	 go func() {
	   wait.Wait()
	   //close(done)
	}()
	//<- done  
	return streamError
}

func main() {
	// Set up a connection to the server.
	clientName := flag.String("X", "Anonymous", "")
	flag.Parse();
	clientUser := &chat.User{
		Name: *clientName,
	}

	conn, err := grpc.Dial(port, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	client = chat.NewChatServiceClient(conn)
	connect(clientUser)

}