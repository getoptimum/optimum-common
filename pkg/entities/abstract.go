package entities

type DCConfigurable[T any] interface {
	ApplyDynamicConfig(dcCfg *DynamicConfig) *T
	Clone() *T
}
