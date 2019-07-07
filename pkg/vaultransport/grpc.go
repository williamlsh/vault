package vaultransport

import (
	"context"
	"errors"
	"time"

	"github.com/go-kit/kit/circuitbreaker"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/ratelimit"
	"github.com/go-kit/kit/transport"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	"github.com/sony/gobreaker"
	"github.com/williamzion/vault/pb"
	"github.com/williamzion/vault/pkg/vaultendpoint"
	"github.com/williamzion/vault/pkg/vaultservice"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
)

type grpcServer struct {
	hash     grpctransport.Handler
	validate grpctransport.Handler
}

// NewGRPCServer makes a set of endpoints available as a gRPC VaultServer.
func NewGRPCServer(endpoints vaultendpoint.Set, logger log.Logger) pb.VaultServer {
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
	}

	return &grpcServer{
		hash: grpctransport.NewServer(
			endpoints.HashEndpoint,
			decodeGRPCHashRequest,
			encodeGRPCHashResponse,
			options...,
		),
		validate: grpctransport.NewServer(
			endpoints.ValidateEndpoint,
			decodeGRPCValidateRequest,
			encodeGRPCValidateResponse,
			options...,
		),
	}
}

// NewGRPCClient returns a VaultService backed by  a gRPC server at the other end of the conn. The caller is responsible for constructuring the conn, and eventually closing the underlying transport.
func NewGRPCClient(conn *grpc.ClientConn, logger log.Logger) vaultservice.Service {
	limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 100))

	var hashEndpoint endpoint.Endpoint
	{
		hashEndpoint = grpctransport.NewClient(
			conn,
			"pb.Vault",
			"Hash",
			encodeGRPCHashRequest,
			decodeGRPCHashResponse,
			pb.HashResponse{},
		).Endpoint()
		hashEndpoint = limiter(hashEndpoint)
		hashEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Hash",
			Timeout: 30 * time.Second,
		}))(hashEndpoint)
	}
	var validateEndpoint endpoint.Endpoint
	{
		validateEndpoint = grpctransport.NewClient(
			conn,
			"pb.Vault",
			"Validate",
			encodeGRPCValidateRequest,
			decodeGRPCValidateResponse,
			pb.ValidateResponse{},
		).Endpoint()
		validateEndpoint = limiter(validateEndpoint)
		validateEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:    "Validate",
			Timeout: 10 * time.Second,
		}))(validateEndpoint)
	}

	return vaultendpoint.Set{
		HashEndpoint:     hashEndpoint,
		ValidateEndpoint: validateEndpoint,
	}
}

func (s *grpcServer) Hash(ctx context.Context, r *pb.HashRequest) (*pb.HashResponse, error) {
	_, resp, err := s.hash.ServeGRPC(ctx, r)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.HashResponse), nil
}

func (s *grpcServer) Validate(ctx context.Context, r *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	_, resp, err := s.validate.ServeGRPC(ctx, r)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.ValidateResponse), nil
}

// decodeGRPCHashRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC hash request to a user-domain hash request. Primarily useful in a
// server.
func decodeGRPCHashRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.HashRequest)
	return vaultendpoint.HashRequest{Password: req.Password}, nil
}

func decodeGRPCValidateRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.ValidateRequest)
	return vaultendpoint.ValidateRequest{Password: req.Password, Hash: req.Hash}, nil
}

// encodeGRPCHashResponse is a transport/grpc.EncodeResponseFunc that converts a user-domain validate response to a gRPC validate reply. Primarily useful in a server.
func encodeGRPCHashResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(vaultendpoint.HashResponse)
	return &pb.HashResponse{Hash: resp.Hash, Err: err2str(resp.Err)}, nil
}

func encodeGRPCValidateResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(vaultendpoint.ValidateResponse)
	return &pb.ValidateResponse{Valid: resp.Valid}, nil
}

func encodeGRPCHashRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(vaultendpoint.HashRequest)
	return &pb.HashRequest{Password: req.Password}, nil
}

func encodeGRPCValidateRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(vaultendpoint.ValidateRequest)
	return &pb.ValidateRequest{Password: req.Password, Hash: req.Hash}, nil
}

func decodeGRPCHashResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.HashResponse)
	return vaultendpoint.HashResponse{Hash: reply.Hash, Err: str2err(reply.Err)}, nil
}

func decodeGRPCValidateResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.ValidateResponse)
	return vaultendpoint.ValidateResponse{Valid: reply.Valid, Err: str2err("")}, nil
}

func err2str(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func str2err(s string) error {
	if s == "" {
		return nil
	}
	return errors.New(s)
}
