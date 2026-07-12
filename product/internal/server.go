package internal

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/rasadov/EcommerceAPI/product/models"
	"github.com/rasadov/EcommerceAPI/product/proto/pb"
)

type grpcServer struct {
	pb.UnimplementedProductServiceServer
	service Service
}

func ListenGRPC(s Service, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	serv := grpc.NewServer()

	pb.RegisterProductServiceServer(serv, &grpcServer{
		UnimplementedProductServiceServer: pb.UnimplementedProductServiceServer{},
		service:                           s})
	reflection.Register(serv)
	return serv.Serve(lis)
}

func (s *grpcServer) GetProduct(ctx context.Context, r *wrapperspb.StringValue) (*pb.ProductResponse, error) {
	p, err := s.service.GetProduct(ctx, r.Value)
	if err != nil {
		return nil, err
	}
	return &pb.ProductResponse{Product: &pb.Product{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		AccountId:   int64(p.AccountID),
	}}, nil
}

func (s *grpcServer) GetProducts(ctx context.Context, r *pb.GetProductsRequest) (*pb.ProductsResponse, error) {
	var res []*models.Product
	var err error
	if r.Query != "" {
		res, err = s.service.SearchProducts(ctx, r.Query, r.Skip, r.Take)
	} else if len(r.Ids) != 0 {
		res, err = s.service.GetProductsWithIDs(ctx, r.Ids)
	} else {
		res, err = s.service.GetProducts(ctx, r.Skip, r.Take)
	}
	if err != nil {
		return nil, err
	}
	var products []*pb.Product
	for _, p := range res {
		products = append(products, &pb.Product{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			AccountId:   int64(p.AccountID),
		})

	}
	return &pb.ProductsResponse{Products: products}, nil
}

func (s *grpcServer) PostProduct(ctx context.Context, r *pb.CreateProductRequest) (*pb.ProductResponse, error) {
	p, err := s.service.PostProduct(ctx, r.GetName(), r.GetDescription(), r.Price, int(r.GetAccountId()))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &pb.ProductResponse{Product: &pb.Product{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		AccountId:   int64(p.AccountID),
	}}, nil
}

func (s *grpcServer) UpdateProduct(ctx context.Context, r *pb.UpdateProductRequest) (*pb.ProductResponse, error) {
	p, err := s.service.UpdateProduct(ctx, r.GetId(), r.GetName(), r.GetDescription(), r.Price, int(r.GetAccountId()))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &pb.ProductResponse{Product: &pb.Product{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		AccountId:   int64(p.AccountID),
	}}, nil
}

func (s *grpcServer) DeleteProduct(ctx context.Context, r *pb.DeleteProductRequest) (*emptypb.Empty, error) {
	err := s.service.DeleteProduct(ctx, r.GetProductId(), int(r.GetAccountId()))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
