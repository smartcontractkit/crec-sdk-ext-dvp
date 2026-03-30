package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"

	gethAbi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"
	"github.com/smartcontractkit/cre-sdk-go/capabilities/blockchain/evm"
	evmmock "github.com/smartcontractkit/cre-sdk-go/capabilities/blockchain/evm/mock"
	httpcap "github.com/smartcontractkit/cre-sdk-go/capabilities/networking/http"
	httpmock "github.com/smartcontractkit/cre-sdk-go/capabilities/networking/http/mock"
	"github.com/smartcontractkit/cre-sdk-go/cre/testutils"

	wfcommon "github.com/smartcontractkit/crec-workflow-utils"
	wf "github.com/smartcontractkit/crec-sdk-ext-dvp/watcher/handler"
)

func ptr(s string) *string { return &s }

func Test_DVPEvent_HTTP_Post_WithCREReportSigs(t *testing.T) {
	rt := testutils.NewRuntime(t, testutils.Secrets{})

	cfg := &wfcommon.Config{
		ChainID:         "1337",
		ChainSelector:   "123456",
		CourierURL:      "http://example.com",
		Service:         ptr("dvp"),
		ConfidenceLevel: "finalized",
		DetectEventTriggerConfig: wfcommon.DetectEventTriggerConfig{
			ContractName:       "CCIPDVPCoordinator",
			ContractEventNames: []string{"SettlementAccepted"},
			ContractAddress:    "0xDVP",
			ContractReaderConfig: wfcommon.ContractReaderConfig{
				Contracts: map[string]wfcommon.ContractDef{
					"CCIPDVPCoordinator": {
						ContractABI: `[{"type":"event","name":"SettlementAccepted","inputs":[{"name":"settlementId","type":"uint256","indexed":true,"internalType":"uint256"},{"name":"settlementHash","type":"bytes32","indexed":true,"internalType":"bytes32"}],"anonymous":false}]`,
					},
				},
			},
		},
	}

	chainSelector, err := strconv.ParseUint(cfg.ChainSelector, 10, 64)
	require.NoError(t, err)
	evmCap, err := evmmock.NewClientCapability(chainSelector, t)
	require.NoError(t, err)

	evmCap.HeaderByNumber = func(_ context.Context, _ *evm.HeaderByNumberRequest) (*evm.HeaderByNumberReply, error) {
		return &evm.HeaderByNumberReply{Header: &evm.Header{Timestamp: 1000}}, nil
	}

	evmCap.CallContract = func(_ context.Context, req *evm.CallContractRequest) (*evm.CallContractReply, error) {
		ab, err := gethAbi.JSON(strings.NewReader(wf.DVPGetSettlementABI))
		if err != nil {
			return nil, fmt.Errorf("failed to parse ABI: %w", err)
		}
		getSettMethod, ok := ab.Methods["getSettlement"]
		if !ok {
			return nil, fmt.Errorf("getSettlement method not found in ABI")
		}

		type Settlement struct {
			CcipCallbackGasLimit *big.Int `abi:"CcipCallbackGasLimit"`
			Data                 []byte   `abi:"Data"`
			DeliveryInfo         struct {
				AssetDestinationChainSelector   uint64 `abi:"AssetDestinationChainSelector"`
				AssetSourceChainSelector        uint64 `abi:"AssetSourceChainSelector"`
				PaymentDestinationChainSelector uint64 `abi:"PaymentDestinationChainSelector"`
				PaymentSourceChainSelector      uint64 `abi:"PaymentSourceChainSelector"`
			} `abi:"DeliveryInfo"`
			ExecuteAfter *big.Int `abi:"ExecuteAfter"`
			Expiration   *big.Int `abi:"Expiration"`
			PartyInfo    struct {
				BuyerDestinationAddress  []byte `abi:"BuyerDestinationAddress"`
				BuyerSourceAddress       []byte `abi:"BuyerSourceAddress"`
				ExecutorAddress          []byte `abi:"ExecutorAddress"`
				SellerDestinationAddress []byte `abi:"SellerDestinationAddress"`
				SellerSourceAddress      []byte `abi:"SellerSourceAddress"`
			} `abi:"PartyInfo"`
			SecretHash   [32]byte `abi:"SecretHash"`
			SettlementId *big.Int `abi:"SettlementId"`
		TokenInfo    struct {
			AssetLockType                  uint8    `abi:"AssetLockType"`
			AssetTokenAmount               *big.Int `abi:"AssetTokenAmount"`
			AssetTokenDestinationAddress   []byte   `abi:"AssetTokenDestinationAddress"`
			AssetTokenSourceAddress        []byte   `abi:"AssetTokenSourceAddress"`
			PaymentCurrency                uint8    `abi:"PaymentCurrency"`
			PaymentLockType                uint8    `abi:"PaymentLockType"`
			PaymentTokenAmount             *big.Int `abi:"PaymentTokenAmount"`
			PaymentTokenDestinationAddress []byte   `abi:"PaymentTokenDestinationAddress"`
			PaymentTokenSourceAddress      []byte   `abi:"PaymentTokenSourceAddress"`
		} `abi:"TokenInfo"`
		}

		secretHashBytes := common.HexToHash("0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdef").Bytes()
		var secretHash [32]byte
		copy(secretHash[:], secretHashBytes)

		settlement := Settlement{
			CcipCallbackGasLimit: big.NewInt(50000),
			Data:                 []byte{0x01, 0x02, 0x03},
			DeliveryInfo: struct {
				AssetDestinationChainSelector   uint64 `abi:"AssetDestinationChainSelector"`
				AssetSourceChainSelector        uint64 `abi:"AssetSourceChainSelector"`
				PaymentDestinationChainSelector uint64 `abi:"PaymentDestinationChainSelector"`
				PaymentSourceChainSelector      uint64 `abi:"PaymentSourceChainSelector"`
			}{
				AssetDestinationChainSelector:   1,
				AssetSourceChainSelector:        2,
				PaymentDestinationChainSelector: 3,
				PaymentSourceChainSelector:      4,
			},
			ExecuteAfter: big.NewInt(1000),
			Expiration:   big.NewInt(2000),
			PartyInfo: struct {
				BuyerDestinationAddress  []byte `abi:"BuyerDestinationAddress"`
				BuyerSourceAddress       []byte `abi:"BuyerSourceAddress"`
				ExecutorAddress          []byte `abi:"ExecutorAddress"`
				SellerDestinationAddress []byte `abi:"SellerDestinationAddress"`
				SellerSourceAddress      []byte `abi:"SellerSourceAddress"`
			}{
				BuyerDestinationAddress:  common.HexToAddress("0x5555555555555555555555555555555555555555").Bytes(),
				BuyerSourceAddress:       common.HexToAddress("0x6666666666666666666666666666666666666666").Bytes(),
				ExecutorAddress:          common.HexToAddress("0x9999999999999999999999999999999999999999").Bytes(),
				SellerDestinationAddress: common.HexToAddress("0x7777777777777777777777777777777777777777").Bytes(),
				SellerSourceAddress:      common.HexToAddress("0x8888888888888888888888888888888888888888").Bytes(),
			},
			SecretHash:   secretHash,
			SettlementId: big.NewInt(123),
		TokenInfo: struct {
			AssetLockType                  uint8    `abi:"AssetLockType"`
			AssetTokenAmount               *big.Int `abi:"AssetTokenAmount"`
			AssetTokenDestinationAddress   []byte   `abi:"AssetTokenDestinationAddress"`
			AssetTokenSourceAddress        []byte   `abi:"AssetTokenSourceAddress"`
			PaymentCurrency                uint8    `abi:"PaymentCurrency"`
			PaymentLockType                uint8    `abi:"PaymentLockType"`
			PaymentTokenAmount             *big.Int `abi:"PaymentTokenAmount"`
			PaymentTokenDestinationAddress []byte   `abi:"PaymentTokenDestinationAddress"`
			PaymentTokenSourceAddress      []byte   `abi:"PaymentTokenSourceAddress"`
		}{
			AssetLockType:                  1,
			AssetTokenAmount:               big.NewInt(100),
			AssetTokenDestinationAddress:   common.HexToAddress("0x2222222222222222222222222222222222222222").Bytes(),
			AssetTokenSourceAddress:        common.HexToAddress("0x1111111111111111111111111111111111111111").Bytes(),
			PaymentCurrency:                0,
			PaymentLockType:                2,
			PaymentTokenAmount:             big.NewInt(200),
			PaymentTokenDestinationAddress: common.HexToAddress("0x4444444444444444444444444444444444444444").Bytes(),
			PaymentTokenSourceAddress:      common.HexToAddress("0x3333333333333333333333333333333333333333").Bytes(),
		},
		}

		packed, err := getSettMethod.Outputs.Pack(settlement)
		if err != nil {
			return nil, fmt.Errorf("failed to pack settlement data: %w", err)
		}
		return &evm.CallContractReply{Data: packed}, nil
	}

	settlementHashHex := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	httpCap, err := httpmock.NewClientCapability(t)
	require.NoError(t, err)

	httpCap.SendRequest = func(_ context.Context, req *httpcap.Request) (*httpcap.Response, error) {
		if req.Method != "POST" {
			return nil, fmt.Errorf("expected POST, got %s", req.Method)
		}
		expectedURL := "http://example.com/system/v1/onchain-watcher-events"
		if req.Url != expectedURL {
			return nil, fmt.Errorf("unexpected url %q (expected %q)", req.Url, expectedURL)
		}

		var body map[string]any
		if err := json.Unmarshal(req.Body, &body); err != nil {
			return nil, fmt.Errorf("failed to unmarshal request body: %w", err)
		}
		if body["verifiable_event"] == nil || body["verifiable_event"] == "" {
			return nil, fmt.Errorf("verifiable_event field should be present")
		}
		if body["ocr_report"] == nil || body["ocr_report"] == "" {
			return nil, fmt.Errorf("ocr_report field should be present")
		}
		if body["ocr_context"] == nil || body["ocr_context"] == "" {
			return nil, fmt.Errorf("ocr_context field should be present")
		}
		if body["signatures"] == nil {
			return nil, fmt.Errorf("signatures field should be present")
		}

		verStr, ok := body["verifiable_event"].(string)
		if !ok {
			return nil, fmt.Errorf("verifiable_event should be a string")
		}
		verBytes, err := base64.StdEncoding.DecodeString(verStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64: %w", err)
		}

		var ve map[string]any
		if err := json.Unmarshal(verBytes, &ve); err != nil {
			return nil, fmt.Errorf("failed to unmarshal verifiable event: %w", err)
		}

		if ve["chain_event"] == nil {
			return nil, fmt.Errorf("chain_event should be present")
		}
		if ve["data"] == nil {
			return nil, fmt.Errorf("data field should be present")
		}

		data, ok := ve["data"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("data should be a map")
		}

		if data["settlement_id"] != "1" {
			return nil, fmt.Errorf("expected settlement_id '1', got %v", data["settlement_id"])
		}
		if data["settlement_hash"] != settlementHashHex {
			return nil, fmt.Errorf("expected settlement_hash %q, got %v", settlementHashHex, data["settlement_hash"])
		}

		if data["dvp"] == nil {
			return nil, fmt.Errorf("data.dvp should be present")
		}

		dvpData, ok := data["dvp"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("data.dvp should be a map")
		}

		toInt64 := func(v any) (int64, error) {
			switch t := v.(type) {
			case int64:
				return t, nil
			case float64:
				return int64(t), nil
			case int:
				return int64(t), nil
			case int32:
				return int64(t), nil
			default:
				return 0, fmt.Errorf("cannot convert %T to int64", v)
			}
		}

		ccipVal, err := toInt64(dvpData["ccip_callback_gas_limit"])
		if err != nil || ccipVal != 50000 {
			return nil, fmt.Errorf("expected ccip_callback_gas_limit 50000, got %v (%v)", dvpData["ccip_callback_gas_limit"], err)
		}
		executeAfter, err := toInt64(dvpData["execute_after"])
		if err != nil || executeAfter != 1000 {
			return nil, fmt.Errorf("expected execute_after 1000, got %v (%v)", dvpData["execute_after"], err)
		}
		expiration, err := toInt64(dvpData["expiration"])
		if err != nil || expiration != 2000 {
			return nil, fmt.Errorf("expected expiration 2000, got %v (%v)", dvpData["expiration"], err)
		}

		tokenInfo, ok := dvpData["token_info"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("token_info should be a map")
		}
		assetTokenAmount, err := toInt64(tokenInfo["asset_token_amount"])
		if err != nil || assetTokenAmount != 100 {
			return nil, fmt.Errorf("expected asset_token_amount 100, got %v (%v)", tokenInfo["asset_token_amount"], err)
		}
		paymentTokenAmount, err := toInt64(tokenInfo["payment_token_amount"])
		if err != nil || paymentTokenAmount != 200 {
			return nil, fmt.Errorf("expected payment_token_amount 200, got %v (%v)", tokenInfo["payment_token_amount"], err)
		}
		assetLockType, err := toInt64(tokenInfo["asset_lock_type"])
		if err != nil || assetLockType != 1 {
			return nil, fmt.Errorf("expected asset_lock_type 1, got %v (%v)", tokenInfo["asset_lock_type"], err)
		}
		paymentLockType, err := toInt64(tokenInfo["payment_lock_type"])
		if err != nil || paymentLockType != 2 {
			return nil, fmt.Errorf("expected payment_lock_type 2, got %v (%v)", tokenInfo["payment_lock_type"], err)
		}
		paymentCurrency, err := toInt64(tokenInfo["payment_currency"])
		if err != nil || paymentCurrency != 0 {
			return nil, fmt.Errorf("expected payment_currency 0, got %v (%v)", tokenInfo["payment_currency"], err)
		}

		return &httpcap.Response{StatusCode: 201}, nil
	}

	eventSig := crypto.Keccak256Hash([]byte("SettlementAccepted(uint256,bytes32)")).Bytes()
	topics := [][]byte{
		eventSig,
		common.BigToHash(big.NewInt(1)).Bytes(),
		common.HexToHash(settlementHashHex).Bytes(),
	}

	log := &evm.Log{
		Address:     common.HexToHash("0xDVP").Bytes(),
		Topics:      topics,
		BlockNumber: &pb.BigInt{AbsVal: []byte{1}},
		TxHash:      common.HexToHash("0xTx").Bytes(),
		Index:       0,
	}

	_, err = wf.OnLog(cfg, rt, log)
	require.NoError(t, err)
}
