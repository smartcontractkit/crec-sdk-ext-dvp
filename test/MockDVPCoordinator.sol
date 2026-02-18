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

    // --- Settlement struct (mirrors DVPCoordinator.Settlement) ---

    struct DeliveryInfo {
        uint64 AssetDestinationChainSelector;
        uint64 AssetSourceChainSelector;
        uint64 PaymentDestinationChainSelector;
        uint64 PaymentSourceChainSelector;
    }

    struct PartyInfo {
        bytes BuyerDestinationAddress;
        bytes BuyerSourceAddress;
        bytes ExecutorAddress;
        bytes SellerDestinationAddress;
        bytes SellerSourceAddress;
    }

    struct TokenInfo {
        uint256 AssetTokenAmount;
        bytes AssetTokenDestinationAddress;
        bytes AssetTokenSourceAddress;
        uint8 AssetTokenType;
        uint8 PaymentCurrency;
        uint256 PaymentTokenAmount;
        bytes PaymentTokenDestinationAddress;
        bytes PaymentTokenSourceAddress;
        uint8 PaymentTokenType;
    }

    struct Settlement {
        uint256 CcipCallbackGasLimit;
        bytes Data;
        DeliveryInfo DeliveryInfo;
        uint256 ExecuteAfter;
        uint256 Expiration;
        PartyInfo PartyInfo;
        bytes32 SecretHash;
        uint256 SettlementId;
        TokenInfo TokenInfo;
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
            CcipCallbackGasLimit: 200000,
            Data: hex"deadbeef",
            DeliveryInfo: DeliveryInfo({
                AssetDestinationChainSelector: 16015286601757825753,
                AssetSourceChainSelector: 16015286601757825753,
                PaymentDestinationChainSelector: 16015286601757825753,
                PaymentSourceChainSelector: 16015286601757825753
            }),
            ExecuteAfter: 0,
            Expiration: 1893456000, // 2030-01-01
            PartyInfo: PartyInfo({
                BuyerDestinationAddress: abi.encodePacked(msg.sender),
                BuyerSourceAddress: abi.encodePacked(msg.sender),
                ExecutorAddress: abi.encodePacked(msg.sender),
                SellerDestinationAddress: abi.encodePacked(address(0xdead)),
                SellerSourceAddress: abi.encodePacked(address(0xdead))
            }),
            SecretHash: bytes32(0),
            SettlementId: 1,
            TokenInfo: TokenInfo({
                AssetTokenAmount: 1000,
                AssetTokenDestinationAddress: abi.encodePacked(msg.sender),
                AssetTokenSourceAddress: abi.encodePacked(address(0xdead)),
                AssetTokenType: 1,
                PaymentCurrency: 1,
                PaymentTokenAmount: 500,
                PaymentTokenDestinationAddress: abi.encodePacked(address(0xdead)),
                PaymentTokenSourceAddress: abi.encodePacked(msg.sender),
                PaymentTokenType: 1
            })
        });
    }
}
