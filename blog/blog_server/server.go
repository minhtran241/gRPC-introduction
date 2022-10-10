package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/minhtran241/grpc-go/blog/blogpb"
)

var collection *mongo.Collection

type server struct{}

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:"content"`
	Title    string             `bson:"title"`
}

func (*server) CreateBlog(ctx context.Context, in *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	fmt.Println("Create blog request")
	blog := in.GetBlog()
	data := blogItem{
		AuthorID: blog.GetAuthorId(),
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
	}

	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintln("Cannot convert to OID"),
		)
	}

	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title:    blog.GetTitle(),
			Content:  blog.GetContent(),
		},
	}, nil
}

func dataToBlogPb(data *blogItem) *blogpb.Blog {
	return &blogpb.Blog{
		Id:       data.ID.Hex(),
		AuthorId: data.AuthorID,
		Title:    data.Title,
		Content:  data.Content,
	}
}

func (*server) ReadBlog(ctx context.Context, in *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	fmt.Println("Read blog request")
	blogID := in.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintln("Can not parse ID"),
		)
	}
	// create an empty struct
	data := &blogItem{}
	filter := bson.D{{"_id", oid}}
	err = collection.FindOne(context.Background(), filter).Decode(data)
	if err == mongo.ErrNoDocuments {
		// Do something when no record was found
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Can not find blog with specified ID: %v\n", err))
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Something went wrong: %v\n", err))
	}
	return &blogpb.ReadBlogResponse{
		Blog: dataToBlogPb(data),
	}, nil
}

func (*server) UpdateBlog(ctx context.Context, in *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	fmt.Println("Update blog request")
	blog := in.GetBlog()
	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintln("Can not parse ID"),
		)
	}
	data := &blogItem{}
	filter := bson.D{{"_id", oid}}
	err = collection.FindOne(context.Background(), filter).Decode(data)
	if err == mongo.ErrNoDocuments {
		// Do something when no record was found
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Can not find blog with specified ID: %v\n", err))
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Something went wrong: %v\n", err))
	}
	// we update our internal struct
	data.AuthorID = blog.GetAuthorId()
	data.Content = blog.GetContent()
	data.Title = blog.GetTitle()

	_, updateErr := collection.ReplaceOne(context.Background(), filter, data)
	if updateErr != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Can not update object in MongoDB: %v\n", updateErr))
	}
	return &blogpb.UpdateBlogResponse{
		Blog: dataToBlogPb(data),
	}, nil
}

func (*server) DeleteBlog(ctx context.Context, in *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	fmt.Println("Delete blog request")
	blogID := in.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintln("Can not parse ID"),
		)
	}
	filter := bson.D{{"_id", oid}}
	deleteRes, deleteErr := collection.DeleteOne(context.Background(), filter)
	if deleteRes.DeletedCount == 0 {
		// Do something when no record was found
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Can not find blog with specified ID: %v\n", deleteErr))
	} else if deleteErr != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Something went wrong: %v\n", deleteErr))
	}
	return &blogpb.DeleteBlogResponse{
		BlogId: blogID,
	}, nil
}

func (*server) ListBlog(in *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	fmt.Println("List blog request")
	// cursor
	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Internal error: %v\n", err))
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		data := &blogItem{}
		err := cur.Decode(data)
		if err != nil {
			return status.Errorf(codes.Internal, fmt.Sprintf("Error while decoding data from MongoDB: %v\n", err))
		}
		stream.Send(&blogpb.ListBlogResponse{
			Blog: dataToBlogPb(data),
		})
	}
	if err := cur.Err(); err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Internal error: %v\n", err))
	}
	return nil
}

func main() {
	// if we crash the go code, we get the file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Connecting to MongoDB...")
	// connect to MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		fmt.Println("Closing MongoDB Connection...")
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	fmt.Println("Blog Server is running...")
	collection = client.Database("mydb").Collection("blog")

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

	s := grpc.NewServer(opts...)                   // grpc server
	blogpb.RegisterBlogServiceServer(s, &server{}) // register greet service

	// Register reflection service on gRPC server.
	reflection.Register(s)

	go func() {
		fmt.Println("Starting Server...")
		if err := s.Serve(lis); err != nil { // bind the port to the grpc server
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for Control C to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// Block until a signal is received
	<-ch
	fmt.Println("Stopping the server...")
	s.Stop()
	fmt.Println("Stopping the listener...")
	lis.Close()
	fmt.Println("End of Program")
}
