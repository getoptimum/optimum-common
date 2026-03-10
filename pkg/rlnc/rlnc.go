package rlnc

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

func WithSystematic(systematic bool) Option {
	return func(cfg *EncoderConfig) {
		cfg.Systematic = systematic
	}
}

func WithNumThreads(numThreads int) Option {
	return func(cfg *EncoderConfig) {
		cfg.NumThreads = numThreads
	}
}

func WithNumShards(numShards int) Option {
	return func(cfg *EncoderConfig) {
		cfg.NumShards = numShards
	}
}

func WithShredFactor(shredFactor int) Option {
	return func(cfg *EncoderConfig) {
		cfg.ShredFactor = shredFactor
	}
}

func WithShredMaxSize(maxSize int) Option {
	return func(cfg *EncoderConfig) {
		cfg.ShredMaxSize = maxSize
	}
}

func WithPrepareData(prepareData bool) Option {
	return func(cfg *EncoderConfig) {
		cfg.PrepareData = prepareData
	}
}
