package membership

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/services/core"
	"github.com/necroskillz/config-service/services/validation"
	"github.com/necroskillz/config-service/services/variation"
	"github.com/necroskillz/config-service/util/ptr"
	"github.com/necroskillz/config-service/util/validator"
)

type Service struct {
	queries                   *db.Queries
	variationContextService   *variation.ContextService
	validationService         *validation.Service
	variationHierarchyService *variation.HierarchyService
	validator                 *validator.Validator
	coreService               *core.Service
}

func NewService(
	queries *db.Queries,
	variationContextService *variation.ContextService,
	validationService *validation.Service,
	variationHierarchyService *variation.HierarchyService,
	validator *validator.Validator,
	coreService *core.Service,
) *Service {
	return &Service{
		queries:                   queries,
		variationContextService:   variationContextService,
		validationService:         validationService,
		variationHierarchyService: variationHierarchyService,
		validator:                 validator,
		coreService:               coreService,
	}
}

type UsersFilter struct {
	Page     int
	PageSize int
	Name     *string
	Type     *string
}

type MembershipObjectType string

const (
	MembershipObjectTypeUser        MembershipObjectType = "user"
	MembershipObjectTypeGlobalAdmin MembershipObjectType = "global_administrator"
	MembershipObjectTypeGroup       MembershipObjectType = "group"
)

type MembershipObjectDto struct {
	ID   uint                 `json:"id" validate:"required"`
	Name string               `json:"name" validate:"required"`
	Type MembershipObjectType `json:"type" validate:"required"`
}

func (s *Service) GetUsersAndGroups(ctx context.Context, filter UsersFilter) (core.PaginatedResult[MembershipObjectDto], error) {
	if filter.Page < 1 {
		return core.PaginatedResult[MembershipObjectDto]{}, core.NewServiceError(core.ErrorCodeInvalidOperation, "Page must be 1 or greater")
	}

	if filter.PageSize < 1 || filter.PageSize > 100 {
		return core.PaginatedResult[MembershipObjectDto]{}, core.NewServiceError(core.ErrorCodeInvalidOperation, "Page size must be between 1 and 100")
	}

	if filter.Type != nil && *filter.Type != "user" && *filter.Type != "group" {
		return core.PaginatedResult[MembershipObjectDto]{}, core.NewServiceError(core.ErrorCodeInvalidOperation, fmt.Sprintf("Invalid type: %s. Allowed types are 'user' and 'group' or nil", *filter.Type))
	}

	membershipObjects, err := s.queries.GetUsersAndGroups(ctx, db.GetUsersAndGroupsParams{
		Limit:  filter.PageSize,
		Offset: (filter.Page - 1) * filter.PageSize,
		Name:   filter.Name,
		Type:   filter.Type,
	})
	if err != nil {
		return core.PaginatedResult[MembershipObjectDto]{}, core.NewDbError(err, "Users")
	}

	items := make([]MembershipObjectDto, len(membershipObjects))
	for i, membershipObject := range membershipObjects {
		items[i] = MembershipObjectDto{
			ID:   membershipObject.ID,
			Name: membershipObject.Name,
			Type: MembershipObjectType(membershipObject.Type),
		}
	}

	var total int
	if len(membershipObjects) > 0 {
		total = membershipObjects[0].TotalCount
	}

	return core.PaginatedResult[MembershipObjectDto]{
		Items:      items,
		TotalCount: total,
	}, nil
}

type UserGroupDto struct {
	ID   uint   `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type PermissionDto struct {
	ID          uint               `json:"id" validate:"required"`
	Kind        db.PermissionKind  `json:"kind" validate:"required"`
	ServiceID   uint               `json:"serviceId" validate:"required"`
	ServiceName string             `json:"serviceName" validate:"required"`
	FeatureID   *uint              `json:"featureId"`
	FeatureName *string            `json:"featureName"`
	KeyID       *uint              `json:"keyId"`
	KeyName     *string            `json:"keyName"`
	Variation   map[string]string  `json:"variation"`
	Permission  db.PermissionLevel `json:"permission"`
	GroupID     *uint              `json:"groupId"`
	GroupName   *string            `json:"groupName"`
}

type UserDto struct {
	Username            string          `json:"username" validate:"required"`
	GlobalAdministrator bool            `json:"globalAdministrator" validate:"required"`
	Groups              []UserGroupDto  `json:"groups" validate:"required"`
	Permissions         []PermissionDto `json:"permissions" validate:"required"`
}

func (s *Service) makePermissionDto(ctx context.Context, permission db.GetPermissionsForMembershipObjectRow, groupMap map[uint]UserGroupDto) (PermissionDto, error) {
	permissionDto := PermissionDto{
		ID:          permission.ID,
		Kind:        permission.Kind,
		ServiceID:   permission.ServiceID,
		ServiceName: permission.ServiceName,
		FeatureID:   permission.FeatureID,
		FeatureName: permission.FeatureName,
		KeyID:       permission.KeyID,
		KeyName:     permission.KeyName,
		Permission:  permission.Permission,
	}

	if groupMap != nil && permission.UserGroupID != nil {
		permissionDto.GroupID = permission.UserGroupID
		name := groupMap[*permission.UserGroupID].Name
		permissionDto.GroupName = &name
	}

	if permission.VariationContextID != nil {
		variation, err := s.variationContextService.GetVariationContextValues(ctx, *permission.VariationContextID)
		if err != nil {
			return PermissionDto{}, err
		}

		variationHierachy, err := s.variationHierarchyService.GetVariationHierarchy(ctx)
		if err != nil {
			return PermissionDto{}, err
		}

		variationStringMap, err := variationHierachy.GetVariationStringMap(variation)
		if err != nil {
			return PermissionDto{}, err
		}

		permissionDto.Variation = variationStringMap
	}

	return permissionDto, nil
}

func (s *Service) GetUser(ctx context.Context, userID uint) (UserDto, error) {
	user, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return UserDto{}, core.NewDbError(err, "User")
	}

	groups, err := s.queries.GetUserGroups(ctx, userID)
	if err != nil {
		return UserDto{}, core.NewDbError(err, "UserGroups")
	}

	permissions, err := s.queries.GetPermissionsForMembershipObject(ctx, db.GetPermissionsForMembershipObjectParams{
		UserID: ptr.To(userID),
	})
	if err != nil {
		return UserDto{}, core.NewDbError(err, "UserPermissions")
	}

	groupMap := make(map[uint]UserGroupDto)
	userGroups := make([]UserGroupDto, len(groups))
	for i, group := range groups {
		userGroups[i] = UserGroupDto{
			ID:   group.ID,
			Name: group.Name,
		}
		groupMap[group.ID] = userGroups[i]
	}

	userPermissions := make([]PermissionDto, len(permissions))
	for i, permission := range permissions {
		userPermissions[i], err = s.makePermissionDto(ctx, permission, groupMap)
		if err != nil {
			return UserDto{}, err
		}
	}

	return UserDto{
		Username:            user.Name,
		GlobalAdministrator: user.GlobalAdministrator,
		Groups:              userGroups,
		Permissions:         userPermissions,
	}, nil
}

type UpdateUserParams struct {
	GlobalAdministrator bool
}

func (s *Service) validateUpdateUser(ctx context.Context, userID uint) error {
	currentUser := auth.GetUserFromContext(ctx)
	if !currentUser.IsGlobalAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to update a user")
	}

	_, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return core.NewDbError(err, "User")
	}

	return nil
}

func (s *Service) UpdateUser(ctx context.Context, userID uint, params UpdateUserParams) error {
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

func (s *Service) validateCreateUser(ctx context.Context, data CreateUserParams) error {
	user := auth.GetUserFromContext(ctx)
	if !user.IsGlobalAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to create a user")
	}

	if err := s.validator.
		Validate(data.Username, "Username").Required().MinLength(1).MaxLength(100).Regex(`^[\w\-_\.]+$`).
		Validate(data.Password, "Password").Required().MinLength(8).
		Error(ctx); err != nil {
		return err
	}

	if taken, err := s.validationService.IsUsernameTaken(ctx, data.Username); err != nil {
		return err
	} else if taken {
		return core.NewServiceError(core.ErrorCodeInvalidInput, "Username already exists")
	}

	return nil
}

func (s *Service) CreateUser(ctx context.Context, params CreateUserParams) (uint, error) {
	if err := s.validateCreateUser(ctx, params); err != nil {
		return 0, err
	}

	passwordHash, err := auth.GeneratePasswordHash(params.Password)
	if err != nil {
		return 0, core.NewServiceError(core.ErrorCodeInvalidOperation, "Failed to hash password")
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

func (s *Service) validateDeleteUser(ctx context.Context, userID uint) error {
	user := auth.GetUserFromContext(ctx)
	if !user.IsGlobalAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to delete a user")
	}

	_, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return core.NewDbError(err, "User")
	}

	return nil
}

func (s *Service) DeleteUser(ctx context.Context, userID uint) error {
	if err := s.validateDeleteUser(ctx, userID); err != nil {
		return err
	}

	return s.queries.DeleteUser(ctx, userID)
}

type GroupUserDto struct {
	ID   uint   `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type GroupDto struct {
	Name        string          `json:"name" validate:"required"`
	Permissions []PermissionDto `json:"permissions" validate:"required"`
}

func (s *Service) GetGroup(ctx context.Context, groupID uint) (GroupDto, error) {
	group, err := s.queries.GetGroupByID(ctx, groupID)
	if err != nil {
		return GroupDto{}, core.NewDbError(err, "Group")
	}

	permissions, err := s.queries.GetPermissionsForMembershipObject(ctx, db.GetPermissionsForMembershipObjectParams{
		GroupID: ptr.To(groupID),
	})
	if err != nil {
		return GroupDto{}, core.NewDbError(err, "GroupPermissions")
	}

	groupPermissions := make([]PermissionDto, len(permissions))
	for i, permission := range permissions {
		groupPermissions[i], err = s.makePermissionDto(ctx, permission, nil)
		if err != nil {
			return GroupDto{}, err
		}
	}

	return GroupDto{
		Name:        group.Name,
		Permissions: groupPermissions,
	}, nil
}

type GetGroupUsersParams struct {
	GroupID  uint
	Page     int
	PageSize int
}

func (s *Service) GetGroupUsers(ctx context.Context, params GetGroupUsersParams) (core.PaginatedResult[GroupUserDto], error) {
	users, err := s.queries.GetGroupUsers(ctx, db.GetGroupUsersParams{
		ID:     params.GroupID,
		Limit:  params.PageSize,
		Offset: (params.Page - 1) * params.PageSize,
	})
	if err != nil {
		return core.PaginatedResult[GroupUserDto]{}, core.NewDbError(err, "GroupUsers")
	}

	groupUsers := make([]GroupUserDto, len(users))
	for i, user := range users {
		groupUsers[i] = GroupUserDto{
			ID:   user.ID,
			Name: user.Name,
		}
	}

	var total int
	if len(users) > 0 {
		total = users[0].TotalCount
	}

	return core.PaginatedResult[GroupUserDto]{
		Items:      groupUsers,
		TotalCount: total,
	}, nil
}

type CreateGroupParams struct {
	Name string
}

func (s *Service) validateCreateGroup(ctx context.Context, data CreateGroupParams) error {
	user := auth.GetUserFromContext(ctx)
	if !user.IsGlobalAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to create a group")
	}

	if err := s.validator.
		Validate(data.Name, "Name").Required().MinLength(1).MaxLength(100).Regex(`^[\w\-_\.]+$`).
		Error(ctx); err != nil {
		return err
	}

	if taken, err := s.validationService.IsGroupNameTaken(ctx, data.Name); err != nil {
		return err
	} else if taken {
		return core.NewServiceError(core.ErrorCodeInvalidInput, "Group name already exists")
	}

	return nil
}

func (s *Service) CreateGroup(ctx context.Context, params CreateGroupParams) (uint, error) {
	if err := s.validateCreateGroup(ctx, params); err != nil {
		return 0, err
	}

	groupID, err := s.queries.CreateGroup(ctx, params.Name)
	if err != nil {
		return 0, err
	}

	return groupID, nil
}

func (s *Service) validateDeleteGroup(ctx context.Context, groupID uint) error {
	user := auth.GetUserFromContext(ctx)
	if !user.IsGlobalAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to delete a group")
	}

	_, err := s.queries.GetGroupByID(ctx, groupID)
	if err != nil {
		return core.NewDbError(err, "Group")
	}

	return nil
}

func (s *Service) DeleteGroup(ctx context.Context, groupID uint) error {
	if err := s.validateDeleteGroup(ctx, groupID); err != nil {
		return err
	}

	return s.queries.DeleteGroup(ctx, groupID)
}

func (s *Service) validateAddUserToGroup(ctx context.Context, userID uint, groupID uint) error {
	user := auth.GetUserFromContext(ctx)
	if !user.IsGlobalAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to add a user to a group")
	}

	_, err := s.queries.GetUserGroupMembership(ctx, db.GetUserGroupMembershipParams{
		UserID:      userID,
		UserGroupID: groupID,
	})

	if err == nil {
		return core.NewServiceError(core.ErrorCodeInvalidInput, "User is already a member of this group")
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	return nil
}

func (s *Service) AddUserToGroup(ctx context.Context, userID uint, groupID uint) error {
	if err := s.validateAddUserToGroup(ctx, userID, groupID); err != nil {
		return err
	}

	return s.queries.CreateUserGroupMembership(ctx, db.CreateUserGroupMembershipParams{
		UserID:      userID,
		UserGroupID: groupID,
	})
}

func (s *Service) validateRemoveUserFromGroup(ctx context.Context, userID uint, groupID uint) error {
	user := auth.GetUserFromContext(ctx)
	if !user.IsGlobalAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to remove a user from a group")
	}

	_, err := s.queries.GetUserGroupMembership(ctx, db.GetUserGroupMembershipParams{
		UserID:      userID,
		UserGroupID: groupID,
	})
	if err != nil {
		return core.NewDbError(err, "UserGroupMembership")
	}

	return nil
}

func (s *Service) RemoveUserFromGroup(ctx context.Context, userID uint, groupID uint) error {
	if err := s.validateRemoveUserFromGroup(ctx, userID, groupID); err != nil {
		return err
	}

	return s.queries.DeleteUserGroupMembership(ctx, db.DeleteUserGroupMembershipParams{
		UserID:      userID,
		UserGroupID: groupID,
	})
}

func (s *Service) validateRemovePermission(ctx context.Context, permissionID uint) error {
	permission, err := s.queries.GetPermissionByID(ctx, permissionID)
	if err != nil {
		return core.NewDbError(err, "Permission")
	}

	user := auth.GetUserFromContext(ctx)
	if user.GetPermissionForService(permission.ServiceID) < constants.PermissionAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to remove this permission")
	}

	return nil
}

func (s *Service) RemovePermission(ctx context.Context, permissionID uint) error {
	if err := s.validateRemovePermission(ctx, permissionID); err != nil {
		return err
	}

	return s.queries.DeletePermission(ctx, permissionID)
}

type GetPermissionsParams struct {
	ServiceVersionID uint
	FeatureVersionID *uint
	KeyID            *uint
	Variation        map[uint]string
}

type EntityPermissionDto struct {
	ID         uint               `json:"id" validate:"required"`
	UserID     *uint              `json:"userId"`
	UserName   *string            `json:"userName"`
	GroupID    *uint              `json:"groupId"`
	GroupName  *string            `json:"groupName"`
	Permission db.PermissionLevel `json:"permission" validate:"required"`
}

func (s *Service) GetPermissions(ctx context.Context, params GetPermissionsParams) ([]EntityPermissionDto, error) {
	serviceVersion, featureVersion, key, err := s.coreService.GetKeyOptional(ctx, params.ServiceVersionID, params.FeatureVersionID, params.KeyID)
	if err != nil {
		return nil, err
	}

	var featureID *uint
	var keyID *uint
	var variationContextID *uint
	if featureVersion != nil {
		featureID = &featureVersion.FeatureID
	}

	if key != nil {
		keyID = &key.ID
	}

	if len(params.Variation) > 0 {
		vcID, err := s.variationContextService.GetVariationContextID(ctx, params.Variation)
		if err != nil {
			return nil, core.NewDbError(err, "VariationContext")
		}

		variationContextID = &vcID
	}

	permissions, err := s.queries.GetPermissionsForEntity(ctx, db.GetPermissionsForEntityParams{
		ServiceID:          serviceVersion.ServiceID,
		FeatureID:          featureID,
		KeyID:              keyID,
		VariationContextID: variationContextID,
	})
	if err != nil {
		return nil, err
	}

	permissionDtos := make([]EntityPermissionDto, len(permissions))
	for i, permission := range permissions {
		permissionDtos[i] = EntityPermissionDto{
			ID:         permission.ID,
			UserID:     permission.UserID,
			UserName:   permission.UserName,
			GroupID:    permission.GroupID,
			GroupName:  permission.GroupName,
			Permission: permission.Permission,
		}
	}

	return permissionDtos, nil
}

type AddPermissionParams struct {
	UserID           *uint
	GroupID          *uint
	ServiceVersionID uint
	FeatureVersionID *uint
	KeyID            *uint
	Variation        map[uint]string
	Permission       db.PermissionLevel
}

func (s *Service) validateAddPermission(ctx context.Context, params AddPermissionParams, serviceVersion db.GetServiceVersionRow, featureID *uint, keyID *uint, variationContextID *uint) error {
	user := auth.GetUserFromContext(ctx)
	if user.GetPermissionForService(serviceVersion.ServiceID) < constants.PermissionAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to add a permission")
	}

	if (params.UserID == nil && params.GroupID == nil) || (params.UserID != nil && params.GroupID != nil) {
		return core.NewServiceError(core.ErrorCodeInvalidInput, "Either user or group must be provided, but not both")
	}

	if len(params.Variation) > 0 {
		if params.FeatureVersionID == nil || params.KeyID == nil {
			return core.NewServiceError(core.ErrorCodeInvalidInput, "Feature version and key must be provided if variation context is provided")
		}
	}

	if params.KeyID != nil && params.FeatureVersionID == nil {
		return core.NewServiceError(core.ErrorCodeInvalidInput, "Feature version must be provided if key is provided")
	}

	if !slices.Contains([]db.PermissionLevel{db.PermissionLevelAdmin, db.PermissionLevelEditor}, params.Permission) {
		return core.NewServiceError(core.ErrorCodeInvalidInput, "Invalid permission level")
	}

	if params.FeatureVersionID != nil && params.Permission == db.PermissionLevelAdmin {
		return core.NewServiceError(core.ErrorCodeInvalidInput, "Admin permission can only be set on the service level")
	}

	if params.UserID != nil {
		_, err := s.queries.GetUserByID(ctx, *params.UserID)
		if err != nil {
			return core.NewDbError(err, "User")
		}
	}

	if params.GroupID != nil {
		_, err := s.queries.GetGroupByID(ctx, *params.GroupID)
		if err != nil {
			return core.NewDbError(err, "Group")
		}
	}

	permission, err := s.queries.GetPermission(ctx, db.GetPermissionParams{
		UserID:             params.UserID,
		UserGroupID:        params.GroupID,
		ServiceID:          serviceVersion.ServiceID,
		FeatureID:          featureID,
		KeyID:              keyID,
		VariationContextID: variationContextID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	} else if err == nil {
		return core.NewServiceError(core.ErrorCodeInvalidInput, fmt.Sprintf("Permission already exists with level %s", permission.Permission))
	}

	return nil
}

func (s *Service) AddPermission(ctx context.Context, params AddPermissionParams) error {
	serviceVersion, featureVersion, key, err := s.coreService.GetKeyOptional(ctx, params.ServiceVersionID, params.FeatureVersionID, params.KeyID)
	if err != nil {
		return err
	}

	var featureID *uint
	var keyID *uint
	var variationContextID *uint
	kind := db.PermissionKindService

	if featureVersion != nil {
		featureID = &featureVersion.FeatureID
		kind = db.PermissionKindFeature
	}

	if key != nil {
		keyID = &key.ID
		kind = db.PermissionKindKey
	}

	if len(params.Variation) > 0 {
		variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx)
		if err != nil {
			return err
		}

		err = variationHierarchy.ValidateIDVariation(serviceVersion.ServiceTypeID, params.Variation)
		if err != nil {
			return err
		}

		vcID, err := s.variationContextService.GetVariationContextID(ctx, params.Variation)
		if err != nil {
			return err
		}

		variationContextID = &vcID
		kind = db.PermissionKindVariation
	}

	err = s.validateAddPermission(ctx, params, serviceVersion, featureID, keyID, variationContextID)
	if err != nil {
		return err
	}

	_, err = s.queries.CreatePermission(ctx, db.CreatePermissionParams{
		UserID:             params.UserID,
		UserGroupID:        params.GroupID,
		ServiceID:          serviceVersion.ServiceID,
		FeatureID:          featureID,
		KeyID:              keyID,
		VariationContextID: variationContextID,
		Permission:         params.Permission,
		Kind:               kind,
	})
	if err != nil {
		return err
	}

	return nil
}
