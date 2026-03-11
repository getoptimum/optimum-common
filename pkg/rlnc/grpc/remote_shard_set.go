package grpc

import (
	"context"
	common "github.com/getoptimum/optimum-common/pkg/rlnc"
	proto "github.com/getoptimum/optimum-common/pkg/rlnc/grpc/proto/v1"
	"log"
)

type RemoteShardSet struct {
	id     uint64
	client proto.ShardSetServiceClient
}

func (r *RemoteShardSet) ID() uint64 {
	if r == nil {
		log.Println("remote shard set is nil: returning ID 0")
		return 0
	}
	return r.id
}

func (r *RemoteShardSet) Add(ctx context.Context, coefficients, data []byte) error {
	_, err := r.client.Add(ctx, &proto.AddShardRequest{
		ShardSetId:   r.id,
		Coefficients: coefficients,
		Data:         data,
	})
	return err
}

func (r *RemoteShardSet) TryDecode(ctx context.Context) (data []byte, ok bool, err error) {
	resp, err := r.client.TryDecode(ctx, &proto.TryDecodeRequest{
		ShardSetId: r.id,
	})
	if err != nil {
		return nil, false, err
	}
	data = append([]byte(nil), resp.DecodedData...)
	ok = resp.Ok
	return
}

func (r *RemoteShardSet) Recode(ctx context.Context) (*common.Shard, error) {
	resp, err := r.client.Recode(ctx, &proto.RecodeRequest{
		ShardSetId: r.id,
	})
	if err != nil {
		return nil, err
	}
	return shardFromProto(resp.Shard), nil
}

func (r *RemoteShardSet) IsEmpty(ctx context.Context) (bool, error) {
	resp, err := r.client.IsEmpty(ctx, &proto.IsEmptyRequest{
		ShardSetId: r.id,
	})
	if err != nil {
		return false, err
	}
	return resp.IsEmpty, nil
}

func (r *RemoteShardSet) Size(ctx context.Context) (int, error) {
	resp, err := r.client.Size(ctx, &proto.SizeRequest{
		ShardSetId: r.id,
	})
	if err != nil {
		return 0, err
	}
	return int(resp.Size), nil
}

func (r *RemoteShardSet) NumCoefficients(ctx context.Context) (int, error) {
	resp, err := r.client.NumCoefficients(ctx, &proto.NumCoefficientsRequest{
		ShardSetId: r.id,
	})
	if err != nil {
		return 0, err
	}
	return int(resp.NumCoefficients), nil
}

func (r *RemoteShardSet) ShardLength(ctx context.Context) (int, error) {
	resp, err := r.client.ShardLength(ctx, &proto.ShardLengthRequest{
		ShardSetId: r.id,
	})
	if err != nil {
		return 0, err
	}
	return int(resp.ShardLength), nil
}

func (r *RemoteShardSet) Close(ctx context.Context) error {
	_, err := r.client.CloseShardSet(ctx, &proto.CloseShardSetRequest{
		ShardSetId: r.id,
	})
	return err
}
