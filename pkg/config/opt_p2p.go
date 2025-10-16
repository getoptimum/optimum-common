package config

type OptimumConfig struct {
	ClusterID      string `yaml:"cluster_id" env:"CLUSTER_ID" flag:"cluster_id"`
	ListenPort     int    `yaml:"listen_port" env:"OPTIMUM_PORT" flag:"listen_port"`
	MaxMessageSize int64  `yaml:"max_message_size_bytes" env:"OPTIMUM_MAX_MSG_SIZE" flag:"max_message_size_bytes"`

	// RLNC and message settings
	RandomMessageSize        int64   `yaml:"random_message_size_bytes" env:"OPTIMUM_RANDOM_MSG_SIZE" flag:"random_message_size_bytes"`
	ShardFactor              int     `yaml:"rlnc_shard_factor" env:"OPTIMUM_SHARD_FACTOR" flag:"rlnc_shard_factor"`
	PublisherShardMultiplier float32 `yaml:"publisher_shard_multiplier" env:"OPTIMUM_SHARD_MULT" flag:"publisher_shard_multiplier"`
	ForwardShardThreshold    float32 `yaml:"forward_shard_threshold" env:"OPTIMUM_THRESHOLD" flag:"forward_shard_threshold"`

	// Mesh topology settings
	MeshDegreeTarget int `yaml:"mesh_degree_target" env:"OPTIMUM_MESH_TARGET" flag:"mesh_degree_target"`
	MeshDegreeMin    int `yaml:"mesh_degree_min" env:"OPTIMUM_MESH_MIN" flag:"mesh_degree_min"`
	MeshDegreeMax    int `yaml:"mesh_degree_max" env:"OPTIMUM_MESH_MAX" flag:"mesh_degree_max"`

	BootstrapPeers []string `yaml:"bootstrap_peers" env:"BOOTSTRAP_PEERS" flag:"bootstrap_peers"`
}
