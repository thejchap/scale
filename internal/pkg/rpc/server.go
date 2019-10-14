package rpc

import (
	"context"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/msmedes/scale/internal/pkg/keyspace"
	pb "github.com/msmedes/scale/internal/pkg/rpc/proto"
	"github.com/msmedes/scale/internal/pkg/scale"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// RPC rpc route handler
type RPC struct {
	node   scale.Node
	sugar  *zap.SugaredLogger
	logger *zap.Logger
	addr   string
}

// NewRPC create a new rpc
func NewRPC(node scale.Node, logger *zap.Logger, sugar *zap.SugaredLogger, addr string) *RPC {
	return &RPC{
		logger: logger,
		sugar:  sugar,
		node:   node,
		addr:   addr,
	}
}

// TransferKeys proxy to node.TransferKeys
func (r *RPC) TransferKeys(ctx context.Context, in *pb.KeyTransferRequest) (*pb.Success, error) {
	r.node.TransferKeys(keyspace.ByteArrayToKey(in.GetId()), in.GetAddr())
	return &pb.Success{}, nil
}

// Ping health check
func (r *RPC) Ping(ctx context.Context, in *pb.Empty) (*pb.Success, error) {
	return &pb.Success{}, nil
}

// Get rpc wrapper for node.Get
func (r *RPC) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	val, err := r.node.Get(keyspace.ByteArrayToKey(in.GetKey()))

	if err != nil {
		return nil, err
	}

	res := &pb.GetResponse{Value: val}

	return res, nil
}

// Set rpc wrapper for node.Set
func (r *RPC) Set(ctx context.Context, in *pb.SetRequest) (*pb.Success, error) {
	r.node.Set(keyspace.ByteArrayToKey(in.GetKey()), in.GetValue())

	return &pb.Success{}, nil
}

// GetLocal rpc wrapper for node.store.Get
func (r *RPC) GetLocal(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	val, err := r.node.GetLocal(keyspace.ByteArrayToKey(in.GetKey()))

	if err != nil {
		return nil, err
	}

	res := &pb.GetResponse{Value: val}

	return res, nil
}

// SetLocal rpc wrapper for node.store.Set
func (r *RPC) SetLocal(ctx context.Context, in *pb.SetRequest) (*pb.Success, error) {
	r.node.SetLocal(keyspace.ByteArrayToKey(in.GetKey()), in.GetValue())

	return &pb.Success{}, nil
}

// FindSuccessor rpc wrapper for node.FindSuccessor
func (r *RPC) FindSuccessor(ctx context.Context, in *pb.RemoteQuery) (*pb.RemoteNode, error) {
	successor, _ := r.node.FindSuccessor(keyspace.ByteArrayToKey(in.Id))
	id := successor.GetID()

	res := &pb.RemoteNode{
		Id:   id[:],
		Addr: successor.GetAddr(),
	}

	return res, nil
}

// GetSuccessor successor of the node
func (r *RPC) GetSuccessor(context.Context, *pb.Empty) (*pb.RemoteNode, error) {
	successor, err := r.node.GetSuccessor()

	if err != nil {
		return nil, err
	}

	id := successor.GetID()

	res := &pb.RemoteNode{
		Id:   id[:],
		Addr: successor.GetAddr(),
	}

	return res, nil
}

// GetPredecessor returns the predecessor of the node
func (r *RPC) GetPredecessor(context.Context, *pb.Empty) (*pb.RemoteNode, error) {
	predecessor, err := r.node.GetPredecessor()

	if err != nil {
		return nil, err
	} else if predecessor == nil {
		empty := &pb.RemoteNode{Present: false}

		return empty, nil
	}

	id := predecessor.GetID()

	res := &pb.RemoteNode{
		Id:      id[:],
		Addr:    predecessor.GetAddr(),
		Present: true,
	}

	return res, nil
}

// Notify tells a node that another node (it thinks) it's its predecessor
// man english is a weird language
func (r *RPC) Notify(ctx context.Context, in *pb.RemoteNode) (*pb.Success, error) {
	err := r.node.Notify(keyspace.ByteArrayToKey(in.Id), in.Addr)

	if err != nil {
		return nil, err
	}

	return &pb.Success{}, nil
}

// GetNodeMetadata return metadata about this node
func (r *RPC) GetNodeMetadata(context.Context, *pb.Empty) (*pb.NodeMetadata, error) {
	id := r.node.GetID()

	var ft [][]byte

	for _, k := range r.node.GetFingerTableIDs() {
		ft = append(ft, k[:])
	}

	meta := &pb.NodeMetadata{
		Id:          id[:],
		Addr:        r.node.GetAddr(),
		FingerTable: ft,
	}

	predecessor, err := r.node.GetPredecessor()

	if err != nil {
		return nil, err
	}

	if predecessor != nil {
		predID := predecessor.GetID()
		meta.PredecessorId = predID[:]
		meta.PredecessorAddr = predecessor.GetAddr()
	}

	successor, err := r.node.GetSuccessor()

	if err != nil {
		return nil, err
	}

	if successor != nil {
		succID := successor.GetID()
		meta.SuccessorId = succID[:]
		meta.SuccessorAddr = successor.GetAddr()
	}

	return meta, nil
}

// ServerListen start up the server
func (r *RPC) ServerListen() {
	server, err := net.Listen("tcp", r.addr)

	if err != nil {
		r.sugar.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_zap.UnaryServerInterceptor(r.logger),
		)),
	)

	pb.RegisterScaleServer(grpcServer, r)
	grpcServer.Serve(server)
}