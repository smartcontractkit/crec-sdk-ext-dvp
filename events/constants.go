package events

// Token type constants for DvP settlements.
const (
	TokenTypeNone = iota
	TokenTypeERC20
	TokenTypeERC3643
)

// Settlement status constants.
const (
	SettlementStatusNew = iota
	SettlementStatusOpen
	SettlementStatusAccepted
	SettlementStatusClosing
	SettlementStatusSettled
	SettlementStatusCanceled
)

// ServiceName is the DvP service identifier.
const ServiceName = "dvp"
