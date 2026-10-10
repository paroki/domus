package authz

type Resource string
type Action string

const (
	ResourceAccounts Resource = "units"
	ResourceDiocese  Resource = "dioceses"
	ResourceParish   Resource = "parishes"
)

const (
	ActionRead  Action = "read"
	ActionWrite Action = "write"
)
