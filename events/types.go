package events

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// PartyInfo represents party addresses for a DvP settlement.
// Maps to struct PartyInfo in CCIPDVPCoordinator.
type PartyInfo struct {
	BuyerSourceAddress      common.Address
	BuyerDestinationAddress common.Address
	SellerSourceAddress     common.Address
	SellerDestinationAddress common.Address
	ExecutorAddress         common.Address
}

// TokenInfo represents token and payment details for a DvP settlement.
// Maps to struct TokenInfo in CCIPDVPCoordinator.
// Field order must match ABI for proposeSettlement.
type TokenInfo struct {
	PaymentTokenSourceAddress      common.Address
	PaymentTokenDestinationAddress common.Address
	AssetTokenSourceAddress        common.Address
	AssetTokenDestinationAddress   common.Address
	PaymentCurrency                uint8
	PaymentTokenAmount             *big.Int
	AssetTokenAmount               *big.Int
	PaymentTokenType               uint8
	AssetTokenType                 uint8
}

// DeliveryInfo represents chain selectors for cross-chain delivery.
// Maps to struct DeliveryInfo in CCIPDVPCoordinator.
type DeliveryInfo struct {
	PaymentSourceChainSelector      uint64
	PaymentDestinationChainSelector uint64
	AssetSourceChainSelector        uint64
	AssetDestinationChainSelector   uint64
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
