# CREC SDK Extension: DVP

A Go SDK extension for Delivery versus Payment (DvP) settlement operations on CREC. The DvP (Delivery vs. Payment) service allows for the secure and trustless transfer of assets between parties.

## Installation

```bash
go get github.com/smartcontractkit/crec-sdk-ext-dvp
```

## Overview

This extension provides utilities for preparing DvP (Delivery versus Payment) settlement operations, enabling atomic exchange of assets between counterparties on blockchain networks.

### Features

- Propose settlements with optional token approvals or holds
- Accept and execute settlements
- Compute settlement hashes for verification
- Support for ERC20 and ERC3643 token types
- ISO 4217 currency code mapping

## Usage

### Basic Setup

```go
import (
    "github.com/smartcontractkit/crec-sdk-ext-dvp/operations"
)

ext, err := operations.New(&operations.Options{
    CCIPDVPCoordinatorUAddress: "0x...",
    AccountAddress:             "0x...",
})
if err != nil {
    log.Fatal(err)
}
```

### Proposing a Settlement

```go
import (
    "github.com/smartcontractkit/crec-sdk-ext-dvp/events"
    "github.com/ethereum/go-ethereum/common"
    "math/big"
)

// Create settlement details
settlement := &events.Settlement{
    SettlementId: big.NewInt(1),
    PartyInfo: events.PartyInfo{
        BuyerSourceAddress:       common.HexToAddress("0xBuyer..."),
        BuyerDestinationAddress:  common.HexToAddress("0xBuyer..."),
        SellerSourceAddress:      common.HexToAddress("0xSeller..."),
        SellerDestinationAddress: common.HexToAddress("0xSeller..."),
        ExecutorAddress:          common.HexToAddress("0xExecutor..."),
    },
    TokenInfo: events.TokenInfo{
        PaymentTokenSourceAddress:      common.HexToAddress("0xPaymentToken..."),
        PaymentTokenDestinationAddress: common.HexToAddress("0xPaymentToken..."),
        AssetTokenSourceAddress:        common.HexToAddress("0xAssetToken..."),
        AssetTokenDestinationAddress:   common.HexToAddress("0xAssetToken..."),
        PaymentCurrency:                events.CurrencyMap["USD"],
        PaymentTokenAmount:             big.NewInt(1000000),
        AssetTokenAmount:               big.NewInt(1000000000000000000),
        PaymentTokenType:               events.TokenTypeERC20,
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

// Option 1: Propose without token approval (assumes approval exists)
op, err := ext.PrepareProposeSettlementOperation(settlement)

// Option 2: Propose with automatic token approval
op, err := ext.PrepareProposeSettlementWithTokenApprovalOperation(settlement)

// Option 3: Propose with token hold (ERC3643 only)
op, err := ext.PrepareProposeSettlementWithTokenHoldOperation(settlement, holdManagerAddress)
```

### Accepting a Settlement

```go
settlementHash := common.HexToHash("0x...")
op, err := ext.PrepareAcceptSettlementOperation([32]byte(settlementHash))
```

### Executing a Settlement

```go
settlementHash := common.HexToHash("0x...")
op, err := ext.PrepareExecuteSettlementOperation([32]byte(settlementHash))
```

### Computing Settlement Hash

```go
hash, err := ext.HashSettlement(settlement)
fmt.Printf("Settlement hash: %s\n", hash.Hex())
```

## Token Types

| Constant | Value | Description |
|----------|-------|-------------|
| `TokenTypeNone` | 0 | No token type specified |
| `TokenTypeERC20` | 1 | Standard ERC20 token |
| `TokenTypeERC3643` | 2 | Security token (T-REX) |

## Settlement Status

| Constant | Value | Description |
|----------|-------|-------------|
| `SettlementStatusNew` | 0 | Settlement created |
| `SettlementStatusOpen` | 1 | Settlement proposed |
| `SettlementStatusAccepted` | 2 | Settlement accepted |
| `SettlementStatusClosing` | 3 | Settlement in closing state |
| `SettlementStatusSettled` | 4 | Settlement completed |
| `SettlementStatusCanceled` | 5 | Settlement canceled |

## Currency Codes

The `CurrencyMap` in the events package provides ISO 4217 currency codes for off-chain payment specifications:

```go
import "github.com/smartcontractkit/crec-sdk-ext-dvp/events"

usdCode := events.CurrencyMap["USD"]  // Returns 147
```

## License

[LICENSE](LICENSE.md)
