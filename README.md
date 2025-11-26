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
    "github.com/smartcontractkit/crec-sdk-ext-dvp"
)

// Create the DVP extension
ext, err := dvp.New(&dvp.Options{
    DvpCoordinatorAddress: "0x...",  // DVP Coordinator contract address
    AccountAddress:        "0x...",  // Your account address
})
if err != nil {
    log.Fatal(err)
}
```

### Proposing a Settlement

```go
import (
    "github.com/smartcontractkit/crec-api-go/services/dvp/gen/contract"
    "github.com/ethereum/go-ethereum/common"
    "math/big"
)

// Create settlement details
settlement := &contract.Settlement{
    SettlementId: big.NewInt(1),
    PartyInfo: contract.PartyInfo{
        BuyerSourceAddress:       common.HexToAddress("0xBuyer..."),
        BuyerDestinationAddress:  common.HexToAddress("0xBuyer..."),
        SellerSourceAddress:      common.HexToAddress("0xSeller..."),
        SellerDestinationAddress: common.HexToAddress("0xSeller..."),
        ExecutorAddress:          common.HexToAddress("0xExecutor..."),
    },
    TokenInfo: contract.TokenInfo{
        PaymentTokenAmount:           big.NewInt(1000000),
        AssetTokenAmount:             big.NewInt(1000000000000000000),
        PaymentTokenSourceAddress:    common.HexToAddress("0xPaymentToken..."),
        AssetTokenSourceAddress:      common.HexToAddress("0xAssetToken..."),
        PaymentTokenType:             dvp.TokenTypeERC20,
        AssetTokenType:               dvp.TokenTypeERC20,
        PaymentCurrency:              dvp.CurrencyMap["USD"],
    },
    // ... other fields
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
op, err := ext.PrepareAcceptSettlementOperation(settlementHash)
```

### Executing a Settlement

```go
settlementHash := common.HexToHash("0x...")
op, err := ext.PrepareExecuteSettlementOperation(settlementHash)
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

The `CurrencyMap` provides ISO 4217 currency codes for off-chain payment specifications:

```go
// Example: Get USD currency code
usdCode := dvp.CurrencyMap["USD"]  // Returns 147
```

## License

See [LICENSE](LICENSE) for details.
