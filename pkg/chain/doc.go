// Package chain maps chain names to numeric IDs and beacon genesis timestamps.
// Built-ins (mainnet, hoodi, sepolia) are always present; devnets and testnets
// are added via RegisterChain at service startup, before lookups run. Entries
// are append-only: nothing can be removed or redefined.
package chain
