package handler

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	gethAbi "github.com/ethereum/go-ethereum/accounts/abi"
	gethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/smartcontractkit/cre-sdk-go/capabilities/blockchain/evm"
	"github.com/smartcontractkit/cre-sdk-go/cre"

	wfcommon "github.com/smartcontractkit/cre-workflow-utils"
)

// Minimal ABI for getSettlement(bytes32) returning the Settlement struct.
// This mirrors the Solidity struct layout used in DVPCoordinator.
const DVPGetSettlementABI = `[
  {
    "type":"function",
    "name":"getSettlement",
    "stateMutability":"view",
    "inputs":[{"name":"settlementHash","type":"bytes32"}],
    "outputs":[{"name":"settlement","type":"tuple","components":[
      {"name":"CcipCallbackGasLimit","type":"uint256"},
      {"name":"Data","type":"bytes"},
      {"name":"DeliveryInfo","type":"tuple","components":[
        {"name":"AssetDestinationChainSelector","type":"uint64"},
        {"name":"AssetSourceChainSelector","type":"uint64"},
        {"name":"PaymentDestinationChainSelector","type":"uint64"},
        {"name":"PaymentSourceChainSelector","type":"uint64"}
      ]},
      {"name":"ExecuteAfter","type":"uint256"},
      {"name":"Expiration","type":"uint256"},
      {"name":"PartyInfo","type":"tuple","components":[
        {"name":"BuyerDestinationAddress","type":"bytes"},
        {"name":"BuyerSourceAddress","type":"bytes"},
        {"name":"ExecutorAddress","type":"bytes"},
        {"name":"SellerDestinationAddress","type":"bytes"},
        {"name":"SellerSourceAddress","type":"bytes"}
      ]},
      {"name":"SecretHash","type":"bytes32"},
      {"name":"SettlementId","type":"uint256"},
      {"name":"TokenInfo","type":"tuple","components":[
        {"name":"AssetTokenAmount","type":"uint256"},
        {"name":"AssetTokenDestinationAddress","type":"bytes"},
        {"name":"AssetTokenSourceAddress","type":"bytes"},
        {"name":"AssetTokenType","type":"uint8"},
        {"name":"PaymentCurrency","type":"uint8"},
        {"name":"PaymentTokenAmount","type":"uint256"},
        {"name":"PaymentTokenDestinationAddress","type":"bytes"},
        {"name":"PaymentTokenSourceAddress","type":"bytes"},
        {"name":"PaymentTokenType","type":"uint8"}
      ]}
    ]}]
  }
]`

// OnLog processes DVP settlement logs:
// - Performs an EVM view call getSettlement(bytes32) and decodes the returned Settlement for metadata
// - Composes a verifiable event, signs it, and posts it to Courier
func OnLog(cfg *wfcommon.Config, rt cre.Runtime, payload *evm.Log) (string, error) {
	selector := cfg.ChainSelector

	if len(payload.Topics) < 3 {
		return "", fmt.Errorf("log topics length < 3")
	}
	settlementID := new(big.Int).SetBytes(payload.Topics[1])
	settlementHashBytes := payload.Topics[2]
	settlementHash := "0x" + fmt.Sprintf("%064x", new(big.Int).SetBytes(settlementHashBytes))

	dvpMeta := map[string]any{}
	if settlementDecoded, err := fetchAndDecodeSettlement(rt, selector, cfg.DetectEventTriggerConfig.ContractAddress, settlementHashBytes); err == nil && settlementDecoded != nil {
		dvpSanitised := wfcommon.SanitiseJSON(settlementDecoded).(map[string]any)
		dvpMeta = fixDVPTypes(dvpSanitised)
	}

	evmEvent, err := wfcommon.BuildEVMEventFromLog(rt, cfg, payload)
	if err != nil {
		return "", err
	}

	abiJSON, err := wfcommon.GetContractABI(cfg, cfg.DetectEventTriggerConfig.ContractName)
	if err != nil {
		return "", err
	}
	eventName, err := wfcommon.GetEventNameFromLog(cfg, payload, abiJSON)
	if err != nil {
		return "", err
	}

	metadata := map[string]any{
		"settlement_id":   settlementID.String(),
		"settlement_hash": settlementHash,
	}
	if len(dvpMeta) > 0 {
		metadata["dvp"] = dvpMeta
	}

	verifiableEvent, err := wfcommon.BuildVerifiableEventForEVMEvent(
		cfg,
		evmEvent,
		cfg.Service,
		eventName,
		&metadata,
	)
	if err != nil {
		return "", err
	}

	return wfcommon.SignAndPostVerifiableEvent(cfg, rt, verifiableEvent)
}

func fetchAndDecodeSettlement(rt cre.Runtime, selector string, contractAddr string, hash []byte) (map[string]any, error) {
	ab, err := gethAbi.JSON(strings.NewReader(DVPGetSettlementABI))
	if err != nil {
		return nil, err
	}
	getSettMethod, ok := ab.Methods["getSettlement"]
	if !ok {
		return nil, fmt.Errorf("getSettlement method not found")
	}

	to := gethCommon.HexToAddress(contractAddr)
	methodID := crypto.Keccak256([]byte("getSettlement(bytes32)"))[:4]
	callData := make([]byte, 4+32)
	copy(callData[:4], methodID)
	copy(callData[4+32-len(hash):], hash)

	chainSelector, err := strconv.ParseUint(selector, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid chain selector: %w", err)
	}
	cli := &evm.Client{ChainSelector: chainSelector}
	callContractReply, err := cli.CallContract(rt, &evm.CallContractRequest{
		Call: &evm.CallMsg{
			To:   to.Bytes(),
			Data: callData,
		},
	}).Await()
	if err != nil {
		return nil, err
	}
	if len(callContractReply.Data) == 0 {
		return nil, fmt.Errorf("empty getSettlement response")
	}

	decoded := map[string]any{}
	if err := getSettMethod.Outputs.UnpackIntoMap(decoded, callContractReply.Data); err != nil {
		return nil, err
	}

	raw, ok := decoded["settlement"]
	if !ok {
		return nil, fmt.Errorf("getSettlement output missing 'settlement'")
	}

	if mm, ok := raw.(map[string]any); ok {
		return mm, nil
	}
	var settlementMap map[string]any
	b, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settlement: %w", err)
	}
	if err := json.Unmarshal(b, &settlementMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settlement: %w", err)
	}
	if settlementMap == nil {
		return nil, fmt.Errorf("unexpected decoded type %T", raw)
	}
	return settlementMap, nil
}

func fixDVPTypes(m map[string]any) map[string]any {
	toInt := func(v any) any {
		switch t := v.(type) {
		case string:
			if bi, ok := new(big.Int).SetString(t, 10); ok {
				return bi.Int64()
			}
			return t
		case *big.Int:
			if t != nil && t.IsInt64() {
				return t.Int64()
			}
			return t.String()
		case float64:
			return int64(t)
		case int64, int32, int, uint64, uint32, uint:
			return t
		default:
			return v
		}
	}
	for _, k := range []string{"execute_after", "expiration", "ccip_callback_gas_limit"} {
		if v, ok := m[k]; ok {
			m[k] = toInt(v)
		}
	}
	if ti, ok := m["token_info"].(map[string]any); ok {
		for _, k := range []string{"payment_currency", "payment_token_type", "asset_token_type", "asset_token_amount", "payment_token_amount"} {
			if v, ok2 := ti[k]; ok2 {
				ti[k] = toInt(v)
			}
		}
		m["token_info"] = ti
	}
	return m
}
