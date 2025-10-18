// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.28;

/// @title OptimumConfigRegistry
/// @notice Stores and manages unique configurations identified by a (networkId, clusterId) pair.
/// @author Optimum Team
/// @dev
/// - Each (networkId, clusterId) combination is unique.
/// - The same networkId can appear in multiple clusters, and the same clusterId can appear in multiple networks.
/// - Configuration payloads are stored as base64-encoded strings (callDataBase64).
/// - Only the contract owner can create, update, or remove configurations.
contract OptimumConfigRegistry {
    /// Errors
    error OptimumConfigRegistry_OnlyOwner();
    error OptimumConfigRegistry_ZeroAddress();
    error OptimumConfigRegistry_EmptyNetwork();
    error OptimumConfigRegistry_EmptyCluster();
    error OptimumConfigRegistry_PairExists();
    error OptimumConfigRegistry_NotFound();

    /// OWNERSHIP
    /// @notice The address with permission to manage configurations.
    address public owner;

    /// @notice Emitted when ownership of the contract is transferred.
    /// @param previousOwner The previous owner of the contract.
    /// @param newOwner The new owner of the contract.
    event OwnershipTransferred(address indexed previousOwner, address indexed newOwner);

    /// @dev Restricts a function to be callable only by the owner.
    modifier onlyOwner() {
        if (msg.sender != owner) revert OptimumConfigRegistry_OnlyOwner();
        _;
    }

    /// @notice Initializes the contract, setting the deployer as the owner.
    constructor() {
        owner = msg.sender;
        emit OwnershipTransferred(address(0), msg.sender);
    }

    /// @notice Transfers ownership of the contract to a new address.
    /// @param newOwner The address of the new owner.
    function transferOwnership(address newOwner) external onlyOwner {
        if (newOwner == address(0)) revert OptimumConfigRegistry_ZeroAddress();
        emit OwnershipTransferred(owner, newOwner);
        owner = newOwner;
    }

    /// STORAGE

    /// @notice Represents a single configuration entry.
    /// @param callDataBase64 Base64-encoded data string representing configuration content.
    struct Config {
        string callDataBase64;
    }

    /// @dev Mapping of a (networkId, clusterId) pair hash → Config.
    /// The key is keccak256(abi.encodePacked(networkId, clusterId)).
    mapping(bytes32 => Config) private _configs;

    /// EVENTS

    /// @notice Emitted when a configuration is created.
    /// @param id        The computed pair key (keccak256 of networkId and clusterId).
    /// @param networkId Identifier of the network.
    /// @param clusterId Identifier of the cluster.
    /// @param callDataBase64 The base64-encoded configuration payload.
    event ConfigCreated(
        uint256 indexed id,
        bytes32 indexed networkId,
        bytes32 indexed clusterId,
        string callDataBase64
    );

    /// @notice Emitted when a configuration is updated.
    /// @param id        The computed pair key (keccak256 of networkId and clusterId).
    /// @param oldBase64 The previous base64 payload.
    /// @param newBase64 The updated base64 payload.
    event ConfigUpdated(uint256 indexed id, string oldBase64, string newBase64);

    /// @notice Emitted when a configuration is removed.
    /// @param id        The computed pair key (keccak256 of networkId and clusterId).
    /// @param networkId Identifier of the network.
    /// @param clusterId Identifier of the cluster.
    event ConfigRemoved(uint256 indexed id, bytes32 indexed networkId, bytes32 indexed clusterId);

    /// WRITE METHODS

    /// @notice Creates a new configuration entry for the specified (networkId, clusterId) pair.
    /// @dev Reverts if the pair already exists.
    /// @param networkId Identifier of the network.
    /// @param clusterId Identifier of the cluster.
    /// @param callDataBase64 Base64-encoded payload to associate with the pair.
    function createConfig(
        bytes32 networkId,
        bytes32 clusterId,
        string calldata callDataBase64
    ) external onlyOwner {
        if (networkId == bytes32(0)) revert OptimumConfigRegistry_EmptyNetwork();
        if (clusterId == bytes32(0)) revert OptimumConfigRegistry_EmptyCluster();

        bytes32 pairKey = _pairKey(networkId, clusterId);
        if (bytes(_configs[pairKey].callDataBase64).length != 0) revert OptimumConfigRegistry_PairExists();

        _configs[pairKey] = Config({ callDataBase64: callDataBase64 });

        emit ConfigCreated(uint256(pairKey), networkId, clusterId, callDataBase64);
    }

    /// @notice Updates the base64-encoded payload for an existing configuration pair.
    /// @dev Reverts if the pair does not exist.
    /// @param networkId Identifier of the network.
    /// @param clusterId Identifier of the cluster.
    /// @param newCallDataBase64 New base64-encoded payload.
    function updateCallDataBase64(
        bytes32 networkId,
        bytes32 clusterId,
        string calldata newCallDataBase64
    ) external onlyOwner {
        bytes32 pairKey = _pairKey(networkId, clusterId);
        Config storage cfg = _configs[pairKey];
        if (bytes(cfg.callDataBase64).length == 0) revert OptimumConfigRegistry_NotFound();

        string memory old = cfg.callDataBase64;
        cfg.callDataBase64 = newCallDataBase64;

        emit ConfigUpdated(uint256(pairKey), old, newCallDataBase64);
    }

    /// @notice Removes a configuration for the specified (networkId, clusterId) pair.
    /// @dev Reverts if the pair does not exist.
    /// @param networkId Identifier of the network.
    /// @param clusterId Identifier of the cluster.
    function removeConfig(bytes32 networkId, bytes32 clusterId) external onlyOwner {
        bytes32 pairKey = _pairKey(networkId, clusterId);
        Config storage cfg = _configs[pairKey];
        if (bytes(cfg.callDataBase64).length == 0) revert OptimumConfigRegistry_NotFound();

        delete _configs[pairKey];
        emit ConfigRemoved(uint256(pairKey), networkId, clusterId);
    }

    /// READ METHODS

    /// @notice Returns a configuration for the given (networkId, clusterId) pair.
    /// @dev Reverts if the configuration does not exist.
    /// @param networkId Identifier of the network.
    /// @param clusterId Identifier of the cluster.
    /// @return cfg The Config struct containing the stored base64 payload.
    function getConfig(
        bytes32 networkId,
        bytes32 clusterId
    ) external view returns (Config memory cfg) {
        bytes32 pairKey = _pairKey(networkId, clusterId);
        cfg = _configs[pairKey];
        if (bytes(cfg.callDataBase64).length == 0) revert OptimumConfigRegistry_NotFound();
    }

    /// INTERNAL UTILITIES

    /// @notice Computes the unique mapping key for a (networkId, clusterId) pair.
    /// @dev Uses keccak256(abi.encodePacked()) for compact hashing.
    /// @param networkId Identifier of the network.
    /// @param clusterId Identifier of the cluster.
    /// @return key The computed pair key.
    function _pairKey(bytes32 networkId, bytes32 clusterId) internal pure returns (bytes32 key) {
        key = keccak256(abi.encodePacked(networkId, clusterId));
    }
}
