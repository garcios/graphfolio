package main

import (
	"log"
	"net"

	"portfolio-api/internal"
	pb "graphfolio/proto/portfolio/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	
	s := grpc.NewServer()
	pb.RegisterPortfolioServiceServer(s, internal.NewPortfolioServer())
	reflection.Register(s)
	
	log.Printf("portfolio-api listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
