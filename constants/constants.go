package constants

type ContextKey string

const (
	UserSessionKey   string     = "user-sessions"
	UserKey          string     = "user"
	UserIdKey        string     = "user_id"
	AuthenticatedKey string     = "authenticated"
	UserContextKey   ContextKey = ContextKey(UserKey)
	UnitOfWorkKey    ContextKey = "unitOfWork"
)

type PermissionLevel int

const (
	PermissionViewer PermissionLevel = iota
	PermissionEditor
	PermissionAdmin
)
