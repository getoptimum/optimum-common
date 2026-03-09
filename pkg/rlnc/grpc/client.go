package grpc

import (
	"context"
	"errors"
	"fmt"
	proto "github.com/getoptimum/optimum-common/pkg/rlnc/grpc/proto/v1"
	"io"
	"math"
	"net"

	common "github.com/getoptimum/optimum-common/pkg/rlnc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn      *grpc.ClientConn
	rlnc      proto.RlncServiceClient
	shardSets proto.ShardSetServiceClient
}

type StreamResult struct {
	Shard *common.Shard
	Err   error
}

func Dial(ctx context.Context, socketPath string, opts ...grpc.DialOption) (*Client, error) {
	baseOpts := make([]grpc.DialOption, 0, 2+len(opts))
	baseOpts = append(baseOpts,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "unix", socketPath)
		}),
	)
	baseOpts = append(baseOpts, opts...)

	conn, err := grpc.NewClient("unix://"+socketPath, baseOpts...)
	if err != nil {
		return nil, fmt.Errorf("dial grpc unix socket %q: %w", socketPath, err)
	}

	return &Client{
		conn:      conn,
		rlnc:      proto.NewRlncServiceClient(conn),
		shardSets: proto.NewShardSetServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) EncodeIntoShards(ctx context.Context, inputData []byte, opts ...common.Option) ([]*common.Shard, error) {
	options, err := encoderOptionsToProto(opts...)
	if err != nil {
		return nil, err
	}
	resp, err := c.rlnc.EncodeIntoShards(ctx, &proto.EncodeIntoShardsRequest{
		InputData: inputData,
		Options:   options,
	})
	if err != nil {
		return nil, err
	}

	out := make([]*common.Shard, 0, len(resp.Shards))
	for _, shard := range resp.Shards {
		out = append(out, shardFromProto(shard))
	}
	return out, nil
}

func (c *Client) StreamShards(ctx context.Context, inputData []byte, opts ...common.Option) (<-chan StreamResult, error) {
	options, err := encoderOptionsToProto(opts...)
	if err != nil {
		return nil, err
	}
	stream, err := c.rlnc.StreamShards(ctx, &proto.StreamShardsRequest{
		InputData: inputData,
		Options:   options,
	})
	if err != nil {
		return nil, err
	}

	out := make(chan StreamResult)

	go func() {
		defer close(out)

		for {
			msg, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				out <- StreamResult{Err: err}
				return
			}

			select {
			case <-ctx.Done():
				out <- StreamResult{Err: ctx.Err()}
				return
			case out <- StreamResult{Shard: shardFromProto(msg)}:
			}
		}
	}()

	return out, nil
}

func (c *Client) IsUncoded(ctx context.Context, shard *common.Shard) (bool, error) {
	resp, err := c.rlnc.IsUncoded(ctx, &proto.IsUncodedRequest{
		Shard: shardToProto(shard),
	})
	if err != nil {
		return false, err
	}
	return resp.IsUncoded, nil
}

func (c *Client) GetEncoderConfig(ctx context.Context, dataLen int, opts ...common.Option) (common.EncoderConfig, error) {
	if dataLen > math.MaxInt32 {
		return common.EncoderConfig{}, fmt.Errorf("dataLen too large: %d", dataLen)
	}

	options, err := encoderOptionsToProto(opts...)
	if err != nil {
		return common.EncoderConfig{}, err
	}
	resp, err := c.rlnc.GetEncoderConfig(ctx, &proto.GetEncoderConfigRequest{
		DataLen: int32(dataLen), // #nosec G115 safe after bounds check.
		Options: options,
	})
	if err != nil {
		return common.EncoderConfig{}, err
	}

	cfg := resp.GetConfig()
	if cfg == nil {
		return common.EncoderConfig{}, nil
	}

	return common.EncoderConfig{
		ShredFactor:  int(cfg.ShredFactor),
		NumShards:    int(cfg.NumShards),
		NumThreads:   int(cfg.NumThreads),
		Systematic:   cfg.Systematic,
		ShredMaxSize: int(cfg.ShredMaxSize),
		PrepareData:  cfg.PrepareData,
	}, nil
}

func (c *Client) NewShardSet(ctx context.Context, numCoefficients, shardLength int) (*RemoteShardSet, error) {
	if numCoefficients > math.MaxInt32 {
		return nil, fmt.Errorf("numCoefficients too large: %d", numCoefficients)
	}

	if shardLength > math.MaxInt32 {
		return nil, fmt.Errorf("shardLength too large: %d", shardLength)
	}

	resp, err := c.shardSets.NewShardSet(ctx, &proto.NewShardSetRequest{
		NumCoefficients: int32(numCoefficients), // #nosec G115 safe after bounds check
		ShardLength:     int32(shardLength),     // #nosec G115 safe after bounds check
	})
	if err != nil {
		return nil, err
	}

	return &RemoteShardSet{
		id:     resp.ShardSetId,
		client: c.shardSets,
	}, nil
}
