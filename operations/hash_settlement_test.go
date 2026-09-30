package operations

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/crec-sdk-ext-dvp/events"
)

// goldenHashFull and goldenHashAlt were produced by the deployed
// CCIPDVPCoordinator's getSettlementHash view function (eth-sepolia,
// 0xfAe9C0dD04c006003bff626E22880eB3413b92f8) for the settlements built below.
// HashSettlement must reproduce the contract's SettlementLib._hash exactly:
// watcher events carry the on-chain hash, and execute/cancel/accept take it as
// an argument, so any drift breaks the settlement lifecycle.
const (
	goldenHashFull = "0x3d8eb58f908fa83dfc02295916e2acaef043b29f9bf2494c7e9c94a4f09328a2"
	goldenHashAlt  = "0x91c8296d05eca196fa44a72e22f16f30e11c20ed37823decf138c659f06cc6c8"
)

func TestHashSettlementMatchesContract(t *testing.T) {
	tests := []struct {
		name       string
		want       string
		settlement *events.Settlement
	}{
		{
			name: "full",
			want: goldenHashFull,
			settlement: &events.Settlement{
				SettlementId: big.NewInt(123456789),
				PartyInfo: events.PartyInfo{
					BuyerSourceAddress:       common.HexToAddress("0x1111111111111111111111111111111111111111"),
					BuyerDestinationAddress:  common.HexToAddress("0x2222222222222222222222222222222222222222"),
					SellerSourceAddress:      common.HexToAddress("0x3333333333333333333333333333333333333333"),
					SellerDestinationAddress: common.HexToAddress("0x4444444444444444444444444444444444444444"),
					ExecutorAddress:          common.HexToAddress("0x5555555555555555555555555555555555555555"),
				},
				TokenInfo: events.TokenInfo{
					PaymentTokenSourceAddress:      common.HexToAddress("0x6666666666666666666666666666666666666666"),
					PaymentTokenDestinationAddress: common.HexToAddress("0x7777777777777777777777777777777777777777"),
					AssetTokenSourceAddress:        common.HexToAddress("0x8888888888888888888888888888888888888888"),
					AssetTokenDestinationAddress:   common.HexToAddress("0x9999999999999999999999999999999999999999"),
					PaymentTokenAmount:             big.NewInt(1_000_000),
					AssetTokenAmount:               big.NewInt(2_000_000_000),
					PaymentCurrency:                147,
					PaymentLockType:                3,
					AssetLockType:                  1,
				},
				DeliveryInfo: events.DeliveryInfo{
					PaymentSourceChainSelector:      16015286601757825753,
					PaymentDestinationChainSelector: 3478487238524512106,
					AssetSourceChainSelector:        16015286601757825753,
					AssetDestinationChainSelector:   3478487238524512106,
				},
				SecretHash:           common.HexToHash("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa01"),
				ExecuteAfter:         big.NewInt(1_700_000_000),
				Expiration:           big.NewInt(1_800_000_000),
				CcipCallbackGasLimit: 300_000,
				Data:                 []byte("hash-parity-vector"),
			},
		},
		{
			name: "zero-values-empty-data",
			want: goldenHashAlt,
			settlement: &events.Settlement{
				SettlementId: new(big.Int).SetUint64(9876543210987654321),
				PartyInfo: events.PartyInfo{
					BuyerSourceAddress:       common.HexToAddress("0x1000000000000000000000000000000000000001"),
					BuyerDestinationAddress:  common.HexToAddress("0x1000000000000000000000000000000000000002"),
					SellerSourceAddress:      common.HexToAddress("0x1000000000000000000000000000000000000003"),
					SellerDestinationAddress: common.HexToAddress("0x1000000000000000000000000000000000000004"),
					ExecutorAddress:          common.Address{},
				},
				TokenInfo: events.TokenInfo{
					PaymentTokenAmount: big.NewInt(1),
					AssetTokenAmount:   big.NewInt(1),
					PaymentCurrency:    0,
					PaymentLockType:    0,
					AssetLockType:      0,
				},
				DeliveryInfo: events.DeliveryInfo{
					PaymentDestinationChainSelector: 5009297550715157269,
				},
				ExecuteAfter: big.NewInt(0),
				Expiration:   big.NewInt(2_000_000_000),
				Data:         []byte{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HashSettlement(tt.settlement)
			require.NoError(t, err)
			require.Equal(t, tt.want, got.Hex(), "HashSettlement must match the contract's getSettlementHash")
		})
	}
}

func TestHashSettlementRejectsMissingFields(t *testing.T) {
	base := func() *events.Settlement {
		return &events.Settlement{
			SettlementId: big.NewInt(1),
			TokenInfo: events.TokenInfo{
				PaymentTokenAmount: big.NewInt(1),
				AssetTokenAmount:   big.NewInt(1),
			},
			ExecuteAfter: big.NewInt(0),
			Expiration:   big.NewInt(1),
		}
	}

	tests := []struct {
		name       string
		settlement *events.Settlement
	}{
		{name: "nil settlement", settlement: nil},
		{name: "nil settlement id", settlement: func() *events.Settlement {
			s := base()
			s.SettlementId = nil
			return s
		}()},
		{name: "nil payment token amount", settlement: func() *events.Settlement {
			s := base()
			s.TokenInfo.PaymentTokenAmount = nil
			return s
		}()},
		{name: "nil asset token amount", settlement: func() *events.Settlement {
			s := base()
			s.TokenInfo.AssetTokenAmount = nil
			return s
		}()},
		{name: "nil execute after", settlement: func() *events.Settlement {
			s := base()
			s.ExecuteAfter = nil
			return s
		}()},
		{name: "nil expiration", settlement: func() *events.Settlement {
			s := base()
			s.Expiration = nil
			return s
		}()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := HashSettlement(tt.settlement)
			require.Error(t, err)
		})
	}
}
