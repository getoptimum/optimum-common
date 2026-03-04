package rlnc

import (
	"context"
)

type Shard struct {
	Coefficients []byte
	Data         []byte
}

type Option func(*EncoderConfig)

type EncoderConfig struct {
	ShredFactor  int  // Number of original splits
	NumShards    int  // Total shards to return
	NumThreads   int  // Number of goroutines to spawn
	Systematic   bool // If true, emit systematic shards first
	ShredMaxSize int  // Maximum size of each shard in bytes
	PrepareData  bool // If true, prepare data for coding
}

type EncodeIntoShards = func(context.Context, []byte, ...Option) ([]*Shard, error)
type StreamShards = func() (chan *Shard, error)

type ShardSet interface {
	Add(coefficients, data []byte) error
	TryDecode() ([]byte, bool)
	Recode() (*Shard, error)
	IsEmpty() bool
	Size() int
	NumCoefficients() int
	ShardLength() int
	IsUncoded() bool
}
