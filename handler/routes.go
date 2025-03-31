package handler

import "github.com/labstack/echo/v4"

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.GET("/", h.Services)
	e.GET("/login", h.LoginPage)
	e.POST("/login", h.Login)
	e.POST("/logout", h.Logout)

	serviceGroup := e.Group("/services")
	serviceGroup.GET("", h.Services)
	serviceGroup.GET("/create", h.CreateService)
	serviceGroup.POST("", h.CreateServiceSubmit)
	serviceGroup.GET("/:service_version_id", h.ServiceDetail)

	featureGroup := serviceGroup.Group("/:service_version_id/features")
	featureGroup.GET("/create", h.CreateFeature)
	featureGroup.POST("", h.CreateFeatureSubmit)
	featureGroup.GET("/:feature_version_id", h.FeatureDetail)

	keyGroup := featureGroup.Group("/:feature_version_id/keys")
	keyGroup.GET("/create", h.CreateKey)
	keyGroup.POST("", h.CreateKeySubmit)

	valueGroup := keyGroup.Group("/:key_id/values")
	valueGroup.GET("", h.ValueMatrix)
	valueGroup.POST("", h.CreateValueSubmit)
	valueGroup.DELETE("/:value_id", h.DeleteValueSubmit)

	changesetsGroup := e.Group("/changesets")

	changesetGroup := changesetsGroup.Group("/:changeset_id")
	changesetGroup.GET("", h.ChangesetDetail)
	changesetGroup.PUT("/apply", h.ApplyChangeset)
	changesetGroup.PUT("/commit", h.CommitChangeset)
	changesetGroup.PUT("/reopen", h.ReopenChangeset)
	changesetGroup.PUT("/stash", h.StashChangeset)
	changesetGroup.DELETE("", h.DiscardChangeset)
}
