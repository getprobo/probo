package coredata

type (
	ExpireReason string
)

const (
	ExpireReasonIdleTimeout ExpireReason = "idle_timeout"
	ExpireReasonRevoked     ExpireReason = "revoked"
	ExpireReasonClosed      ExpireReason = "closed"
)
