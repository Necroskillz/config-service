package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
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
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrRecordNotFound
		}

		return 0, err
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
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrRecordNotFound
		}
		return User{}, err
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
