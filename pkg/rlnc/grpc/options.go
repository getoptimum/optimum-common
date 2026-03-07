package grpc

import (
	common "github.com/getoptimum/optimum-common/pkg/rlnc"
	proto "github.com/getoptimum/optimum-common/pkg/rlnc/grpc/proto/v1"
)

func encoderOptionsToProto(opts ...common.Option) *proto.EncoderOptions {
	var cfg common.EncoderConfig
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	return &proto.EncoderOptions{
		ShredFactor:  int32Ptr(int32(cfg.ShredFactor)),
		NumShards:    int32Ptr(int32(cfg.NumShards)),
		NumThreads:   int32Ptr(int32(cfg.NumThreads)),
		Systematic:   boolPtr(cfg.Systematic),
		ShredMaxSize: int32Ptr(int32(cfg.ShredMaxSize)),
		PrepareData:  boolPtr(cfg.PrepareData),
	}
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
