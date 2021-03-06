package scale

import "context"

// Canonical types for the implementation

// M bit keyspace
const M int = 32

// Key 20 byte key
type Key = [M / 8]byte

type baseNode interface {
	ClosestPrecedingFinger(context.Context, Key) (RemoteNode, error)
	FindPredecessor(context.Context, Key) (RemoteNode, error)
	FindSuccessor(context.Context, Key) (RemoteNode, error)
	GetAddr() string
	GetID() Key
	GetLocal(context.Context, Key) ([]byte, error)
	GetPredecessor(context.Context) (RemoteNode, error)
	GetSuccessor(context.Context) (RemoteNode, error)
	SetLocal(context.Context, Key, []byte) error
	SetPredecessor(string, string) error
	SetSuccessor(string, string) error
}

// RemoteNode contains metadata (ID and Address) about another node in the network
type RemoteNode interface {
	baseNode

	CloseConnection() error
	GetNetwork([]string) ([]string, error)
	Notify(Node) error
	Ping() error
}

// Node represents the current node and operations it is responsible for
type Node interface {
	baseNode

	Get(context.Context, Key) ([]byte, error)
	GetFingerTableIDs() []Key
	GetFingerTableAddrs() []string
	GetPort() string
	GetKeys() []string
	Notify(Key, string) error
	SendTraceID(context.Context, string) context.Context
	SendTraceIdRPC(context.Context, string) context.Context
	Set(context.Context, Key, []byte) error
	TransferKeys(Key, string) int
}

// Store represents a Scale-compatible underlying data store
type Store interface {
	Del(Key) error
	Get(Key) []byte
	Keys() []Key
	Set(Key, []byte) error
}
