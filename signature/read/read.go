package main

import (
	"flag"
	"log"
	"net"

	"github.com/cjhouser/washere/models"
	"google.golang.org/grpc"
)

type signatureServer struct {
	models.UnimplementedSignatureServer
}

func (s *signatureServer) Get(request *models.SignatureRequest, stream models.Signature_GetServer) error {
	err := stream.Send(&models.SignatureResponse{Id: 0, Signature: "bruh"})
	if err != nil {
		return err
	}
	err = stream.Send(&models.SignatureResponse{Id: 1, Signature: "dude"})
	if err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", "localhost:8081")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	models.RegisterSignatureServer(grpcServer, &signatureServer{})
	grpcServer.Serve(lis)
}
