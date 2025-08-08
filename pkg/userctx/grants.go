package userctx

type Role string

const (
	// Provides the ability to view all data relating to the application,
	// as well as managing its lifecycle and reading logs.
	RoleAdmin Role = "admin"
	// Allows lifecycle management and log viewing, but doesn't permit the
	// viewing of potentially sensitive information like environment variables.
	RoleOperator Role = "operator"
	// Allows viewing application state and logs, but no control over application lifecycle
	// or access to other sensitive information.
	RoleViewer Role = "viewer"
	// Prohibits access to any resources associated with a given application
	RoleNone Role = ""
)

func (r Role) IsAllowed() bool {
	return r != RoleNone
}

type RoleAssignment struct {
	Role         Role     `json:"role"`
	Applications []string `json:"applications"`
}
