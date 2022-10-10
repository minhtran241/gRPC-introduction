package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/minhtran241/grpc-go/blog/blogpb"
)

func main() {
	fmt.Println("Client is running...")

	tls := false // use tls for security or not
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

	c := blogpb.NewBlogServiceClient(cc)

	// create Blog
	fmt.Println("Creating the blog...")
	blog := &blogpb.Blog{
		AuthorId: "Balamurugan Balusamy",
		Title:    "Big Data",
		Content:  "An introduction to Big Data",
	}
	createBlogRes, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{Blog: blog})
	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}
	fmt.Printf("Blog has been created: %v\n", createBlogRes)
}
