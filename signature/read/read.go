package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"

	"github.com/cjhouser/washere/models"
	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc"
)

type signatureServer struct {
	models.UnimplementedSignatureServer
	databaseConnection *pgx.Conn
	context            context.Context
}

func (s *signatureServer) Get(request *models.SignatureRequest, stream models.Signature_GetServer) error {
	resp, err := s.databaseConnection.Query(s.context, "SELECT * FROM signatures;")
	if err != nil {
		log.Println("failed to select from database", err)
	}

	for resp.Next() {
		var id uint64
		var signature string
		err = resp.Scan(&id, &signature)
		if err != nil {
			log.Println("scan failure", err)
		}
		err = stream.Send(&models.SignatureResponse{Id: id, Signature: signature})
		if err != nil {
			log.Println("failed sending data to frontend", err)
		}
	}
	return nil
}

func main() {
	flag.Parse()
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalln("unable to connect", err)
	}
	defer conn.Close(context.Background())

	log.Println("listening on the meme thing")
	lis, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	models.RegisterSignatureServer(grpcServer, &signatureServer{databaseConnection: conn, context: context.Background()})
	grpcServer.Serve(lis)
}
