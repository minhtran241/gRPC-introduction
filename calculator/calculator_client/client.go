package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/minhtran241/grpc-go/calculator/calculatorpb"
)

func main() {

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}

	defer cc.Close()

	c := calculatorpb.NewCalculatorServiceClient(cc)

	// doUnary(c)

	// doServerStreaming(c)

	// doClientStreaming(c)

	// doBiDiStreaming(c)

	doErrorUnary(c)
}

func doErrorUnary(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do SquareRoot Unary RPC...")

	//  correct call
	doErrorCall(c, 10)

	// error call
	doErrorCall(c, -10)
}

func doErrorCall(c calculatorpb.CalculatorServiceClient, number int32) {
	res, err := c.SquareRoot(context.Background(), &calculatorpb.SquareRootRequest{
		Number: number,
	})

	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			// actual error from gRPC (user error)
			fmt.Printf("(Error meaage from server) %s\n", respErr.Message())
			fmt.Println(respErr.Code())
			if respErr.Code() == codes.InvalidArgument {
				fmt.Println("We probably sent a negative number!")
				return
			}
		} else {
			log.Fatalf("Big Error calling SquareRoot: %v", err)
			return
		}
	}

	fmt.Printf("Result of SquareRoot of %v: %v\n", number, res.GetNumberRoot())
}

func doBiDiStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do BiDi Streaming RPC...")

	stream, err := c.FindMaximum(context.Background())

	if err != nil {
		log.Fatalf("Error while opening stream and calling FindMaximum: %v", err)
	}

	waitc := make(chan struct{})

	//  send go routine
	go func() {
		numbers := []int32{4, 7, 2, 19, 4, 6, 32}

		for i, number := range numbers {
			fmt.Printf("(%d)Sending number: %d...\n", i, number)
			stream.Send(&calculatorpb.FindMaximumRequest{
				Number: number,
			})
			time.Sleep(time.Second * 1)
		}
		stream.CloseSend()
	}()

	// receive go routine
	go func() {
		for {
			res, err := stream.Recv()

			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatalf("Problem while reading server stream response: %v", err)
				break
			}

			maximum := res.GetMaximum()
			log.Printf("Received a new maximum of...: %d\n", maximum)
		}
		close(waitc)
	}()

	// block until for loop broken
	<-waitc
}

func doClientStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do Client Streaming RPC...")

	stream, err := c.ComputeAverage(context.Background())

	if err != nil {
		log.Fatalf("Error while opening client stream: %v", err)
	}

	numbers := []int32{3, 5, 9, 54, 23}

	for _, number := range numbers {
		fmt.Printf("Sending number: %v\n", number)
		stream.Send(&calculatorpb.ComputeAverageRequest{
			Number: number,
		})
	}

	res, err := stream.CloseAndRecv()

	if err != nil {
		log.Fatalf("Error while receiving response: %v", err)
	}

	fmt.Printf("The Average is: %v\n", res.Average)
}

func doServerStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do Server Streaming RPC...")

	req := &calculatorpb.PrimeNumberDecompositionRequest{
		Number: 12,
	}

	streamRes, err := c.PrimeNumberDecomposition(context.Background(), req)

	if err != nil {
		log.Fatalf("error while calling GreetManyTimes RPC: %v", err)
	}

	results := []int32{}

	for {
		msg, err := streamRes.Recv()

		if err == io.EOF {
			// we've reached the end of the stream
			break
		}
		if err != nil {
			log.Fatalf("error while calling GreetManyTimes RPC: %v", err)
		}

		results = append(results, msg.GetPrimeNumberDecomposition())
	}

	fmt.Printf("Result: %v\n", results)

}

func doUnary(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do Unary RPC...")

	req := &calculatorpb.SumRequest{
		FirstNumber:  10,
		SecondNumber: 90,
	}

	res, err := c.Sum(context.Background(), req)

	if err != nil {
		log.Fatalf("error while calling Greet RPC: %v", err)
	}
	log.Printf("Response from Greet: %v", res.SumResult)
}
