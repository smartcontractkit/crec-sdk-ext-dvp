// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

/// @title MockDVPCoordinator
/// @notice Minimal mock of CCIPDVPCoordinator for testing the DVP extension watcher.
///         Emits settlement events and returns hardcoded data from getSettlement().
contract MockDVPCoordinator {
    // --- Events (must match CCIPDVPCoordinator ABI) ---

    event SettlementOpened(uint256 indexed settlementId, bytes32 indexed settlementHash);
    event SettlementAccepted(uint256 indexed settlementId, bytes32 indexed settlementHash);
    event SettlementClosing(uint256 indexed settlementId, bytes32 indexed settlementHash);
    event SettlementSettled(uint256 indexed settlementId, bytes32 indexed settlementHash);
    event SettlementCanceling(uint256 indexed settlementId, bytes32 indexed settlementHash);
    event SettlementCanceled(uint256 indexed settlementId, bytes32 indexed settlementHash);

    // --- Settlement struct (mirrors CCIPDVPCoordinator.Settlement) ---

    struct PartyInfo {
        address buyerSourceAddress;
        address buyerDestinationAddress;
        address sellerSourceAddress;
        address sellerDestinationAddress;
        address executorAddress;
    }

    struct TokenInfo {
        address paymentTokenSourceAddress;
        address paymentTokenDestinationAddress;
        address assetTokenSourceAddress;
        address assetTokenDestinationAddress;
        uint256 paymentTokenAmount;
        uint256 assetTokenAmount;
        uint8 paymentCurrency;
        uint8 paymentLockType;
        uint8 assetLockType;
    }

    struct DeliveryInfo {
        uint64 paymentSourceChainSelector;
        uint64 paymentDestinationChainSelector;
        uint64 assetSourceChainSelector;
        uint64 assetDestinationChainSelector;
    }

    struct Settlement {
        uint256 settlementId;
        PartyInfo partyInfo;
        TokenInfo tokenInfo;
        DeliveryInfo deliveryInfo;
        bytes32 secretHash;
        uint48 executeAfter;
        uint48 expiration;
        uint32 ccipCallbackGasLimit;
        bytes data;
    }

    uint256 private _nextId = 1;

    /// @notice Emit a SettlementAccepted event with a deterministic hash.
    function emitSettlementAccepted() external returns (uint256 settlementId, bytes32 settlementHash) {
        settlementId = _nextId++;
        settlementHash = keccak256(abi.encodePacked(settlementId, msg.sender));
        emit SettlementAccepted(settlementId, settlementHash);
    }

    /// @notice Returns a hardcoded Settlement for any hash.
    ///         The DVP watcher handler calls this to enrich events.
    function getSettlement(bytes32 /* settlementHash */) external view returns (Settlement memory) {
        return Settlement({
            settlementId: 1,
            partyInfo: PartyInfo({
                buyerSourceAddress: msg.sender,
                buyerDestinationAddress: msg.sender,
                sellerSourceAddress: address(0xdead),
                sellerDestinationAddress: address(0xdead),
                executorAddress: msg.sender
            }),
            tokenInfo: TokenInfo({
                paymentTokenSourceAddress: msg.sender,
                paymentTokenDestinationAddress: address(0xdead),
                assetTokenSourceAddress: address(0xdead),
                assetTokenDestinationAddress: msg.sender,
                paymentTokenAmount: 500,
                assetTokenAmount: 1000,
                paymentCurrency: 1,
                paymentLockType: 1,
                assetLockType: 1
            }),
            deliveryInfo: DeliveryInfo({
                paymentSourceChainSelector: 16015286601757825753,
                paymentDestinationChainSelector: 16015286601757825753,
                assetSourceChainSelector: 16015286601757825753,
                assetDestinationChainSelector: 16015286601757825753
            }),
            secretHash: bytes32(0),
            executeAfter: 0,
            expiration: 1893456000, // 2030-01-01
            ccipCallbackGasLimit: 200000,
            data: hex"deadbeef"
        });
    }
}
