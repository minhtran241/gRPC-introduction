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
	blogId := createBlogRes.GetBlog().GetId()

	// read Blog
	fmt.Println("Reading the blog...")
	_, err2 := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{
		BlogId: "some wrong id",
	})
	if err2 != nil {
		fmt.Printf("Error happened while reading: %v\n", err2)
	}
	readBlogRes, readBlogErr := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{
		BlogId: blogId,
	})
	if readBlogErr != nil {
		fmt.Printf("Error happened while reading: %v\n", readBlogErr)
	}
	fmt.Printf("Blog was read: %v\n", readBlogRes)

	// update Blog
	newBlog := &blogpb.Blog{
		Id:       blogId,
		AuthorId: "Changed Author",
		Title:    "Big Data (edited)",
		Content:  "An introduction to Big Data (edited)",
	}
	updateRes, updateErr := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{
		Blog: newBlog,
	})
	if updateErr != nil {
		fmt.Printf("Error happened while updating: %v\n", updateErr)
	}
	fmt.Printf("Blog was updated: %v\n", updateRes)

	// delete Blog
	deleteRes, deleteErr := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{
		BlogId: blogId,
	})
	if deleteErr != nil {
		fmt.Printf("Error happened while deleting: %v\n", deleteErr)
	}
	fmt.Printf("Blog was deleted: %v\n", deleteRes)
}
