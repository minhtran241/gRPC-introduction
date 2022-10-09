package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"github.com/minhtran241/grpc-go/greet/greetpb"
)

func main() {
	fmt.Println("Client is running...")

	tls := true // use tls for security or not
	opts := grpc.WithInsecure()

	if tls {
		certFile := "ssl/ca.crt" // Certificate Authority Trust certificate
		creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")

		if sslErr != nil {
			log.Fatalf("Error while loading CA trust certificate: %v", sslErr)
			return
		}

		opts = grpc.WithTransportCredentials(creds)
	}

	cc, err := grpc.Dial("localhost:50051", opts)

	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}

	defer cc.Close()

	c := greetpb.NewGreetServiceClient(cc)

	doUnary(c)

	// doServerStreaming(c)

	// doClientStreaming(c)

	// doBiDiStreaming(c)

	// doUnaryWithDeadline(c, 5*time.Second) // should complete
	// doUnaryWithDeadline(c, 1*time.Second) // should timeout
}

func doUnaryWithDeadline(c greetpb.GreetServiceClient, timeout time.Duration) {
	fmt.Println("Starting to do UnaryWithDeadline RPC...")

	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Minh",
			LastName:  "Tran",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := c.GreetWithDeadline(ctx, req)

	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("Timeout was hit! Deadline was exceeded")
			} else {
				fmt.Printf("Unexpected error: %v\n", statusErr)
			}
		}
		log.Fatalf("(error while calling GreetWithDeadline RPC) %v\n", err)
		return
	}
	log.Printf("Response from GreetWithDeadline: %v\n", res)
}

func doBiDiStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do BiDi Streaming RPC...")

	// we create a stream by invoking the client
	stream, err := c.GreetEveryone(context.Background())

	if err != nil {
		log.Fatalf("Error while creating stream: %v", err)
		return
	}

	requests := []*greetpb.GreetEveryoneRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Minh",
				LastName:  "Tran",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Minh",
				LastName:  "Tran",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Minh",
				LastName:  "Tran",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Minh",
				LastName:  "Tran",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Minh",
				LastName:  "Tran",
			},
		},
	}

	waitc := make(chan struct{})

	// we send a bunch of messages to the server (go routine)
	go func() {
		// function to send a bunch of messages to the server
		for i, req := range requests {
			fmt.Printf("Sending message number %d: %v\n", i, req)
			err := stream.Send(req)
			if err != nil {
				log.Fatalf("Error sending request number: %d", i)
			}

			time.Sleep(time.Second * 1)
		}
		stream.CloseSend()
	}()

	// we receive a bunch of messages from the server  (go routine)
	go func() {
		// function to receive a bunch of messages to the server

		for {
			res, err := stream.Recv()

			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatalf("Error while receiving: %v", err)
				break
			}

			fmt.Printf("Received: %v\n", res.GetResult())
		}
		close(waitc)
	}()

	// block until everything is done
	<-waitc
}

func doClientStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do Client Streaming RPC...")

	requests := []*greetpb.LongGreetRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Minh",
				LastName:  "Tran",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Minh",
				LastName:  "Tran",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Minh",
				LastName:  "Tran",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Minh",
				LastName:  "Tran",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Minh",
				LastName:  "Tran",
			},
		},
	}

	stream, err := c.LongGreet(context.Background())

	if err != nil {
		log.Fatalf("error while calling LongGreet: %v", err)
	}

	// we iterate over out slice and send each message individually
	for _, req := range requests {
		fmt.Printf("Sending req: %v\n", req)
		stream.Send(req)
		time.Sleep(time.Second * 1)
	}

	res, err := stream.CloseAndRecv()

	if err != nil {
		log.Fatalf("error while receiving response from LongGreet: %v", err)
	}

	fmt.Printf("LongGreet Response: %v\n", res)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do Server Streaming RPC...")

	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Minh",
			LastName:  "Tran",
		},
	}

	streamRes, err := c.GreetManyTimes(context.Background(), req)

	if err != nil {
		log.Fatalf("error while calling GreetManyTimes RPC: %v", err)
	}

	for {
		msg, err := streamRes.Recv()

		if err == io.EOF {
			// we've reached the end of the stream
			break
		}
		if err != nil {
			log.Fatalf("error while calling GreetManyTimes RPC: %v", err)
		}
		log.Printf("Response from GreetManyTimes: %v", msg.GetResult())
	}

}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do Unary RPC...")

	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Minh",
			LastName:  "Tran",
		},
	}

	res, err := c.Greet(context.Background(), req)

	if err != nil {
		log.Fatalf("error while calling Greet RPC: %v", err)
	}
	log.Printf("Response from Greet: %v", res.Result)
}
