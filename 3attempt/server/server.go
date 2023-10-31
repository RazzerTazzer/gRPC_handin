package main

import (
	proto "chitChat/grpc"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
)

// Struct that will be used to represent the Server.
type Server struct {
	proto.UnimplementedChatRoomServer        // You need this line if you have a server
	name                             string // Not required but useful if you want to name your server
	port                             string // Not required but useful if your server needs to know what port it's listening to
	clients					[]*ClientStream

	mutex          sync.Mutex // used to lock the server to avoid race conditions.
}

type ClientStream struct {
    client *proto.Client
    stream proto.ChatRoom_JoinChatServer
}




func main() {
	// f := setLog() //uncomment this line to log to a log.txt file instead of the console
	// defer f.Close()
	
	// Get the port from the command line when the server is run
	flag.Parse()

	// Create a server struct
	server := &Server{
		name: "Server",
		port: "5400",
	}

	// Start the server
	go startServer(server)

	// Keep the server running until it is manually quit
	for {

	}
}

func startServer(server *Server) {

	// Create a new grpc server
	grpcServer := grpc.NewServer()

	// Make the server listen at the given port (convert int port to string)
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", server.port))

	if err != nil {
		log.Fatalf("Could not create the server %v", err)
	}
	log.Printf("Started server at port: %d\n", server.port)

	// Register the grpc server and serve its listener
	proto.RegisterChatRoomServer(grpcServer, server)
	serveError := grpcServer.Serve(listener)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
}

func (s *Server) JoinChat(in *proto.Client, stream proto.ChatRoom_JoinChatServer) error {
	//create stream to cleint
	  clientStream := &ClientStream{
        client: in,
        stream: stream,
    }
	

	s.clients = append(s.clients, clientStream)

	//send message to all clients "hello world"
	BroadCastMessage(in.Name + " has joined the chat", "server", time.Now().String(), s.clients)

	return nil
}

func (s *Server) SendMessage(ctx context.Context, in *proto.Chat) (*proto.Ack, error) {

	BroadCastMessage(in.Message, in.Name, in.Time, s.clients)

	ack:= &proto.Ack{
		Succes: true,
	}
	return ack, nil
}

func BroadCastMessage(message, name, time string, clients []*ClientStream) {
   for _, client := range clients {
        err := client.stream.Send(&proto.Chat{
            Time: time,
            Message: message,
			Name: name,
        })
        if err != nil {
            log.Fatalf("Could not send message to client %d", client.client.Port)
        }
    }
}

func (s *Server) LeaveChat(ctx context.Context, in *proto.Client) (*proto.Ack, error) {

	BroadCastMessage(in.Name + "has left the chat", "Server", time.Now().String(), s.clients)

	ack:= &proto.Ack{
		Succes: true,
	}
	return ack, nil
}

func setLog() *os.File {
	// Clears the log.txt file when a new server is started
	if err := os.Truncate("log.txt", 0); err != nil {
		log.Printf("Failed to truncate: %v", err)
	}

	// This connects to the log file/changes the output of the log informaiton to the log.txt file.
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
	return f
}