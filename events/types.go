package events

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// PartyInfo represents party addresses for a DvP settlement.
// Maps to struct PartyInfo in CCIPDVPCoordinator.
type PartyInfo struct {
	// BuyerSourceAddress is the buyer's address on the source chain.
	BuyerSourceAddress common.Address
	// BuyerDestinationAddress is the buyer's address on the destination chain.
	BuyerDestinationAddress common.Address
	// SellerSourceAddress is the seller's address on the source chain.
	SellerSourceAddress common.Address
	// SellerDestinationAddress is the seller's address on the destination chain.
	SellerDestinationAddress common.Address
	// ExecutorAddress is the address authorized to execute the settlement.
	ExecutorAddress common.Address
}

// TokenInfo represents token and payment details for a DvP settlement.
// Maps to struct TokenInfo in CCIPDVPCoordinator.
// Field order must match ABI for proposeSettlement.
type TokenInfo struct {
	// PaymentTokenSourceAddress is the payment token contract on the source chain.
	PaymentTokenSourceAddress common.Address
	// PaymentTokenDestinationAddress is the payment token contract on the destination chain.
	PaymentTokenDestinationAddress common.Address
	// AssetTokenSourceAddress is the asset token contract on the source chain.
	AssetTokenSourceAddress common.Address
	// AssetTokenDestinationAddress is the asset token contract on the destination chain.
	AssetTokenDestinationAddress common.Address
	// PaymentTokenAmount is the amount of payment token to transfer.
	PaymentTokenAmount *big.Int
	// AssetTokenAmount is the amount of asset token to transfer.
	AssetTokenAmount *big.Int
	// PaymentCurrency is the ISO 4217 currency code (see currency.Map).
	PaymentCurrency uint8
	// PaymentLockType indicates the lock mechanism for payment tokens (see LockType constants).
	PaymentLockType uint8
	// AssetLockType indicates the lock mechanism for asset tokens (see LockType constants).
	AssetLockType uint8
}

// DeliveryInfo represents chain selectors for cross-chain delivery.
// Maps to struct DeliveryInfo in CCIPDVPCoordinator.
type DeliveryInfo struct {
	// PaymentSourceChainSelector is the CCIP chain selector for the payment source.
	PaymentSourceChainSelector uint64
	// PaymentDestinationChainSelector is the CCIP chain selector for the payment destination.
	PaymentDestinationChainSelector uint64
	// AssetSourceChainSelector is the CCIP chain selector for the asset source.
	AssetSourceChainSelector uint64
	// AssetDestinationChainSelector is the CCIP chain selector for the asset destination.
	AssetDestinationChainSelector uint64
}

// Settlement represents a DvP settlement proposal.
// Maps to struct Settlement in CCIPDVPCoordinator.
// Field order must match ABI for proposeSettlement.
type Settlement struct {
	SettlementId         *big.Int
	PartyInfo            PartyInfo
	TokenInfo            TokenInfo
	DeliveryInfo         DeliveryInfo
	SecretHash           common.Hash
	ExecuteAfter         *big.Int
	Expiration           *big.Int
	CcipCallbackGasLimit uint32
	Data                 []byte
}
