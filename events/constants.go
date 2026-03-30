package events

// Lock type constants for DvP settlements.
const (
	LockTypeNone = iota
	LockTypeERC20
	LockTypeERC3643
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
