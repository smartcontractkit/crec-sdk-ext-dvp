// Package dvp provides a CREC SDK extension for Delivery versus Payment (DvP) settlement operations.
//
// This extension enables the preparation of DvP settlement operations for atomic
// exchange of assets between counterparties on blockchain networks.
//
// # Usage
//
//	ext, err := dvp.New(&dvp.Options{
//		DvpCoordinatorAddress: "0x...",
//		AccountAddress:        "0x...",
//	})
//	if err != nil {
//		return err
//	}
//
//	op, err := ext.PrepareProposeSettlementOperation(settlement)
package dvp

import (
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"

	apiClient "github.com/smartcontractkit/crec-api-go/client"
	"github.com/smartcontractkit/crec-api-go/services/dvp/gen/contract"
	"github.com/smartcontractkit/crec-api-go/services/dvp/gen/events"

	"github.com/smartcontractkit/crec-sdk/interfaces/erc20"
	"github.com/smartcontractkit/crec-sdk/interfaces/holdmanager"
	transactTypes "github.com/smartcontractkit/crec-sdk/transact/types"
)

// Token type constants for DvP settlements.
const (
	TokenTypeNone = iota
	TokenTypeERC20
	TokenTypeERC3643
)

// Settlement status constants.
const (
	SettlementStatusNew = iota
	SettlementStatusOpen
	SettlementStatusAccepted
	SettlementStatusClosing
	SettlementStatusSettled
	SettlementStatusCanceled
)

// Event name constants for DvP settlement events.
const (
	ServiceName         = "dvp"
	SettlementOpened    = "SettlementOpened"
	SettlementAccepted  = "SettlementAccepted"
	SettlementCanceling = "SettlementCanceling"
	SettlementCanceled  = "SettlementCanceled"
	SettlementClosing   = "SettlementClosing"
	SettlementSettled   = "SettlementSettled"
)

// Options defines the configuration for creating a new CREC DvP extension.
type Options struct {
	// Logger is an optional logger instance. If nil, a default logger is created.
	Logger *zerolog.Logger

	// DvpCoordinatorAddress is the address of the DvP coordinator contract.
	DvpCoordinatorAddress string

	// AccountAddress is the address of the account performing the DvP operations.
	AccountAddress string
}

// Extension provides methods for preparing DvP settlement operations.
type Extension struct {
	logger                *zerolog.Logger
	dvpCoordinatorAddress common.Address
	accountAddress        common.Address
}

// New creates a new CREC DvP extension with the provided options.
// Returns a pointer to the Extension and an error if any issues occur during initialization.
func New(opts *Options) (*Extension, error) {
	if opts == nil {
		return nil, fmt.Errorf("options is required")
	}

	logger := opts.Logger
	if logger == nil {
		lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger = &lgr
	}

	logger.Debug().Msg("Creating CREC DvP extension")

	return &Extension{
		logger:                logger,
		dvpCoordinatorAddress: common.HexToAddress(opts.DvpCoordinatorAddress),
		accountAddress:        common.HexToAddress(opts.AccountAddress),
	}, nil
}

// DecodeDvpEvent decodes a DvP settlement event from the provided CREC event.
func (e *Extension) DecodeDvpEvent(event *apiClient.Event) (*events.DvpEvent, error) {
	jsonBytes, err := e.toJson(event)
	if err != nil {
		return nil, fmt.Errorf("failed to decode event: %w", err)
	}

	var dvpEvent events.DvpEvent
	err = dvpEvent.UnmarshalJSON(jsonBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal event json: %w", err)
	}

	return &dvpEvent, nil
}

// PrepareProposeSettlementOperation prepares a DvP propose settlement operation.
// It assumes a token approval has already been issued for the asset token.
//
// Parameters:
//   - settlement: The settlement details to be included in the operation.
func (e *Extension) PrepareProposeSettlementOperation(settlement *contract.Settlement) (
	*transactTypes.Operation, error,
) {
	settlementHash, err := e.HashSettlement(settlement)
	if err != nil {
		return nil, fmt.Errorf("failed to hash settlement: %w", err)
	}

	e.logger.Trace().
		Str("settlementHash", settlementHash.Hex()).
		Msg("Preparing proposeSettlement operation")

	return e.prepareSettlementOperation(
		"proposeSettlement",
		settlementHash,
		nil,
		settlement,
	)
}

// PrepareProposeSettlementWithTokenApprovalOperation prepares a DvP propose settlement operation,
// including a token approval transaction.
//
// Parameters:
//   - settlement: The settlement details to be included in the operation.
func (e *Extension) PrepareProposeSettlementWithTokenApprovalOperation(settlement *contract.Settlement) (
	*transactTypes.Operation, error,
) {
	settlementHash, err := e.HashSettlement(settlement)
	if err != nil {
		return nil, fmt.Errorf("failed to hash settlement: %w", err)
	}

	e.logger.Trace().
		Str("settlementHash", settlementHash.Hex()).
		Msg("Preparing proposeSettlement with token approval operation")

	approveTransaction, err := e.prepareTokenApproveTransaction(
		settlement.TokenInfo.AssetTokenSourceAddress, settlement.TokenInfo.AssetTokenAmount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare approve transaction: %w", err)
	}

	return e.prepareSettlementOperation(
		"proposeSettlement",
		settlementHash,
		[]transactTypes.Transaction{*approveTransaction},
		settlement,
	)
}

// PrepareProposeSettlementWithTokenHoldOperation prepares a DvP propose settlement operation,
// including issuing a token hold for the asset token.
//
// Parameters:
//   - settlement: The settlement details to be included in the operation.
//   - holdManagerAddress: The address of the hold manager contract to be used for the token hold.
func (e *Extension) PrepareProposeSettlementWithTokenHoldOperation(
	settlement *contract.Settlement, holdManagerAddress common.Address,
) (
	*transactTypes.Operation, error,
) {
	if settlement.TokenInfo.AssetTokenType != TokenTypeERC3643 {
		return nil, fmt.Errorf("token hold is only supported for ERC3643 asset tokens")
	}

	settlementHash, err := e.HashSettlement(settlement)
	if err != nil {
		return nil, fmt.Errorf("failed to hash settlement: %w", err)
	}

	e.logger.Trace().
		Str("settlementHash", settlementHash.Hex()).
		Msg("Preparing proposeSettlement with token hold operation")

	holdTransaction, err := e.prepareTokenHoldTransaction(
		holdManagerAddress,
		settlementHash,
		settlement.TokenInfo.AssetTokenSourceAddress,
		settlement.PartyInfo.SellerSourceAddress,
		settlement.Expiration,
		settlement.TokenInfo.AssetTokenAmount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare hold transaction: %w", err)
	}

	return e.prepareSettlementOperation(
		"proposeSettlement",
		settlementHash,
		[]transactTypes.Transaction{*holdTransaction},
		settlement,
	)
}

// PrepareAcceptSettlementOperation prepares a DvP accept settlement operation.
// It constructs the necessary transaction to accept a settlement based on its hash.
//
// Parameters:
//   - settlementHash: The hash of the settlement to be accepted.
func (e *Extension) PrepareAcceptSettlementOperation(settlementHash common.Hash) (*transactTypes.Operation, error) {
	e.logger.Trace().
		Str("settlementHash", settlementHash.Hex()).
		Msg("Preparing acceptSettlement operation")

	return e.prepareSettlementOperation(
		"acceptSettlement",
		settlementHash,
		nil,
		settlementHash,
	)
}

// PrepareExecuteSettlementOperation prepares a DvP executeSettlement operation.
// It constructs the necessary transaction to execute a settlement based on its hash.
//
// Parameters:
//   - settlementHash: The hash of the settlement to be executed.
func (e *Extension) PrepareExecuteSettlementOperation(settlementHash common.Hash) (*transactTypes.Operation, error) {
	e.logger.Trace().
		Str("settlementHash", settlementHash.Hex()).
		Msg("Preparing executeSettlement operation")

	return e.prepareSettlementOperation(
		"executeSettlement",
		settlementHash,
		nil,
		settlementHash,
	)
}

// HashSettlement computes the hash of a DvP settlement.
//
// Parameters:
//   - settlement: The settlement to be hashed.
func (e *Extension) HashSettlement(settlement *contract.Settlement) (common.Hash, error) {
	uint256Ty, _ := abi.NewType("uint256", "", nil)
	uint64Ty, _ := abi.NewType("uint64", "", nil)
	uint48Ty, _ := abi.NewType("uint48", "", nil)
	uint8Ty, _ := abi.NewType("uint8", "", nil)
	addressTy, _ := abi.NewType("address", "", nil)
	bytes32Ty, _ := abi.NewType("bytes32", "", nil)
	bytesTy, _ := abi.NewType("bytes", "", nil)

	partyInfoArgs := abi.Arguments{
		{Type: addressTy}, // buyerSourceAddress
		{Type: addressTy}, // buyerDestinationAddress
		{Type: addressTy}, // sellerSourceAddress
		{Type: addressTy}, // sellerDestinationAddress
		{Type: addressTy}, // executorAddress
	}
	partyInfoData, err := partyInfoArgs.Pack(
		settlement.PartyInfo.BuyerSourceAddress,
		settlement.PartyInfo.BuyerDestinationAddress,
		settlement.PartyInfo.SellerSourceAddress,
		settlement.PartyInfo.SellerDestinationAddress,
		settlement.PartyInfo.ExecutorAddress,
	)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack party info data: %w", err)
	}

	tokenInfoArgs := abi.Arguments{
		{Type: uint256Ty}, // paymentTokenAmount
		{Type: uint256Ty}, // assetTokenAmount
		{Type: addressTy}, // paymentTokenSourceAddress
		{Type: addressTy}, // paymentTokenDestinationAddress
		{Type: addressTy}, // assetTokenSourceAddress
		{Type: addressTy}, // assetTokenDestinationAddress
		{Type: uint8Ty},   // paymentTokenType
		{Type: uint8Ty},   // assetTokenType
	}
	tokenInfoData, err := tokenInfoArgs.Pack(
		settlement.TokenInfo.PaymentTokenAmount,
		settlement.TokenInfo.AssetTokenAmount,
		settlement.TokenInfo.PaymentTokenSourceAddress,
		settlement.TokenInfo.PaymentTokenDestinationAddress,
		settlement.TokenInfo.AssetTokenSourceAddress,
		settlement.TokenInfo.AssetTokenDestinationAddress,
		settlement.TokenInfo.PaymentTokenType,
		settlement.TokenInfo.AssetTokenType,
	)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack token info data: %w", err)
	}

	deliveryInfoArgs := abi.Arguments{
		{Type: uint64Ty}, // paymentSourceChainSelector
		{Type: uint64Ty}, // paymentDestinationChainSelector
		{Type: uint64Ty}, // assetSourceChainSelector
		{Type: uint64Ty}, // assetDestinationChainSelector
	}
	deliveryInfoData, err := deliveryInfoArgs.Pack(
		settlement.DeliveryInfo.PaymentSourceChainSelector,
		settlement.DeliveryInfo.PaymentDestinationChainSelector,
		settlement.DeliveryInfo.AssetSourceChainSelector,
		settlement.DeliveryInfo.AssetDestinationChainSelector,
	)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack delivery info data: %w", err)
	}

	partyInfoHash := crypto.Keccak256Hash(partyInfoData)
	tokenInfoHash := crypto.Keccak256Hash(tokenInfoData)
	deliveryDataHash := crypto.Keccak256Hash(deliveryInfoData)

	settlementInfoArgs := abi.Arguments{
		{Type: bytes32Ty}, // keccak256("CCIP_DVP_COORDINATOR_V1_SETTLEMENT")
		{Type: uint256Ty}, // settlementId
		{Type: bytes32Ty}, // partyInfoHash
		{Type: bytes32Ty}, // tokenInfoHash
		{Type: bytes32Ty}, // deliveryDataHash
		{Type: bytes32Ty}, // secretHash
		{Type: uint48Ty},  // expiration
		{Type: bytesTy},   // data
	}
	settlementInfoData, err := settlementInfoArgs.Pack(
		crypto.Keccak256Hash([]byte("CCIP_DVP_COORDINATOR_V1_SETTLEMENT")),
		settlement.SettlementId,
		partyInfoHash,
		tokenInfoHash,
		deliveryDataHash,
		settlement.SecretHash,
		settlement.Expiration,
		settlement.Data,
	)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to pack settlement info data: %w", err)
	}

	return crypto.Keccak256Hash(settlementInfoData), nil
}

// toJson decodes an encoded VerifiableEvent from a CREC event into a JSON byte slice.
//
// NOTE: This method is temporarily disabled pending migration to new Event structure.
func (e *Extension) toJson(event *apiClient.Event) ([]byte, error) {
	return nil, fmt.Errorf("toJson is temporarily disabled - needs migration to new Event structure")
}

// prepareSettlementOperation is a helper function that abstracts the common logic for preparing settlement operations.
func (e *Extension) prepareSettlementOperation(
	operationName string,
	settlementHash common.Hash,
	extraTransactions []transactTypes.Transaction,
	packArgs ...interface{},
) (*transactTypes.Operation, error) {
	abiEncoder, err := contract.ContractMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get DVP ABI: %w", err)
	}

	calldata, err := abiEncoder.Pack(operationName, packArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack calldata for %s: %w", operationName, err)
	}

	mainTransaction := transactTypes.Transaction{
		To:    e.dvpCoordinatorAddress,
		Value: big.NewInt(0),
		Data:  calldata,
	}

	transactions := append(extraTransactions, mainTransaction)

	return &transactTypes.Operation{
		ID:           crypto.Keccak256Hash(append(settlementHash[:], []byte(operationName)...)).Big(),
		Account:      e.accountAddress,
		Transactions: transactions,
	}, nil
}

func (e *Extension) prepareTokenApproveTransaction(
	tokenAddress common.Address, tokenAmount *big.Int,
) (*transactTypes.Transaction, error) {
	erc20Abi, err := erc20.Erc20MetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get ERC20 ABI: %w", err)
	}

	calldata, err := erc20Abi.Pack("approve", e.dvpCoordinatorAddress, tokenAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to pack calldata for token approve: %w", err)
	}

	return &transactTypes.Transaction{
		To:    tokenAddress,
		Value: big.NewInt(0),
		Data:  calldata,
	}, nil
}

func (e *Extension) prepareTokenHoldTransaction(
	holdManagerAddress common.Address, holdId common.Hash, tokenAddress common.Address, sender common.Address,
	expiresAt *big.Int, tokenAmount *big.Int,
) (*transactTypes.Transaction, error) {
	holdmanagerAbi, err := holdmanager.HoldmanagerMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get HoldManager ABI: %w", err)
	}

	hold := holdmanager.IHoldManagerHold{
		Token:     tokenAddress,
		Sender:    sender,
		Recipient: e.dvpCoordinatorAddress,
		Executor:  e.dvpCoordinatorAddress,
		ExpiresAt: expiresAt,
		Value:     tokenAmount,
	}

	calldata, err := holdmanagerAbi.Pack("createHold", holdId, hold)
	if err != nil {
		return nil, fmt.Errorf("failed to pack calldata for token hold: %w", err)
	}

	return &transactTypes.Transaction{
		To:    holdManagerAddress,
		Value: big.NewInt(0),
		Data:  calldata,
	}, nil
}

