package bundle

import (
	_ "embed"
	"encoding/json"

	crecbundle "github.com/smartcontractkit/crec-sdk/extension/bundle"
)

//go:embed binary.wasm
var wasmBinary []byte

//go:embed config.tmpl
var configTemplate []byte

//go:embed CCIPDVPCoordinatorU.abi.json
var dvpCoordinatorABI string

// Get returns the DVP extension watcher bundle.
func Get() *crecbundle.Bundle {
	return &crecbundle.Bundle{
		Service:        "dvp",
		WasmBinary:     wasmBinary,
		ConfigTemplate: configTemplate,
		Contracts:      contracts,
		Events:         events,
	}
}

var contracts = []crecbundle.Contract{
	{Name: "CCIPDVPCoordinator", ABI: dvpCoordinatorABI},
}

// All DVP settlement events share the same params: (uint256 indexed settlementId, bytes32 indexed settlementHash).
var settlementParamsSchema = json.RawMessage(`{
	"type": "object",
	"properties": {
		"settlementId": {"type": "string", "description": "uint256 settlement ID"},
		"settlementHash": {"type": "string", "description": "bytes32 settlement hash"}
	},
	"required": ["settlementId", "settlementHash"]
}`)

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
	{Name: "SettlementOpened", TriggerContract: "CCIPDVPCoordinator", Description: "New settlement proposed", ParamsSchema: settlementParamsSchema, DataSchema: settlementDataSchema},
	{Name: "SettlementAccepted", TriggerContract: "CCIPDVPCoordinator", Description: "Settlement accepted by counterparty", ParamsSchema: settlementParamsSchema, DataSchema: settlementDataSchema},
	{Name: "SettlementClosing", TriggerContract: "CCIPDVPCoordinator", Description: "Settlement in closing process", ParamsSchema: settlementParamsSchema, DataSchema: settlementDataSchema},
	{Name: "SettlementSettled", TriggerContract: "CCIPDVPCoordinator", Description: "Settlement completed", ParamsSchema: settlementParamsSchema, DataSchema: settlementDataSchema},
	{Name: "SettlementCanceling", TriggerContract: "CCIPDVPCoordinator", Description: "Settlement being canceled", ParamsSchema: settlementParamsSchema, DataSchema: settlementDataSchema},
	{Name: "SettlementCanceled", TriggerContract: "CCIPDVPCoordinator", Description: "Settlement canceled", ParamsSchema: settlementParamsSchema, DataSchema: settlementDataSchema},
}
