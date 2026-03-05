package operations

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/smartcontractkit/crec-sdk-ext-dvp/events"
	"github.com/smartcontractkit/crec-sdk/interfaces/erc20"
	"github.com/smartcontractkit/crec-sdk/interfaces/holdmanager"
	transactTypes "github.com/smartcontractkit/crec-sdk/transact/types"
)

// HashSettlement computes the hash of a DvP settlement (method form for backward compatibility).
func (e *Extension) HashSettlement(settlement *events.Settlement) (common.Hash, error) {
	return HashSettlement(settlement)
}

// PrepareProposeSettlementOperation prepares a DvP propose settlement operation.
// It assumes a token approval has already been issued for the asset token.
func (e *Extension) PrepareProposeSettlementOperation(settlement *events.Settlement) (*transactTypes.Operation, error) {
	settlementHash, err := HashSettlement(settlement)
	if err != nil {
		return nil, fmt.Errorf("failed to hash settlement: %w", err)
	}

	e.logger.Debug("Preparing proposeSettlement operation", "settlementHash", settlementHash.Hex())

	return e.prepareSettlementOperation("proposeSettlement", settlementHash, nil, settlement)
}

// PrepareProposeSettlementWithTokenApprovalOperation prepares a DvP propose settlement operation,
// including a token approval transaction.
func (e *Extension) PrepareProposeSettlementWithTokenApprovalOperation(settlement *events.Settlement) (*transactTypes.Operation, error) {
	settlementHash, err := HashSettlement(settlement)
	if err != nil {
		return nil, fmt.Errorf("failed to hash settlement: %w", err)
	}

	e.logger.Debug("Preparing proposeSettlement with token approval operation", "settlementHash", settlementHash.Hex())

	approveTransaction, err := e.prepareTokenApproveTransaction(
		settlement.TokenInfo.AssetTokenSourceAddress, settlement.TokenInfo.AssetTokenAmount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare approve transaction: %w", err)
	}

	return e.prepareSettlementOperation("proposeSettlement", settlementHash, []transactTypes.Transaction{*approveTransaction}, settlement)
}

// PrepareProposeSettlementWithTokenHoldOperation prepares a DvP propose settlement operation,
// including issuing a token hold for the asset token. Only supported for ERC3643 asset tokens.
func (e *Extension) PrepareProposeSettlementWithTokenHoldOperation(
	settlement *events.Settlement, holdManagerAddress common.Address,
) (*transactTypes.Operation, error) {
	if settlement.TokenInfo.AssetTokenType != events.TokenTypeERC3643 {
		return nil, fmt.Errorf("token hold is only supported for ERC3643 asset tokens")
	}

	settlementHash, err := HashSettlement(settlement)
	if err != nil {
		return nil, fmt.Errorf("failed to hash settlement: %w", err)
	}

	e.logger.Debug("Preparing proposeSettlement with token hold operation", "settlementHash", settlementHash.Hex())

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

	return e.prepareSettlementOperation("proposeSettlement", settlementHash, []transactTypes.Transaction{*holdTransaction}, settlement)
}

// HashSettlement computes the hash of a DvP settlement.
func HashSettlement(settlement *events.Settlement) (common.Hash, error) {
	uint256Ty, _ := abi.NewType("uint256", "", nil)
	uint64Ty, _ := abi.NewType("uint64", "", nil)
	uint48Ty, _ := abi.NewType("uint48", "", nil)
	uint8Ty, _ := abi.NewType("uint8", "", nil)
	addressTy, _ := abi.NewType("address", "", nil)
	bytes32Ty, _ := abi.NewType("bytes32", "", nil)
	bytesTy, _ := abi.NewType("bytes", "", nil)

	partyInfoArgs := abi.Arguments{
		{Type: addressTy}, {Type: addressTy}, {Type: addressTy}, {Type: addressTy}, {Type: addressTy},
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
		{Type: uint256Ty}, {Type: uint256Ty}, {Type: addressTy}, {Type: addressTy},
		{Type: addressTy}, {Type: addressTy}, {Type: uint8Ty}, {Type: uint8Ty},
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
		{Type: uint64Ty}, {Type: uint64Ty}, {Type: uint64Ty}, {Type: uint64Ty},
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
		{Type: bytes32Ty}, {Type: uint256Ty}, {Type: bytes32Ty}, {Type: bytes32Ty},
		{Type: bytes32Ty}, {Type: bytes32Ty}, {Type: uint48Ty}, {Type: bytesTy},
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

func (e *Extension) prepareSettlementOperation(
	operationName string,
	settlementHash common.Hash,
	extraTransactions []transactTypes.Transaction,
	packArgs ...interface{},
) (*transactTypes.Operation, error) {
	calldata, err := CCIPDVPCoordinatorUABI().Pack(operationName, packArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack calldata for %s: %w", operationName, err)
	}

	mainTransaction := transactTypes.Transaction{
		To:    e.ccipdvpCoordinatorUAddress,
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

	calldata, err := erc20Abi.Pack("approve", e.ccipdvpCoordinatorUAddress, tokenAmount)
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
		Recipient: e.ccipdvpCoordinatorUAddress,
		Executor:  e.ccipdvpCoordinatorUAddress,
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
