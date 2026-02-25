package bundle

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"

	"github.com/andybalholm/brotli"
	crecbundle "github.com/smartcontractkit/crec-sdk/extension/bundle"
)

//go:embed binary.wasm.br.b64
var wasmBinaryBrB64 string

var wasmBinary = decodeCompressedBinary(wasmBinaryBrB64)

//go:embed CCIPDVPCoordinatorU.abi.json
var ccipdvpCoordinatorUABI string

// Get returns the dvp watcher bundle.
func Get() *crecbundle.Bundle {
	return &crecbundle.Bundle{
		Service:    "dvp",
		WasmBinary: wasmBinary,
		Contracts:  contracts,
		Events:         events,
	}
}

var contracts = []crecbundle.Contract{
	{Name: "CCIPDVPCoordinatorU", ABI: ccipdvpCoordinatorUABI},
}

// DataSchema for events enriched with on-chain settlement details from getSettlement(bytes32).
var settlementDataSchema = json.RawMessage(`{
	"type": "object",
	"properties": {
		"settlement_id": {"type": "string", "description": "String representation of the settlement ID"},
		"settlement_hash": {"type": "string", "description": "Hex-encoded settlement hash"},
		"dvp": {
			"type": "object",
			"description": "Full settlement struct from getSettlement(bytes32) on-chain call",
			"properties": {
				"ccip_callback_gas_limit": {"type": "integer"},
				"data": {"type": "string"},
				"delivery_info": {
					"type": "object",
					"properties": {
						"asset_destination_chain_selector": {"type": "integer"},
						"asset_source_chain_selector": {"type": "integer"},
						"payment_destination_chain_selector": {"type": "integer"},
						"payment_source_chain_selector": {"type": "integer"}
					}
				},
				"execute_after": {"type": "integer"},
				"expiration": {"type": "integer"},
				"party_info": {
					"type": "object",
					"properties": {
						"buyer_destination_address": {"type": "string"},
						"buyer_source_address": {"type": "string"},
						"executor_address": {"type": "string"},
						"seller_destination_address": {"type": "string"},
						"seller_source_address": {"type": "string"}
					}
				},
				"secret_hash": {"type": "string"},
				"settlement_id": {"type": "integer"},
				"token_info": {
					"type": "object",
					"properties": {
						"asset_token_amount": {"type": "integer"},
						"asset_token_destination_address": {"type": "string"},
						"asset_token_source_address": {"type": "string"},
						"asset_token_type": {"type": "integer"},
						"payment_currency": {"type": "integer"},
						"payment_token_amount": {"type": "integer"},
						"payment_token_destination_address": {"type": "string"},
						"payment_token_source_address": {"type": "string"},
						"payment_token_type": {"type": "integer"}
					}
				}
			}
		}
	}
}`)

var events = []crecbundle.Event{
	{Name: "SettlementOpened", TriggerContract: "CCIPDVPCoordinatorU", Description: "New settlement proposed", ParamsSchema: json.RawMessage(ParamsSchemas["SettlementOpened"]), DataSchema: settlementDataSchema},
	{Name: "SettlementAccepted", TriggerContract: "CCIPDVPCoordinatorU", Description: "Settlement accepted by counterparty", ParamsSchema: json.RawMessage(ParamsSchemas["SettlementAccepted"]), DataSchema: settlementDataSchema},
	{Name: "SettlementClosing", TriggerContract: "CCIPDVPCoordinatorU", Description: "Settlement in closing process", ParamsSchema: json.RawMessage(ParamsSchemas["SettlementClosing"]), DataSchema: settlementDataSchema},
	{Name: "SettlementSettled", TriggerContract: "CCIPDVPCoordinatorU", Description: "Settlement completed", ParamsSchema: json.RawMessage(ParamsSchemas["SettlementSettled"]), DataSchema: settlementDataSchema},
	{Name: "SettlementCanceling", TriggerContract: "CCIPDVPCoordinatorU", Description: "Settlement being canceled", ParamsSchema: json.RawMessage(ParamsSchemas["SettlementCanceling"]), DataSchema: settlementDataSchema},
	{Name: "SettlementCanceled", TriggerContract: "CCIPDVPCoordinatorU", Description: "Settlement canceled", ParamsSchema: json.RawMessage(ParamsSchemas["SettlementCanceled"]), DataSchema: settlementDataSchema},
}

func decodeCompressedBinary(encoded string) []byte {
	data := strings.TrimSpace(encoded)
	if data == "" {
		return nil
	}
	compressed, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		panic("crec bundle: invalid base64 wasm: " + err.Error())
	}
	raw, err := io.ReadAll(brotli.NewReader(bytes.NewReader(compressed)))
	if err != nil {
		panic("crec bundle: decompress wasm: " + err.Error())
	}
	return raw
}
