package grpc

import (
	"fmt"
	common "github.com/getoptimum/optimum-common/pkg/rlnc"
	proto "github.com/getoptimum/optimum-common/pkg/rlnc/grpc/proto/v1"
	"math"
)

func encoderOptionsToProto(opts ...common.Option) (*proto.EncoderOptions, error) {
	var cfg common.EncoderConfig
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	if cfg.ShredFactor > math.MaxInt32 {
		return nil, fmt.Errorf("ShredFactor too large: %d", cfg.ShredFactor)
	}

	if cfg.NumShards > math.MaxInt32 {
		return nil, fmt.Errorf("NumShards too large: %d", cfg.NumShards)
	}

	if cfg.NumThreads > math.MaxInt32 {
		return nil, fmt.Errorf("NumThreads too large: %d", cfg.NumThreads)
	}

	if cfg.ShredMaxSize > math.MaxInt32 {
		return nil, fmt.Errorf("ShredMaxSize too large: %d", cfg.ShredMaxSize)
	}

	return &proto.EncoderOptions{
		ShredFactor:  int32Ptr(int32(cfg.ShredFactor)), // #nosec G115 safe after bounds check.
		NumShards:    int32Ptr(int32(cfg.NumShards)),   // #nosec G115 safe after bounds check.
		NumThreads:   int32Ptr(int32(cfg.NumThreads)),  // #nosec G115 safe after bounds check.
		Systematic:   boolPtr(cfg.Systematic),
		ShredMaxSize: int32Ptr(int32(cfg.ShredMaxSize)), // #nosec G115 safe after bounds check.
		PrepareData:  boolPtr(cfg.PrepareData),
	}, nil
}

func shardFromProto(in *proto.Shard) *common.Shard {
	if in == nil {
		return nil
	}
	return &common.Shard{
		Coefficients: append([]byte(nil), in.Coefficients...),
		Data:         append([]byte(nil), in.Data...),
	}
}

func shardToProto(in *common.Shard) *proto.Shard {
	if in == nil {
		return nil
	}
	return &proto.Shard{
		Coefficients: append([]byte(nil), in.Coefficients...),
		Data:         append([]byte(nil), in.Data...),
	}
}

func int32Ptr(v int32) *int32 { return &v }
func boolPtr(v bool) *bool    { return &v }
