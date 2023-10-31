package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	proto "chitChat/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Same principle as in client. Flags allows for user specific arguments/values
var clientsName = flag.String("name", "default", "Senders name")
var clientsPort = flag.String("port", "8080", "Senders name")
var serverPort = "5400";

var server proto.ChatRoomClient  
var ServerConn *grpc.ClientConn 


func main() {
	//parse flag/arguments
	flag.Parse()

	fmt.Println("--- CLIENT APP ---")

	//log to file instead of console
	//f := setLog()
	//defer f.Close()

	//connect to server and close the connection when program closes
	fmt.Println("--- join Server ---")
	ConnectToServer()
	defer ServerConn.Close()

	//start the biding
	parseInput()
}


// connect to server
func ConnectToServer() {

	//dial options
	//the server is not using TLS, so we use insecure credentials
	//(should be fine for local testing but not in the real world)
	opts := []grpc.DialOption {
		grpc.WithBlock(), 
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	//dial the server, with the flag "server", to get a connection to it
	log.Printf("client %s: Attempts to dial on port %s\n", *clientsName, serverPort)
	conn, err := grpc.Dial(fmt.Sprintf(":%s", serverPort), opts...)
	if err != nil {
		log.Printf("Fail to Dial : %v", err)
		return
	}

	// makes a client from the server connection and saves the connection
	// and prints rather or not the connection was is READY
	server = proto.NewChatRoomClient(conn)
	ServerConn = conn
	log.Println("the connection is: ", conn.GetState().String())
}


func parseInput() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Write 'join' to join the chatroom")
	fmt.Println("--------------------")

	//Infinite loop to listen for clients input.
	for {
		fmt.Print("-> ")

		//Read input into var input and any errors into err
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		
		input = strings.TrimSpace(input) //Trim input

		if !conReady(server) {
			log.Printf("Client %s: something was wrong with the connection to the server :(", *clientsName)
			continue
		}

		
		if input == "join" {
			go joinChat();
		} else if input == "leave" {
			go LeaveChat();
		} else {
			go SendMessage(input);
		}
	}
}

func joinChat() {
	//create amount type
	client := &proto.Client{
		Name: *clientsName,
		Port: *clientsPort,
	}

	//Make gRPC call to server with amount, and recieve acknowlegdement back.
	chatStream, err := server.JoinChat(context.Background(), client)
	if err != nil {
		log.Printf("Client %s: no response from the server, attempting to reconnect", *clientsName)
		log.Println(err)
	}

	for {
		chatStream.Recv()
	}
}

func LeaveChat() {
	client := &proto.Client{
		Name: *clientsName,
		Port: *clientsPort,
	}
	// get a stream to the server
	ack, err := server.LeaveChat(context.Background(), client)
	if err != nil {
		log.Println(err)
		return
	} 
	if ack.Succes {
		log.Printf("You have succesfully left the chatroom")
	}
}

func SendMessage(message string) {

	chat:= &proto.Chat{
		Message: message,
		Name: *clientsName,
		Time: time.Now().String(),
	}
	// get a stream to the server
	ack, err := server.SendMessage(context.Background(), chat)
	if err != nil {
		log.Println(err)
		return
	} 
	if ack.Succes {
		log.Printf("You have succesfully left the chatroom")
	}
}

// Function which returns a true boolean if the connection to the server is ready, and false if it's not.
func conReady(s proto.ChatRoomClient) bool {
	return ServerConn.GetState().String() == "READY"
}

// sets the logger to use a log.txt file instead of the console
func setLog() *os.File {
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}