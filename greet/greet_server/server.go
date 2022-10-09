package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/minhtran241/grpc-go/greet/greetpb"
)

type server struct{}

func (*server) Greet(ctx context.Context, in *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	fmt.Printf("Greet function was invoked with %v\n", in)

	firstName := in.GetGreeting().GetFirstName()
	lastName := in.GetGreeting().GetLastName()

	result := "Hello " + firstName + " " + lastName

	res := &greetpb.GreetResponse{
		Result: result,
	}

	return res, nil
}

func (*server) GreetManyTimes(in *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	fmt.Printf("GreetManyTimes function was invoked with %v\n", in)

	firstName := in.GetGreeting().GetFirstName()
	lastName := in.GetGreeting().GetLastName()

	for i := 0; i < 10; i++ {
		result := "Hello " + firstName + " " + lastName + " number " + strconv.Itoa(i)
		res := &greetpb.GreetManyTimesResponse{
			Result: result,
		}
		stream.Send(res)
		time.Sleep(time.Second * 1)
	}

	return nil
}

func (*server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	fmt.Println("LongGreet function was invoked as a streaming request")

	result := "Hello "

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			// we've finished reading the client stream
			return stream.SendAndClose(&greetpb.LongGreetResponse{
				Result: result,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}

		firstName := req.GetGreeting().GetFirstName()
		lastName := req.GetGreeting().GetLastName()

		result += firstName + " " + lastName + "! "
	}
}

func (*server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	fmt.Println("GreetEveryone function was invoked as a streaming request")

	for {

		req, err := stream.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
			return err
		}

		firstName := req.GetGreeting().GetFirstName()
		lastName := req.GetGreeting().GetLastName()

		result := "Hello " + firstName + " " + lastName + "! "
		sendErr := stream.Send(&greetpb.GreetEveryoneResponse{
			Result: result,
		})

		if sendErr != nil {
			log.Fatalf("Error while sending data to client: %v", sendErr)
			return sendErr
		}
	}
}

func (*server) GreetWithDeadline(ctx context.Context, in *greetpb.GreetWithDeadlineRequest) (*greetpb.GreetWithDeadlineResponse, error) {
	fmt.Printf("GreetWithDeadline function was invoked with %v\n", in)

	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			fmt.Println("The client cancelled the request!")
			return nil, status.Error(codes.Canceled, "the client cancelled the request")
		}
		time.Sleep(time.Second * 1)
	}

	firstName := in.GetGreeting().GetFirstName()
	lastName := in.GetGreeting().GetLastName()

	result := "Hello " + firstName + " " + lastName

	res := &greetpb.GreetWithDeadlineResponse{
		Result: result,
	}

	return res, nil
}

func main() {
	fmt.Println("Server is running...")

	lis, err := net.Listen("tcp", "localhost:50051") // port binding

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	tls := true // use tls for security or not
	opts := []grpc.ServerOption{}

	if tls {
		certFile := "ssl/server.crt"
		keyFile := "ssl/server.pem"
		creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)
		if sslErr != nil {
			log.Fatalf("Failed loading certificates: %v", sslErr)
			return
		}
		opts = append(opts, grpc.Creds(creds))
	}

	s := grpc.NewServer(opts...)                     // grpc server
	greetpb.RegisterGreetServiceServer(s, &server{}) // register greet service

	// Register reflection service on gRPC server.
	reflection.Register(s)

	if err := s.Serve(lis); err != nil { // bind the port to the grpc server
		log.Fatalf("Failed to serve: %v", err)
	}
}
