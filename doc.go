// Package dvp provides a CREC SDK extension for Delivery versus Payment (DvP) settlement operations.
//
// This package is organized into sub-packages:
//
//   - dvp/events:     Event types, Settlement struct, constants (TokenType), and decoders.
//   - dvp/currency:   ISO 4217 currency code mapping for DvP settlements.
//   - dvp/operations: Extension client for preparing on-chain operations.
//
// The root dvp package provides [DecodeFromEvent] for SDK consumers to decode
// watcher event payloads into typed Go structs.
//
// # Decoding Events
//
//	decoded, err := dvp.DecodeFromEvent(ctx, event)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(decoded.EventName())
//
// # Preparing Operations
//
//	import "github.com/smartcontractkit/crec-sdk-ext-dvp/operations"
//	import "github.com/smartcontractkit/crec-sdk-ext-dvp/events"
//
//	ext, err := operations.New(&operations.Options{
//		CCIPDVPCoordinatorUAddress: "0x...",
//		AccountAddress:             "0x...",
//	})
//	op, err := ext.PrepareProposeSettlementOperation(settlement)
//	hash, err := operations.HashSettlement(settlement)
package dvp
