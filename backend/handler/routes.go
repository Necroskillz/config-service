package handler

import "github.com/labstack/echo/v4"

// @title Config Service API
// @version 1.0
// @description This is the API for the Config Service
// @host localhost:1323
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func (h *Handler) RegisterRoutes(e *echo.Echo) {
	apiGroup := e.Group("/api")

	authGroup := apiGroup.Group("/auth")
	authGroup.POST("/login", h.Login)
	authGroup.POST("/refresh_token", h.RefreshToken)
	authGroup.GET("/user", h.User)

	serviceGroup := apiGroup.Group("/services")
	serviceGroup.GET("", h.Services)
	serviceGroup.POST("", h.CreateService)
	serviceGroup.PUT("/:service_version_id", h.UpdateService)
	serviceGroup.PUT("/:service_version_id/publish", h.PublishServiceVersion)
	serviceGroup.GET("/:service_version_id", h.Service)
	serviceGroup.GET("/:service_version_id/versions", h.ServiceVersions)
	serviceGroup.GET("/name-taken/:name", h.IsServiceNameTaken)

	featureGroup := serviceGroup.Group("/:service_version_id/features")
	featureGroup.GET("", h.Features)
	featureGroup.GET("/linkable", h.LinkableFeatures)
	featureGroup.POST("", h.CreateFeature)
	featureGroup.GET("/:feature_version_id", h.Feature)
	featureGroup.GET("/:feature_version_id/versions", h.FeatureVersions)
	featureGroup.POST("/:feature_version_id/link", h.LinkFeatureVersion)
	featureGroup.DELETE("/:feature_version_id/unlink", h.UnlinkFeatureVersion)
	featureGroup.GET("/name-taken/:name", h.IsFeatureNameTaken)

	keyGroup := featureGroup.Group("/:feature_version_id/keys")
	keyGroup.GET("", h.Keys)
	keyGroup.GET("/create", h.CreateKey)
	keyGroup.POST("", h.CreateKey)
	keyGroup.GET("/:key_id", h.Key)
	keyGroup.GET("/name-taken/:name", h.IsKeyNameTaken)

	serviceTypeGroup := apiGroup.Group("/service-types")
	serviceTypeGroup.GET("", h.GetServiceTypes)
	serviceTypeGroup.GET("/:service_type_id/variation-properties", h.GetProperties)

	valueTypeGroup := apiGroup.Group("/value-types")
	valueTypeGroup.GET("", h.GetValueTypes)

	valueGroup := keyGroup.Group("/:key_id/values")
	valueGroup.GET("", h.Values)
	valueGroup.PUT("/:value_id", h.UpdateValue)
	valueGroup.POST("", h.CreateValue)
	valueGroup.DELETE("/:value_id", h.DeleteValue)
	valueGroup.GET("/can-add", h.CanAddValue)
	valueGroup.GET("/:value_id/can-edit", h.CanEditValue)

	changesetsGroup := apiGroup.Group("/changesets")
	changesetsGroup.GET("/current", h.GetCurrentChangesetInfo)

	changesetGroup := changesetsGroup.Group("/:changeset_id")
	changesetGroup.GET("", h.Changeset)
	changesetGroup.PUT("/apply", h.ApplyChangeset)
	changesetGroup.PUT("/commit", h.CommitChangeset)
	changesetGroup.PUT("/reopen", h.ReopenChangeset)
	changesetGroup.PUT("/stash", h.StashChangeset)
	changesetGroup.DELETE("", h.DiscardChangeset)
	changesetGroup.POST("/comment", h.AddComment)

}
