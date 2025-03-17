package constants

type ContextKey string

const (
	UserSessionKey        string     = "user-sessions"
	UserKey               string     = "user"
	UserIdKey             string     = "user_id"
	AuthenticatedKey      string     = "authenticated"
	ChangesetIdKey        string     = "changeset_id"
	UserContextKey        ContextKey = ContextKey(UserKey)
	UnitOfWorkKey         ContextKey = "unitOfWork"
	ChangesetIdContextKey ContextKey = ContextKey(ChangesetIdKey)
)

type PermissionLevel int

const (
	PermissionViewer PermissionLevel = iota
	PermissionEditor
	PermissionAdmin
)
