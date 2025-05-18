package service

import (
	"context"

	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/db"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	queries                 *db.Queries
	variationContextService *VariationContextService
}

func NewUserService(queries *db.Queries, variationContextService *VariationContextService) *UserService {
	return &UserService{queries: queries, variationContextService: variationContextService}
}

func (s *UserService) Authenticate(ctx context.Context, name, password string) (uint, error) {
	user, err := s.queries.GetUserByName(ctx, name)
	if err != nil {
		return 0, NewDbError(err, "User")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return 0, ErrInvalidPassword
	}

	return user.ID, nil
}

type User struct {
	ID                  uint
	Username            string
	GlobalAdministrator bool
	Permissions         []UserPermission
}

type UserPermission struct {
	ServiceID  uint
	FeatureID  *uint
	KeyID      *uint
	Variation  map[uint]string
	Permission constants.PermissionLevel
}

func dbPermissionToConstant(dbPerm db.PermissionLevel) constants.PermissionLevel {
	switch dbPerm {
	case db.PermissionLevelAdmin:
		return constants.PermissionAdmin
	case db.PermissionLevelEditor:
		return constants.PermissionEditor
	default:
		return constants.PermissionViewer
	}
}

func (s *UserService) Get(ctx context.Context, id uint) (User, error) {
	dbUser, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		return User{}, NewDbError(err, "User")
	}

	dbPermissions, err := s.queries.GetUserPermissions(ctx, id)
	if err != nil {
		return User{}, err
	}

	permissions := make([]UserPermission, len(dbPermissions))
	for i, p := range dbPermissions {
		var variationContext map[uint]string

		if p.VariationContextID != nil {
			variationContext, err = s.variationContextService.GetVariationContextValues(ctx, *p.VariationContextID)
			if err != nil {
				return User{}, err
			}
		}

		perm := UserPermission{
			ServiceID:  p.ServiceID,
			Permission: dbPermissionToConstant(p.Permission),
			FeatureID:  p.FeatureID,
			KeyID:      p.KeyID,
			Variation:  variationContext,
		}

		permissions[i] = perm
	}

	return User{
		ID:                  dbUser.ID,
		Username:            dbUser.Name,
		GlobalAdministrator: dbUser.GlobalAdministrator,
		Permissions:         permissions,
	}, nil
}

func (s *UserService) GetUsers(ctx context.Context, filter UsersFilter) (PaginatedResult[UserDto], error) {
	if filter.Limit > 100 {
		return PaginatedResult[UserDto]{}, NewServiceError(ErrorCodeInvalidOperation, "Limit cannot be greater than 100")
	}

	users, err := s.queries.GetUsers(ctx, db.GetUsersParams{
		Limit:  filter.Limit,
		Offset: filter.Offset,
	})
	if err != nil {
		return PaginatedResult[UserDto]{}, NewDbError(err, "Users")
	}

	userItems := make([]UserDto, len(users))
	for i, user := range users {
		userItems[i] = UserDto{
			ID:                  user.ID,
			Username:            user.Name,
			GlobalAdministrator: user.GlobalAdministrator,
		}
	}

	var total int
	if len(users) > 0 {
		total = users[0].TotalCount
	}

	return PaginatedResult[UserDto]{
		Items:      userItems,
		TotalCount: total,
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID uint, globalAdministrator bool) error {
	currentUser := auth.GetUserFromContext(ctx)
	if !currentUser.IsGlobalAdmin {
		return NewServiceError(ErrorCodePermissionDenied, "You are not authorized to update a user")
	}

	err := s.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:                  userID,
		GlobalAdministrator: globalAdministrator,
	})
	if err != nil {
		return NewDbError(err, "UpdateUser")
	}
	return nil
}

func (s *UserService) CreateUser(ctx context.Context, name string, password string, globalAdministrator bool) (uint, error) {
	currentUser := auth.GetUserFromContext(ctx)
	if !currentUser.IsGlobalAdmin {
		return 0, NewServiceError(ErrorCodePermissionDenied, "You are not authorized to create a user")
	}

	userID, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Name:                name,
		Password:            password,
		GlobalAdministrator: globalAdministrator,
	})
	if err != nil {
		return 0, NewDbError(err, "CreateUser")
	}
	return userID, nil
}

type UsersFilter struct {
	Limit  int
	Offset int
}

type UserDto struct {
	ID                  uint   `json:"id" validate:"required"`
	Username            string `json:"username" validate:"required"`
	GlobalAdministrator bool   `json:"globalAdministrator" validate:"required"`
}
