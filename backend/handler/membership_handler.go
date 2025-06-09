package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/db"
	_ "github.com/necroskillz/config-service/services/core"
	"github.com/necroskillz/config-service/services/membership"
	"github.com/necroskillz/config-service/util/ptr"
)

// @Summary Get users and groups
// @Description Get list of users and groups
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Param name query string false "Name"
// @Param type query string false "Type" enum(user, group)
// @Success 200 {object} core.PaginatedResult[membership.MembershipObjectDto]
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /membership [get]
func (h *Handler) UsersAndGroups(c echo.Context) error {
	page := 1
	pageSize := 20
	var objectType string
	var name string

	err := echo.QueryParamsBinder(c).
		Int("page", &page).
		Int("pageSize", &pageSize).
		String("name", &name).
		String("type", &objectType).
		BindError()

	if err != nil {
		return ToHTTPError(err)
	}

	users, err := h.MembershipService.GetUsersAndGroups(c.Request().Context(), membership.UsersFilter{
		Page:     page,
		PageSize: pageSize,
		Name:     ptr.To(name, ptr.NilIfZero()),
		Type:     ptr.To(objectType, ptr.NilIfZero()),
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, users)
}

// @Summary Get a user
// @Description Get a user by ID
// @Produce json
// @Security BearerAuth
// @Param user_id path uint true "User ID"
// @Success 200 {object} membership.UserDto
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /membership/users/{user_id} [get]
func (h *Handler) GetUser(c echo.Context) error {
	var userID uint

	err := echo.PathParamsBinder(c).MustUint("user_id", &userID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	user, err := h.MembershipService.GetUser(c.Request().Context(), userID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, user)
}

type CreateUserRequest struct {
	Username            string `json:"username" validate:"required"`
	Password            string `json:"password" validate:"required"`
	GlobalAdministrator bool   `json:"globalAdministrator" validate:"required"`
}

// @Summary Create a user
// @Description Create a new user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body CreateUserRequest true "User"
// @Success 200 {object} CreateResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 422 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /membership/users [post]
func (h *Handler) CreateUser(c echo.Context) error {
	var request CreateUserRequest

	err := c.Bind(&request)
	if err != nil {
		return ToHTTPError(err)
	}

	userID, err := h.MembershipService.CreateUser(c.Request().Context(), membership.CreateUserParams{
		Username:            request.Username,
		Password:            request.Password,
		GlobalAdministrator: request.GlobalAdministrator,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, NewCreateResponse(userID))
}

type UpdateUserRequest struct {
	GlobalAdministrator bool `json:"globalAdministrator" validate:"required"`
}

// @Summary Update a user
// @Description Update a user by ID
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path uint true "User ID"
// @Param user body UpdateUserRequest true "User"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /membership/users/{user_id} [put]
func (h *Handler) UpdateUser(c echo.Context) error {
	var userID uint

	err := echo.PathParamsBinder(c).MustUint("user_id", &userID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	var request UpdateUserRequest
	err = c.Bind(&request)
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.MembershipService.UpdateUser(c.Request().Context(), userID, membership.UpdateUserParams{
		GlobalAdministrator: request.GlobalAdministrator,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Delete a user
// @Description Delete a user by ID
// @Produce json
// @Security BearerAuth
// @Param user_id path uint true "User ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /membership/users/{user_id} [delete]
func (h *Handler) DeleteUser(c echo.Context) error {
	var userID uint

	err := echo.PathParamsBinder(c).MustUint("user_id", &userID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.MembershipService.DeleteUser(c.Request().Context(), userID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Get a group
// @Description Get a group by ID
// @Produce json
// @Security BearerAuth
// @Param group_id path uint true "Group ID"
// @Success 200 {object} membership.GroupDto
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /membership/groups/{group_id} [get]
func (h *Handler) GetGroup(c echo.Context) error {
	var groupID uint

	err := echo.PathParamsBinder(c).MustUint("group_id", &groupID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	group, err := h.MembershipService.GetGroup(c.Request().Context(), groupID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, group)
}

// @Summary Get group users
// @Description Get group users by ID
// @Produce json
// @Security BearerAuth
// @Param group_id path uint true "Group ID"
// @Param page query int false "Page" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} core.PaginatedResult[membership.GroupUserDto]
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /membership/groups/{group_id}/users [get]
func (h *Handler) GetGroupUsers(c echo.Context) error {
	page := 1
	pageSize := 20
	var groupID uint

	err := echo.PathParamsBinder(c).MustUint("group_id", &groupID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	err = echo.QueryParamsBinder(c).
		Int("page", &page).
		Int("pageSize", &pageSize).
		BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	groupUsers, err := h.MembershipService.GetGroupUsers(c.Request().Context(), membership.GetGroupUsersParams{
		GroupID:  groupID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, groupUsers)
}

type CreateGroupRequest struct {
	Name string `json:"name" validate:"required"`
}

// @Summary Create a group
// @Description Create a new group
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param group body CreateGroupRequest true "Group"
// @Success 200 {object} CreateResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 422 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /membership/groups [post]
func (h *Handler) CreateGroup(c echo.Context) error {
	var request CreateGroupRequest

	err := c.Bind(&request)
	if err != nil {
		return ToHTTPError(err)
	}

	groupID, err := h.MembershipService.CreateGroup(c.Request().Context(), membership.CreateGroupParams{
		Name: request.Name,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, NewCreateResponse(groupID))
}

// @Summary Delete a group
// @Description Delete a group by ID
// @Produce json
// @Security BearerAuth
// @Param group_id path uint true "Group ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /membership/groups/{group_id} [delete]
func (h *Handler) DeleteGroup(c echo.Context) error {
	var groupID uint

	err := echo.PathParamsBinder(c).MustUint("group_id", &groupID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.MembershipService.DeleteGroup(c.Request().Context(), groupID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Add a user to a group
// @Description Add a user to a group
// @Produce json
// @Security BearerAuth
// @Param user_id path uint true "User ID"
// @Param group_id path uint true "Group ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /membership/groups/{group_id}/users/{user_id} [post]
func (h *Handler) AddUserToGroup(c echo.Context) error {
	var userID uint
	var groupID uint

	err := echo.PathParamsBinder(c).MustUint("user_id", &userID).MustUint("group_id", &groupID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.MembershipService.AddUserToGroup(c.Request().Context(), userID, groupID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Remove a user from a group
// @Description Remove a user from a group
// @Produce json
// @Security BearerAuth
// @Param user_id path uint true "User ID"
// @Param group_id path uint true "Group ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /membership/groups/{group_id}/users/{user_id} [delete]
func (h *Handler) RemoveUserFromGroup(c echo.Context) error {
	var userID uint
	var groupID uint

	err := echo.PathParamsBinder(c).MustUint("user_id", &userID).MustUint("group_id", &groupID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.MembershipService.RemoveUserFromGroup(c.Request().Context(), userID, groupID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Get permissions
// @Description Get permissions for a service, feature, key, or variation
// @Produce json
// @Security BearerAuth
// @Param serviceVersionId query uint true "Service version ID"
// @Param featureVersionId query uint false "Feature version ID"
// @Param keyId query uint false "Key ID"
// @Param variation[] query []string false "Variation" example(1:prod) collectionFormat(multi)
// @Success 200 {object} []membership.EntityPermissionDto
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /membership/permissions [get]
func (h *Handler) GetPermissions(c echo.Context) error {
	var serviceVersionID uint
	var featureVersionID uint
	var keyID uint

	err := echo.QueryParamsBinder(c).
		Uint("serviceVersionId", &serviceVersionID).
		Uint("featureVersionId", &featureVersionID).
		Uint("keyId", &keyID).
		BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	variation, err := h.GetVariationFromQueryIds(c)
	if err != nil {
		return ToHTTPError(err)
	}

	permissions, err := h.MembershipService.GetPermissions(c.Request().Context(), membership.GetPermissionsParams{
		ServiceVersionID: serviceVersionID,
		FeatureVersionID: ptr.To(featureVersionID, ptr.NilIfZero()),
		KeyID:            ptr.To(keyID, ptr.NilIfZero()),
		Variation:        variation,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, permissions)
}

type AddPermissionRequest struct {
	UserID           *uint              `json:"userId"`
	GroupID          *uint              `json:"groupId"`
	ServiceVersionID uint               `json:"serviceVersionId" validate:"required"`
	FeatureVersionID *uint              `json:"featureVersionId"`
	KeyID            *uint              `json:"keyId"`
	Variation        map[uint]string    `json:"variation"`
	Permission       db.PermissionLevel `json:"permission" validate:"required"`
}

// @Summary Add a permission
// @Description Add a permission to a user or group
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param permission body AddPermissionRequest true "Permission"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 422 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /membership/permissions [post]
func (h *Handler) AddPermission(c echo.Context) error {
	var request AddPermissionRequest

	err := c.Bind(&request)
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.MembershipService.AddPermission(c.Request().Context(), membership.AddPermissionParams{
		UserID:           request.UserID,
		GroupID:          request.GroupID,
		ServiceVersionID: request.ServiceVersionID,
		FeatureVersionID: request.FeatureVersionID,
		KeyID:            request.KeyID,
		Variation:        request.Variation,
		Permission:       request.Permission,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Remove a permission from a group
// @Description Remove a permission from a group
// @Produce json
// @Security BearerAuth
// @Param permission_id path uint true "Permission ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /membership/permissions/{permission_id} [delete]
func (h *Handler) RemovePermission(c echo.Context) error {
	var permissionID uint

	err := echo.PathParamsBinder(c).MustUint("permission_id", &permissionID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.MembershipService.RemovePermission(c.Request().Context(), permissionID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
