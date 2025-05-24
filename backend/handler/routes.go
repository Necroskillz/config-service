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

	usersGroup := apiGroup.Group("/users")
	usersGroup.GET("", h.Users)
	usersGroup.POST("", h.CreateUser)
	usersGroup.GET("/:user_id", h.GetUser)
	usersGroup.PUT("/:user_id", h.UpdateUser)

	servicesGroup := apiGroup.Group("/services")
	servicesGroup.GET("", h.Services)
	servicesGroup.POST("", h.CreateService)
	servicesGroup.GET("/name-taken/:name", h.IsServiceNameTaken)

	serviceGroup := servicesGroup.Group("/:service_version_id")
	serviceGroup.PUT("", h.UpdateService)
	serviceGroup.PUT("/publish", h.PublishServiceVersion)
	serviceGroup.GET("", h.Service)
	serviceGroup.GET("/versions", h.ServiceVersions)
	serviceGroup.POST("/versions", h.CreateServiceVersion)

	featuresGroup := serviceGroup.Group("/features")
	featuresGroup.GET("", h.Features)
	featuresGroup.GET("/linkable", h.LinkableFeatures)
	featuresGroup.POST("", h.CreateFeature)
	featuresGroup.GET("/name-taken/:name", h.IsFeatureNameTaken)

	featureGroup := featuresGroup.Group("/:feature_version_id")
	featureGroup.GET("", h.Feature)
	featureGroup.PUT("", h.UpdateFeature)
	featureGroup.GET("/versions", h.FeatureVersions)
	featureGroup.POST("/versions", h.CreateFeatureVersion)
	featureGroup.POST("/link", h.LinkFeatureVersion)
	featureGroup.DELETE("/unlink", h.UnlinkFeatureVersion)

	keysGroup := featureGroup.Group("/keys")
	keysGroup.GET("", h.Keys)
	keysGroup.POST("", h.CreateKey)
	keysGroup.GET("/name-taken/:name", h.IsKeyNameTaken)

	keyGroup := keysGroup.Group("/:key_id")
	keyGroup.GET("", h.Key)
	keyGroup.PUT("", h.UpdateKey)
	keyGroup.DELETE("", h.DeleteKey)

	valuesGroup := keyGroup.Group("/values")
	valuesGroup.GET("", h.Values)
	valuesGroup.POST("", h.CreateValue)
	valuesGroup.GET("/can-add", h.CanAddValue)

	valueGroup := valuesGroup.Group("/:value_id")
	valueGroup.PUT("", h.UpdateValue)
	valueGroup.DELETE("", h.DeleteValue)
	valueGroup.GET("/can-edit", h.CanEditValue)

	serviceTypesGroup := apiGroup.Group("/service-types")
	serviceTypesGroup.GET("", h.GetServiceTypes)
	serviceTypesGroup.POST("", h.CreateServiceType)

	serviceTypeGroup := serviceTypesGroup.Group("/:service_type_id")
	serviceTypeGroup.GET("", h.GetServiceType)
	serviceTypeGroup.DELETE("", h.DeleteServiceType)
	serviceTypeGroup.GET("/variation-properties", h.GetProperties)
	serviceTypeGroup.POST("/variation-properties", h.LinkVariationPropertyToServiceType)
	serviceTypeGroup.DELETE("/variation-properties/:variation_property_id", h.UnlinkVariationPropertyFromServiceType)
	serviceTypeGroup.PUT("/variation-properties/:variation_property_id/priority", h.UpdateServiceTypeVariationPropertyPriority)

	valueTypeGroup := apiGroup.Group("/value-types")
	valueTypeGroup.GET("", h.GetValueTypes)
	valueTypeGroup.GET("/:value_type_id", h.GetValueType)

	changesetsGroup := apiGroup.Group("/changesets")
	changesetsGroup.GET("", h.Changesets)
	changesetsGroup.GET("/current", h.GetCurrentChangesetInfo)
	changesetsGroup.GET("/approvable-count", h.GetApprovableChangesetCount)

	changesetGroup := changesetsGroup.Group("/:changeset_id")
	changesetGroup.GET("", h.Changeset)
	changesetGroup.PUT("/apply", h.ApplyChangeset)
	changesetGroup.PUT("/commit", h.CommitChangeset)
	changesetGroup.PUT("/reopen", h.ReopenChangeset)
	changesetGroup.PUT("/stash", h.StashChangeset)
	changesetGroup.DELETE("", h.DiscardChangeset)
	changesetGroup.POST("/comment", h.AddComment)

	changesetGroup.DELETE("/changes/:change_id", h.DiscardChange)

	variationPropertiesGroup := apiGroup.Group("/variation-properties")
	variationPropertiesGroup.GET("", h.VariationProperties)
	variationPropertiesGroup.POST("", h.CreateVariationProperty)
	variationPropertiesGroup.GET("/name-taken/:name", h.IsVariationPropertyNameTaken)
	variationPropertiesGroup.PUT("/:property_id", h.UpdateVariationProperty)

	variationPropertyGroup := variationPropertiesGroup.Group("/:property_id")
	variationPropertyGroup.GET("", h.VariationProperty)
	variationPropertyGroup.PUT("", h.UpdateVariationProperty)
	variationPropertyGroup.DELETE("", h.DeleteVariationProperty)
	variationPropertyGroup.GET("/value-taken/:value", h.IsVariationPropertyValueTaken)
	variationPropertyGroup.POST("/values", h.CreateVariationPropertyValue)

	variationPropertyValueGroup := variationPropertyGroup.Group("/values/:value_id")
	variationPropertyValueGroup.DELETE("", h.DeleteVariationPropertyValue)
	variationPropertyValueGroup.PUT("/order", h.UpdateVariationPropertyValueOrder)
	variationPropertyValueGroup.PUT("/archive", h.ArchiveVariationPropertyValue)
	variationPropertyValueGroup.PUT("/unarchive", h.UnarchiveVariationPropertyValue)

	configurationGroup := apiGroup.Group("/configuration")
	configurationGroup.GET("", h.GetConfiguration)
	configurationGroup.GET("/changesets", h.GetNextChangesets)
}
