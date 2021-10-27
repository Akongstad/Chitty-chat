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
var wait *sync.WaitGroup

func init(){
	wait = &sync.WaitGroup{}
}

func connect(user *chat.User) (error){
	var streamError error

	stream, err := client.OpenConnection(context.Background(), &chat.Connect{
		User: user,
		Active: true,
	})

	if err != nil {
		log.Fatalf("Connect failed: %v", err)
		return err
	}

	wait.Add(1)
	go func(str chat.ChatService_OpenConnectionClient){
		defer wait.Done()

		for{
			msg, err := str.Recv()
			if(err != nil){
				log.Fatalf("Error reading message, %v", err)
				streamError = err
				break
			}

			log.Printf("%v: %s,(%d)", msg.GetUser().GetName(), msg.GetBody(), msg.GetUser().GetTimestamp())
		}
	}(stream)
	
	return streamError
} 

func main() {
	
	//init channel
	done := make(chan int)
	
	//Get User info
	clientName := flag.String("U", "Anonymous", "ClientName")
	flag.Parse()
	userId := rand.Intn(999)
	clientUser := &chat.User{
		Id: int32(userId),
		Name: *clientName,
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(port, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	client = chat.NewChatServiceClient(conn)
	
	//Create stream
	
	log.Println(*clientName, " Connecting")
	connect(clientUser)

	//Send messages
	wait.Add(1)

	go func(){
		defer wait.Done()

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan(){
			msg := &chat.Message{
				Body: scanner.Text(),
				User: clientUser,
			}

			_, err := client.Publish(context.Background(), msg)
			if err != nil{
				log.Printf("Error publishing Message: %v", err)
				break
			}
		}
	}()

	go func(){
		wait.Wait()
		close(done)
	}()

	<- done
}
