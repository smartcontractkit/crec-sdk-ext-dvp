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

	wfcommon "github.com/smartcontractkit/crec-workflow-utils"
)

// Minimal ABI for getSettlement(bytes32) returning the Settlement struct.
// Field order and types MUST match the actual Solidity struct layout in
// CCIPDVPCoordinator — eth_call returns ABI-encoded data in struct order.
const DVPGetSettlementABI = `[
  {
    "type":"function",
    "name":"getSettlement",
    "stateMutability":"view",
    "inputs":[{"name":"settlementHash","type":"bytes32"}],
    "outputs":[{"name":"settlement","type":"tuple","components":[
      {"name":"settlementId","type":"uint256"},
      {"name":"partyInfo","type":"tuple","components":[
        {"name":"buyerSourceAddress","type":"address"},
        {"name":"buyerDestinationAddress","type":"address"},
        {"name":"sellerSourceAddress","type":"address"},
        {"name":"sellerDestinationAddress","type":"address"},
        {"name":"executorAddress","type":"address"}
      ]},
      {"name":"tokenInfo","type":"tuple","components":[
        {"name":"paymentTokenSourceAddress","type":"address"},
        {"name":"paymentTokenDestinationAddress","type":"address"},
        {"name":"assetTokenSourceAddress","type":"address"},
        {"name":"assetTokenDestinationAddress","type":"address"},
        {"name":"paymentTokenAmount","type":"uint256"},
        {"name":"assetTokenAmount","type":"uint256"},
        {"name":"paymentCurrency","type":"uint8"},
        {"name":"paymentLockType","type":"uint8"},
        {"name":"assetLockType","type":"uint8"}
      ]},
      {"name":"deliveryInfo","type":"tuple","components":[
        {"name":"paymentSourceChainSelector","type":"uint64"},
        {"name":"paymentDestinationChainSelector","type":"uint64"},
        {"name":"assetSourceChainSelector","type":"uint64"},
        {"name":"assetDestinationChainSelector","type":"uint64"}
      ]},
      {"name":"secretHash","type":"bytes32"},
      {"name":"executeAfter","type":"uint48"},
      {"name":"expiration","type":"uint48"},
      {"name":"ccipCallbackGasLimit","type":"uint32"},
      {"name":"data","type":"bytes"}
    ]}]
  }
]`

// OnLog processes DVP settlement logs:
// - Performs an EVM view call getSettlement(bytes32) and decodes the returned Settlement for metadata
// - Composes a verifiable event, signs it, and posts it to the CREC API
func OnLog(cfg *wfcommon.Config, rt cre.Runtime, payload *evm.Log) (string, error) {
	selector := cfg.ChainSelector

	if len(payload.Topics) < 3 {
		return "", fmt.Errorf("log topics length < 3")
	}
	settlementID := new(big.Int).SetBytes(payload.Topics[1])
	settlementHashBytes := payload.Topics[2]
	settlementHash := "0x" + fmt.Sprintf("%064x", new(big.Int).SetBytes(settlementHashBytes))

	rt.Logger().Info("processing DVP log", "settlementId", settlementID.String(), "settlementHash", settlementHash)

	dvpMeta := map[string]any{}
	settlementDecoded, err := fetchAndDecodeSettlement(rt, selector, cfg.DetectEventTriggerConfig.ContractAddress, settlementHashBytes)
	if err != nil {
		rt.Logger().Warn("failed to fetch settlement metadata, skipping enrichment", "settlementHash", settlementHash, "error", err)
	} else if settlementDecoded != nil {
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

	rt.Logger().Info("posting verifiable event", "eventName", eventName, "settlementId", settlementID.String(), "hasMetadata", len(dvpMeta) > 0)

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
		case uint8:
			return int64(t)
		case uint16:
			return int64(t)
		case uint32:
			return int64(t)
		case int64, int, uint64:
			return t
		default:
			return v
		}
	}
	for _, k := range []string{"settlement_id", "execute_after", "expiration", "ccip_callback_gas_limit"} {
		if v, ok := m[k]; ok {
			m[k] = toInt(v)
		}
	}
	if ti, ok := m["token_info"].(map[string]any); ok {
		for _, k := range []string{"payment_currency", "payment_lock_type", "asset_lock_type", "asset_token_amount", "payment_token_amount"} {
			if v, ok2 := ti[k]; ok2 {
				ti[k] = toInt(v)
			}
		}
		m["token_info"] = ti
	}
	return m
}
