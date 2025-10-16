package entities

type DCConfigurable interface {
	ApplyDynamicConfig(dcCfg *DynamicConfig)
}
