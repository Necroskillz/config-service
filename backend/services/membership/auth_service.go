package membership

import (
	"context"

	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/services/core"
	"github.com/necroskillz/config-service/services/validation"
	"github.com/necroskillz/config-service/services/variation"
	"github.com/necroskillz/config-service/util/validator"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	queries                 *db.Queries
	variationContextService *variation.ContextService
	validationService       *validation.Service
	validator               *validator.Validator
}

func NewAuthService(queries *db.Queries, variationContextService *variation.ContextService, validationService *validation.Service, validator *validator.Validator) *AuthService {
	return &AuthService{queries: queries, variationContextService: variationContextService, validationService: validationService, validator: validator}
}

func (s *AuthService) Authenticate(ctx context.Context, name, password string) (uint, error) {
	user, err := s.queries.GetUserByName(ctx, name)
	if err != nil {
		return 0, core.NewDbError(err, "User")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return 0, core.NewServiceError(core.ErrorCodeInvalidPassword, "Invalid password")
	}

	return user.ID, nil
}

type User struct {
	ID                  uint
	Username            string
	GlobalAdministrator bool
	Permissions         []Permission
}

type Permission struct {
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

func (s *AuthService) GetUser(ctx context.Context, id uint) (User, error) {
	dbUser, err := s.queries.GetUserByID(ctx, id)
	if err != nil {
		return User{}, core.NewDbError(err, "User")
	}

	dbPermissions, err := s.queries.GetPermissions(ctx, id)
	if err != nil {
		return User{}, err
	}

	permissions := make([]Permission, len(dbPermissions))
	for i, p := range dbPermissions {
		var variation map[uint]string

		if p.VariationContextID != nil {
			variation, err = s.variationContextService.GetVariationContextValues(ctx, *p.VariationContextID)
			if err != nil {
				return User{}, err
			}
		}

		perm := Permission{
			ServiceID:  p.ServiceID,
			Permission: dbPermissionToConstant(p.Permission),
			FeatureID:  p.FeatureID,
			KeyID:      p.KeyID,
			Variation:  variation,
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
