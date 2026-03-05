package dvp_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/crec-sdk-ext-dvp/events"
	"github.com/smartcontractkit/crec-sdk-ext-dvp/operations"
)

func TestOperations_New(t *testing.T) {
	tests := []struct {
		name    string
		opts    *operations.Options
		wantErr bool
	}{
		{
			name:    "nil options returns error",
			opts:    nil,
			wantErr: true,
		},
		{
			name: "valid options creates extension",
			opts: &operations.Options{
				CCIPDVPCoordinatorUAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
				AccountAddress:             "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, err := operations.New(tt.opts)
			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, ext)
			} else {
				require.NoError(t, err)
				require.NotNil(t, ext)
			}
		})
	}
}

func TestOperations_HashSettlement(t *testing.T) {
	ext, err := operations.New(&operations.Options{
		CCIPDVPCoordinatorUAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:             "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
	})
	require.NoError(t, err)

	settlement := &events.Settlement{
		SettlementId: big.NewInt(1751404299),
		PartyInfo: events.PartyInfo{
			BuyerSourceAddress:       common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621"),
			BuyerDestinationAddress:  common.HexToAddress("0xeb457346d2218f7f77aa23ac6d9e394b505dd621"),
			SellerSourceAddress:      common.HexToAddress("0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc"),
			SellerDestinationAddress: common.HexToAddress("0xce2152bfcd0995f56a07dcbfef2bc85d404d65bc"),
			ExecutorAddress:          common.HexToAddress("0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1"),
		},
		TokenInfo: events.TokenInfo{
			PaymentTokenSourceAddress:      common.HexToAddress("0x0000000000000000000000000000000000000000"),
			PaymentTokenDestinationAddress: common.HexToAddress("0x0000000000000000000000000000000000000000"),
			AssetTokenSourceAddress:        common.HexToAddress("0xA5F12FDA3e8B7209a3019141F105e5DB43445B86"),
			AssetTokenDestinationAddress:   common.HexToAddress("0xA5F12FDA3e8B7209a3019141F105e5DB43445B86"),
			PaymentCurrency:                events.CurrencyMap["USD"],
			PaymentTokenAmount:             big.NewInt(1000000),
			AssetTokenAmount:               big.NewInt(1000000000000000000),
			PaymentTokenType:               events.TokenTypeNone,
			AssetTokenType:                 events.TokenTypeERC20,
		},
		DeliveryInfo: events.DeliveryInfo{
			PaymentSourceChainSelector:      uint64(1234567890),
			PaymentDestinationChainSelector: uint64(1234567890),
			AssetSourceChainSelector:        uint64(1234567890),
			AssetDestinationChainSelector:   uint64(1234567890),
		},
		SecretHash:           common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000000"),
		ExecuteAfter:         big.NewInt(0),
		Expiration:           big.NewInt(1751490699),
		CcipCallbackGasLimit: 0,
		Data:                 []byte{},
	}

	hash, err := ext.HashSettlement(settlement)
	require.NoError(t, err)
	require.Equal(t, common.HexToHash("0xc36535b1628c991180c156e097d0fa8062c2d1bce2d7bfca8aefe88034005eec"), hash)
}

func TestOperations_PrepareProposeSettlementOperation(t *testing.T) {
	ext, err := operations.New(&operations.Options{
		CCIPDVPCoordinatorUAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:             "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
	})
	require.NoError(t, err)

	settlement := &events.Settlement{
		SettlementId: big.NewInt(1),
		PartyInfo: events.PartyInfo{
			BuyerSourceAddress:       common.HexToAddress("0x1111111111111111111111111111111111111111"),
			BuyerDestinationAddress:  common.HexToAddress("0x1111111111111111111111111111111111111111"),
			SellerSourceAddress:      common.HexToAddress("0x2222222222222222222222222222222222222222"),
			SellerDestinationAddress: common.HexToAddress("0x2222222222222222222222222222222222222222"),
			ExecutorAddress:          common.HexToAddress("0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1"),
		},
		TokenInfo: events.TokenInfo{
			PaymentTokenSourceAddress:      common.HexToAddress("0x0000000000000000000000000000000000000000"),
			PaymentTokenDestinationAddress: common.HexToAddress("0x0000000000000000000000000000000000000000"),
			AssetTokenSourceAddress:        common.HexToAddress("0x3333333333333333333333333333333333333333"),
			AssetTokenDestinationAddress:   common.HexToAddress("0x3333333333333333333333333333333333333333"),
			PaymentCurrency:                events.CurrencyMap["USD"],
			PaymentTokenAmount:             big.NewInt(100),
			AssetTokenAmount:               big.NewInt(1000),
			PaymentTokenType:               events.TokenTypeNone,
			AssetTokenType:                 events.TokenTypeERC20,
		},
		DeliveryInfo: events.DeliveryInfo{
			PaymentSourceChainSelector:      uint64(1),
			PaymentDestinationChainSelector: uint64(1),
			AssetSourceChainSelector:        uint64(1),
			AssetDestinationChainSelector:   uint64(1),
		},
		SecretHash:           common.Hash{},
		ExecuteAfter:         big.NewInt(0),
		Expiration:           big.NewInt(9999999999),
		CcipCallbackGasLimit: 0,
		Data:                 []byte{},
	}

	op, err := ext.PrepareProposeSettlementOperation(settlement)
	require.NoError(t, err)
	require.NotNil(t, op)
	require.Len(t, op.Transactions, 1)
	require.Equal(t, common.HexToAddress("0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE"), op.Transactions[0].To)
}

func TestOperations_PrepareProposeSettlementWithTokenApprovalOperation(t *testing.T) {
	ext, err := operations.New(&operations.Options{
		CCIPDVPCoordinatorUAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:             "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
	})
	require.NoError(t, err)

	settlement := &events.Settlement{
		SettlementId: big.NewInt(1),
		PartyInfo: events.PartyInfo{
			BuyerSourceAddress:       common.HexToAddress("0x1111111111111111111111111111111111111111"),
			BuyerDestinationAddress:  common.HexToAddress("0x1111111111111111111111111111111111111111"),
			SellerSourceAddress:      common.HexToAddress("0x2222222222222222222222222222222222222222"),
			SellerDestinationAddress: common.HexToAddress("0x2222222222222222222222222222222222222222"),
			ExecutorAddress:          common.HexToAddress("0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1"),
		},
		TokenInfo: events.TokenInfo{
			PaymentTokenSourceAddress:      common.HexToAddress("0x0000000000000000000000000000000000000000"),
			PaymentTokenDestinationAddress: common.HexToAddress("0x0000000000000000000000000000000000000000"),
			AssetTokenSourceAddress:        common.HexToAddress("0x3333333333333333333333333333333333333333"),
			AssetTokenDestinationAddress:   common.HexToAddress("0x3333333333333333333333333333333333333333"),
			PaymentCurrency:                events.CurrencyMap["USD"],
			PaymentTokenAmount:             big.NewInt(100),
			AssetTokenAmount:               big.NewInt(1000),
			PaymentTokenType:               events.TokenTypeNone,
			AssetTokenType:                 events.TokenTypeERC20,
		},
		DeliveryInfo: events.DeliveryInfo{
			PaymentSourceChainSelector:      uint64(1),
			PaymentDestinationChainSelector: uint64(1),
			AssetSourceChainSelector:        uint64(1),
			AssetDestinationChainSelector:   uint64(1),
		},
		SecretHash:           common.Hash{},
		ExecuteAfter:         big.NewInt(0),
		Expiration:           big.NewInt(9999999999),
		CcipCallbackGasLimit: 0,
		Data:                 []byte{},
	}

	op, err := ext.PrepareProposeSettlementWithTokenApprovalOperation(settlement)
	require.NoError(t, err)
	require.NotNil(t, op)
	require.Len(t, op.Transactions, 2)
	require.Equal(t, common.HexToAddress("0x3333333333333333333333333333333333333333"), op.Transactions[0].To)
	require.Equal(t, common.HexToAddress("0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE"), op.Transactions[1].To)
}

func TestOperations_PrepareProposeSettlementWithTokenHoldOperation(t *testing.T) {
	holdManagerAddr := common.HexToAddress("0x4444444444444444444444444444444444444444")
	ext, err := operations.New(&operations.Options{
		CCIPDVPCoordinatorUAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:             "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
	})
	require.NoError(t, err)

	settlement := &events.Settlement{
		SettlementId: big.NewInt(1),
		PartyInfo: events.PartyInfo{
			BuyerSourceAddress:       common.HexToAddress("0x1111111111111111111111111111111111111111"),
			BuyerDestinationAddress:  common.HexToAddress("0x1111111111111111111111111111111111111111"),
			SellerSourceAddress:      common.HexToAddress("0x2222222222222222222222222222222222222222"),
			SellerDestinationAddress: common.HexToAddress("0x2222222222222222222222222222222222222222"),
			ExecutorAddress:          common.HexToAddress("0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1"),
		},
		TokenInfo: events.TokenInfo{
			PaymentTokenSourceAddress:      common.HexToAddress("0x0000000000000000000000000000000000000000"),
			PaymentTokenDestinationAddress: common.HexToAddress("0x0000000000000000000000000000000000000000"),
			AssetTokenSourceAddress:        common.HexToAddress("0x3333333333333333333333333333333333333333"),
			AssetTokenDestinationAddress:   common.HexToAddress("0x3333333333333333333333333333333333333333"),
			PaymentCurrency:                events.CurrencyMap["USD"],
			PaymentTokenAmount:             big.NewInt(100),
			AssetTokenAmount:               big.NewInt(1000),
			PaymentTokenType:               events.TokenTypeNone,
			AssetTokenType:                 events.TokenTypeERC3643,
		},
		DeliveryInfo: events.DeliveryInfo{
			PaymentSourceChainSelector:      uint64(1),
			PaymentDestinationChainSelector: uint64(1),
			AssetSourceChainSelector:        uint64(1),
			AssetDestinationChainSelector:   uint64(1),
		},
		SecretHash:           common.Hash{},
		ExecuteAfter:         big.NewInt(0),
		Expiration:           big.NewInt(9999999999),
		CcipCallbackGasLimit: 0,
		Data:                 []byte{},
	}

	op, err := ext.PrepareProposeSettlementWithTokenHoldOperation(settlement, holdManagerAddr)
	require.NoError(t, err)
	require.NotNil(t, op)
	require.Len(t, op.Transactions, 2)
	require.Equal(t, holdManagerAddr, op.Transactions[0].To)
	require.Equal(t, common.HexToAddress("0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE"), op.Transactions[1].To)
}

func TestOperations_PrepareProposeSettlementWithTokenHoldOperation_NonERC3643Fails(t *testing.T) {
	holdManagerAddr := common.HexToAddress("0x4444444444444444444444444444444444444444")
	ext, err := operations.New(&operations.Options{
		CCIPDVPCoordinatorUAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:             "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
	})
	require.NoError(t, err)

	settlement := &events.Settlement{
		SettlementId: big.NewInt(1),
		PartyInfo: events.PartyInfo{
			BuyerSourceAddress:       common.HexToAddress("0x1111111111111111111111111111111111111111"),
			BuyerDestinationAddress:  common.HexToAddress("0x1111111111111111111111111111111111111111"),
			SellerSourceAddress:      common.HexToAddress("0x2222222222222222222222222222222222222222"),
			SellerDestinationAddress: common.HexToAddress("0x2222222222222222222222222222222222222222"),
			ExecutorAddress:          common.HexToAddress("0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1"),
		},
		TokenInfo: events.TokenInfo{
			PaymentTokenSourceAddress:      common.HexToAddress("0x0000000000000000000000000000000000000000"),
			PaymentTokenDestinationAddress: common.HexToAddress("0x0000000000000000000000000000000000000000"),
			AssetTokenSourceAddress:        common.HexToAddress("0x3333333333333333333333333333333333333333"),
			AssetTokenDestinationAddress:   common.HexToAddress("0x3333333333333333333333333333333333333333"),
			PaymentCurrency:                events.CurrencyMap["USD"],
			PaymentTokenAmount:             big.NewInt(100),
			AssetTokenAmount:               big.NewInt(1000),
			PaymentTokenType:               events.TokenTypeNone,
			AssetTokenType:                 events.TokenTypeERC20,
		},
		DeliveryInfo: events.DeliveryInfo{
			PaymentSourceChainSelector:      uint64(1),
			PaymentDestinationChainSelector: uint64(1),
			AssetSourceChainSelector:        uint64(1),
			AssetDestinationChainSelector:   uint64(1),
		},
		SecretHash:           common.Hash{},
		ExecuteAfter:         big.NewInt(0),
		Expiration:           big.NewInt(9999999999),
		CcipCallbackGasLimit: 0,
		Data:                 []byte{},
	}

	op, err := ext.PrepareProposeSettlementWithTokenHoldOperation(settlement, holdManagerAddr)
	require.Error(t, err)
	require.Nil(t, op)
	require.Contains(t, err.Error(), "ERC3643")
}

func TestOperations_PrepareCancelOperation(t *testing.T) {
	ext, err := operations.New(&operations.Options{
		CCIPDVPCoordinatorUAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:             "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
	})
	require.NoError(t, err)

	settlementHash := common.HexToHash("0xc36535b1628c991180c156e097d0fa8062c2d1bce2d7bfca8aefe88034005eec")

	op, err := ext.PrepareCancelOperation([32]byte(settlementHash))
	require.NoError(t, err)
	require.NotNil(t, op)
	require.Len(t, op.Transactions, 1)
	require.Equal(t, common.HexToAddress("0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE"), op.Transactions[0].To)
}

func TestOperations_PrepareExecuteSettlementWithTokenDataOperation(t *testing.T) {
	ext, err := operations.New(&operations.Options{
		CCIPDVPCoordinatorUAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:             "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
	})
	require.NoError(t, err)

	settlementHash := common.HexToHash("0xc36535b1628c991180c156e097d0fa8062c2d1bce2d7bfca8aefe88034005eec")
	tokenData := []byte{0x01, 0x02, 0x03}

	op, err := ext.PrepareExecuteSettlementWithTokenDataOperation([32]byte(settlementHash), tokenData)
	require.NoError(t, err)
	require.NotNil(t, op)
	require.Len(t, op.Transactions, 1)
	require.Equal(t, common.HexToAddress("0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE"), op.Transactions[0].To)
}

func TestOperations_PrepareAcceptSettlementOperation(t *testing.T) {
	ext, err := operations.New(&operations.Options{
		CCIPDVPCoordinatorUAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:             "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
	})
	require.NoError(t, err)

	settlementHash := common.HexToHash("0xc36535b1628c991180c156e097d0fa8062c2d1bce2d7bfca8aefe88034005eec")

	op, err := ext.PrepareAcceptSettlementOperation([32]byte(settlementHash))
	require.NoError(t, err)
	require.NotNil(t, op)
	require.Len(t, op.Transactions, 1)
	require.Equal(t, common.HexToAddress("0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE"), op.Transactions[0].To)
}

func TestOperations_PrepareExecuteSettlementOperation(t *testing.T) {
	ext, err := operations.New(&operations.Options{
		CCIPDVPCoordinatorUAddress: "0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE",
		AccountAddress:             "0x7Eb6D2Bf84C394A1718a60f0f89FBc4626eCdbA1",
	})
	require.NoError(t, err)

	settlementHash := common.HexToHash("0xc36535b1628c991180c156e097d0fa8062c2d1bce2d7bfca8aefe88034005eec")

	op, err := ext.PrepareExecuteSettlementOperation([32]byte(settlementHash))
	require.NoError(t, err)
	require.NotNil(t, op)
	require.Len(t, op.Transactions, 1)
	require.Equal(t, common.HexToAddress("0x9A9f2CCfdE556A7E9Ff0848998Aa4a0CFD8863AE"), op.Transactions[0].To)
}

func TestEvents_CurrencyMap(t *testing.T) {
	tests := []struct {
		code     string
		expected uint8
	}{
		{"USD", 147},
		{"EUR", 48},
		{"GBP", 51},
		{"None", 0},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			require.Equal(t, tt.expected, events.CurrencyMap[tt.code])
		})
	}
}
