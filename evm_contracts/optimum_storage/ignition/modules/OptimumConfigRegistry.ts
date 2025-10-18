import { buildModule } from "@nomicfoundation/hardhat-ignition/modules"

export default buildModule("OptimumConfigRegistryModule", (m) => {
	const configRegistry = m.contract("OptimumConfigRegistry")
	return { configRegistry }
})
