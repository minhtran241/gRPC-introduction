package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	// "time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/minhtran241/grpc-go/calculator/calculatorpb"
)

type server struct{}

func (*server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	fmt.Printf("Received Sum RPC: %v\n", req)

	firstNumber := req.FirstNumber
	secondNumber := req.SecondNumber

	sum := firstNumber + secondNumber

	res := &calculatorpb.SumResponse{
		SumResult: sum,
	}

	return res, nil
}

func (*server) PrimeNumberDecomposition(in *calculatorpb.PrimeNumberDecompositionRequest, stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer) error {

	fmt.Printf("PrimeNumberDecomposition function was invoked with %v\n", in)

	number := in.GetNumber()

	primeDecomposition := getPrimeNumberDecomposition(number)

	for _, value := range primeDecomposition {
		res := &calculatorpb.PrimeNumberDecompositionResponse{
			PrimeNumberDecomposition: value,
		}

		stream.Send(res)
		// time.Sleep(time.Second * 1)
	}

	return nil
}

func getPrimeNumberDecomposition(number int32) (primeDecomposition []int32) {

	k := int32(2)
	res := []int32{}

	for number > 1 {
		if number%k == 0 {
			res = append(res, k)
			number = number / k
		} else {
			k++
		}
	}

	return res
}

func (*server) ComputeAverage(stream calculatorpb.CalculatorService_ComputeAverageServer) error {
	fmt.Println("Received ComputeAverage RPC")

	sum := int32(0)
	count := 0

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			average := float64(sum) / float64(count)
			return stream.SendAndClose(&calculatorpb.ComputeAverageResponse{
				Average: average,
			})
		}

		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}

		sum += req.GetNumber()
		count++
	}

}

func (*server) FindMaximum(stream calculatorpb.CalculatorService_FindMaximumServer) error {
	fmt.Println("Received FindMaximum RPC")

	maximum := int32(math.MinInt32)

	for {
		req, err := stream.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
			return err
		}
		number := req.GetNumber()

		if number > maximum {
			maximum = number
			sendErr := stream.Send(&calculatorpb.FindMaximumResponse{
				Maximum: maximum,
			})
			if sendErr != nil {
				log.Fatalf("Error while sending data to client: %v", sendErr)
			}
		}
	}
}

func (*server) SquareRoot(ctx context.Context, in *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	fmt.Println("Received SquareRoot RPC")

	number := in.GetNumber()

	if number < 0 {
		//  send back an error
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Received a negative number: %v", number),
		)
	}

	return &calculatorpb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	}, nil
}

func main() {
	fmt.Println("Server is running...")

	lis, err := net.Listen("tcp", "0.0.0.0:50051") // port binding

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()                                      // grpc server
	calculatorpb.RegisterCalculatorServiceServer(s, &server{}) // register greet service

	// Register reflection service on gRPC server.
	reflection.Register(s)

	if err := s.Serve(lis); err != nil { // bind the port to the grpc server
		log.Fatalf("Failed to serve: %v", err)
	}
}
