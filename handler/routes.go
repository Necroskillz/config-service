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
	changesetGroup := e.Group("/changesets")
	changesetGroup.GET("/:changeset_id", h.ChangesetDetail)

	keyGroup := featureGroup.Group("/:feature_version_id/keys")
	keyGroup.GET("/create", h.CreateKey)
	keyGroup.POST("", h.CreateKeySubmit)

	valueGroup := keyGroup.Group("/:key_id/values")
	valueGroup.GET("", h.ValueMatrix)
}
