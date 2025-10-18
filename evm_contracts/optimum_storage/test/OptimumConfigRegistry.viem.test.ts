import assert from "node:assert/strict"
import { describe, it, beforeEach } from "node:test"
import { toBytes, keccak256 } from "viem"

import { network } from "hardhat"

describe("OptimumConfigRegistry (viem)", async function () {
	const { viem } = await network.connect()

	// helpers
	const toB32 = (s: string) => keccak256(toBytes(s)) // bytes32
	const b64 = (s: string) => Buffer.from(s).toString("base64")

	let ownerClient: any
	let strangerClient: any
	let registry: any

	beforeEach(async () => {
		const [owner, stranger] = await viem.getWalletClients()
		ownerClient = owner
		strangerClient = stranger

		registry = await viem.deployContract("OptimumConfigRegistry", [], {
			client: ownerClient,
		})
	})

	it("create → get → update → get → remove → get(revert)", async () => {
		const net = toB32("ethereum-mainnet")
		const cl = toB32("optimum_hoodi_v0_1")
		const data1 = b64('{"x":1}')
		const data2 = b64('{"x":2}')

		// create
		await registry.write.createConfig([net, cl, data1], {
			account: ownerClient.account,
		})

		// get
		const cfg1 = await registry.read.getConfig([net, cl])
		assert.equal(cfg1.callDataBase64, data1)

		// update
		await registry.write.updateCallDataBase64([net, cl, data2], {
			account: ownerClient.account,
		})
		const cfg2 = await registry.read.getConfig([net, cl])
		assert.equal(cfg2.callDataBase64, data2)

		// remove
		await registry.write.removeConfig([net, cl], {
			account: ownerClient.account,
		})

		// must revert now
		await assert.rejects(() => registry.read.getConfig([net, cl]), /NOT_FOUND/)
	})

	it("enforces uniqueness only for the (networkId, clusterId) pair", async () => {
		const netA = toB32("ethereum-mainnet")
		const clA = toB32("cluster-a")
		const clB = toB32("cluster-b")
		const data = b64("payload")

		// same network, different clusters → allowed
		await registry.write.createConfig([netA, clA, data], {
			account: ownerClient.account,
		})
		await registry.write.createConfig([netA, clB, data], {
			account: ownerClient.account,
		})

		// duplicate pair → revert
		await assert.rejects(
			() =>
				registry.write.createConfig([netA, clA, data], {
					account: ownerClient.account,
				}),
			/PAIR_EXISTS/,
		)
	})

	it("only owner can write", async () => {
		const net = toB32("ethereum-mainnet")
		const cl = toB32("cluster-a")
		const data = b64("x")

		await assert.rejects(
			() =>
				registry.write.createConfig([net, cl, data], {
					account: strangerClient.account,
				}),
			/ONLY_OWNER/,
		)

		// owner creates
		await registry.write.createConfig([net, cl, data], {
			account: ownerClient.account,
		})

		await assert.rejects(
			() =>
				registry.write.updateCallDataBase64([net, cl, b64("y")], {
					account: strangerClient.account,
				}),
			/ONLY_OWNER/,
		)
		await assert.rejects(
			() =>
				registry.write.removeConfig([net, cl], {
					account: strangerClient.account,
				}),
			/ONLY_OWNER/,
		)
	})

	it("guards NOT_FOUND on update/remove", async () => {
		const net = toB32("ethereum-mainnet")
		const cl = toB32("cluster-z")

		await assert.rejects(
			() =>
				registry.write.updateCallDataBase64([net, cl, b64("nope")], {
					account: ownerClient.account,
				}),
			/NOT_FOUND/,
		)
		await assert.rejects(
			() =>
				registry.write.removeConfig([net, cl], {
					account: ownerClient.account,
				}),
			/NOT_FOUND/,
		)
	})
})
