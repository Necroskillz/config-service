package handler

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/service"
	"github.com/necroskillz/config-service/views/components"
)

func MakeSelectOptions[T any](items []T, fn func(item T) (uint, string)) []components.SelectOption {
	options := []components.SelectOption{}
	for _, item := range items {
		value, text := fn(item)
		options = append(options, components.SelectOption{Value: fmt.Sprintf("%d", value), Text: text})
	}
	return options
}

func getVariation(serviceTypeID uint, variationHierarchy *service.VariationHierarchy, getter func(string) string) map[string]string {
	properties := variationHierarchy.GetProperties(serviceTypeID)
	variation := make(map[string]string)

	for _, property := range properties {
		propertyValue := getter(property.Name)
		if propertyValue != "" && propertyValue != "any" {
			variation[property.Name] = propertyValue
		}
	}

	return variation
}

func GetVariationFromForm(c echo.Context, serviceTypeID uint, variationHierarchy *service.VariationHierarchy) map[string]string {
	return getVariation(serviceTypeID, variationHierarchy, func(name string) string {
		return c.FormValue(name)
	})
}

func GetVariationFromQuery(c echo.Context, serviceTypeID uint, variationHierarchy *service.VariationHierarchy) map[string]string {
	return getVariation(serviceTypeID, variationHierarchy, func(name string) string {
		return c.QueryParam(name)
	})
}
