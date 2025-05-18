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
	validationService       *ValidationService
	validator               *Validator
}

func NewUserService(queries *db.Queries, variationContextService *VariationContextService, validationService *ValidationService, validator *Validator) *UserService {
	return &UserService{queries: queries, variationContextService: variationContextService, validationService: validationService, validator: validator}
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

type UserDto struct {
	ID                  uint   `json:"id" validate:"required"`
	Username            string `json:"username" validate:"required"`
	GlobalAdministrator bool   `json:"globalAdministrator" validate:"required"`
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

type UsersFilter struct {
	Limit  int
	Offset int
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

type UpdateUserParams struct {
	GlobalAdministrator bool
}

func (s *UserService) validateUpdateUser(ctx context.Context, userID uint) error {
	currentUser := auth.GetUserFromContext(ctx)
	if !currentUser.IsGlobalAdmin {
		return NewServiceError(ErrorCodePermissionDenied, "You are not authorized to update a user")
	}

	_, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return NewDbError(err, "User")
	}

	return nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID uint, params UpdateUserParams) error {
	if err := s.validateUpdateUser(ctx, userID); err != nil {
		return err
	}

	return s.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:                  userID,
		GlobalAdministrator: params.GlobalAdministrator,
	})
}

type CreateUserParams struct {
	Username            string
	Password            string
	GlobalAdministrator bool
}

func (s *UserService) validateCreateUser(ctx context.Context, data CreateUserParams) error {
	user := auth.GetUserFromContext(ctx)
	if !user.IsGlobalAdmin {
		return NewServiceError(ErrorCodePermissionDenied, "You are not authorized to create a user")
	}

	if taken, err := s.validationService.IsUsernameTaken(ctx, data.Username); err != nil {
		return err
	} else if taken {
		return NewServiceError(ErrorCodeInvalidInput, "Username already exists")
	}

	return s.validator.
		Validate(data.Username, "Username").Required().MinLength(1).MaxLength(100).
		Validate(data.Password, "Password").Required().MinLength(8).
		Error(ctx)
}

func (s *UserService) CreateUser(ctx context.Context, params CreateUserParams) (uint, error) {
	if err := s.validateCreateUser(ctx, params); err != nil {
		return 0, err
	}

	passwordHash, err := auth.GeneratePasswordHash(params.Password)
	if err != nil {
		return 0, NewServiceError(ErrorCodeInvalidOperation, "Failed to hash password")
	}

	userID, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Name:                params.Username,
		Password:            string(passwordHash),
		GlobalAdministrator: params.GlobalAdministrator,
	})
	if err != nil {
		return 0, err
	}

	return userID, nil
}
