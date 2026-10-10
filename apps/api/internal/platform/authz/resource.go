package authz

type Resource string
type Action string

const (
	ResourceAccounts Resource = "units"
)

const (
	ActionRead  Action = "read"
	ActionWrite Action = "write"
)
