package grpc

import (
	"context"

	pb "github.com/necroskillz/config-service/grpc/gen"
	"github.com/necroskillz/config-service/services"
	"github.com/necroskillz/config-service/services/configuration"
	"github.com/necroskillz/config-service/services/core"
	"github.com/necroskillz/config-service/services/variation"
	"github.com/necroskillz/config-service/util/ptr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ConfigurationServer struct {
	pb.UnimplementedConfigServiceServer
	ConfigurationService      *configuration.Service
	VariationHierarchyService *variation.HierarchyService
}

func NewConfigurationServer(svc *services.Services) *ConfigurationServer {
	return &ConfigurationServer{
		ConfigurationService:      svc.ConfigurationService,
		VariationHierarchyService: svc.VariationHierarchyService,
	}
}

func (s *ConfigurationServer) GetNextChangesets(ctx context.Context, req *pb.GetNextChangesetsRequest) (*pb.GetNextChangesetsResponse, error) {
	serviceVersionSpecifiers, err := core.ParseServiceVersionSpecifiers(req.Services)
	if err != nil {
		return nil, ToGRPCError(err)
	}

	changesets, err := s.ConfigurationService.GetNextChangesets(ctx, serviceVersionSpecifiers, uint(req.AfterChangesetId))
	if err != nil {
		return nil, ToGRPCError(err)
	}

	changesetIds := make([]uint32, len(changesets))
	for i, changeset := range changesets {
		changesetIds[i] = uint32(changeset)
	}

	response := &pb.GetNextChangesetsResponse{
		ChangesetIds: changesetIds,
	}

	return response, nil
}

func makeVariationHierarchyPropertyValues(values []configuration.VariationHierarchyPropertyValueDto) []*pb.VariationHierarchyPropertyValue {
	dtos := make([]*pb.VariationHierarchyPropertyValue, len(values))
	for i, value := range values {
		dtos[i] = &pb.VariationHierarchyPropertyValue{
			Value:    value.Value,
			Children: makeVariationHierarchyPropertyValues(value.Children),
		}
	}

	return dtos
}

func (s *ConfigurationServer) GetVariationHierarchy(ctx context.Context, req *pb.GetVariationHierarchyRequest) (*pb.GetVariationHierarchyResponse, error) {
	serviceVersionSpecifiers, err := core.ParseServiceVersionSpecifiers(req.Services)
	if err != nil {
		return nil, ToGRPCError(err)
	}

	variationHierarchy, err := s.ConfigurationService.GetVariationHierarchy(ctx, serviceVersionSpecifiers)
	if err != nil {
		return nil, ToGRPCError(err)
	}

	properties := make([]*pb.VariationHierarchyProperty, len(variationHierarchy.Properties))
	for i, property := range variationHierarchy.Properties {
		properties[i] = &pb.VariationHierarchyProperty{
			Name:   property.Name,
			Values: makeVariationHierarchyPropertyValues(property.Values),
		}
	}

	response := &pb.GetVariationHierarchyResponse{
		Properties: properties,
	}

	return response, nil
}

func (s *ConfigurationServer) GetConfiguration(ctx context.Context, req *pb.GetConfigurationRequest) (*pb.GetConfigurationResponse, error) {
	serviceVersionSpecifiers, err := core.ParseServiceVersionSpecifiers(req.Services)
	if err != nil {
		return nil, ToGRPCError(err)
	}

	variationHierarchy, err := s.VariationHierarchyService.GetVariationHierarchy(ctx)
	if err != nil {
		return nil, ToGRPCError(err)
	}

	variation, err := variationHierarchy.GetVariationIDMap(req.Variation)
	if err != nil {
		return nil, ToGRPCError(err)
	}

	configuration, err := s.ConfigurationService.GetConfiguration(ctx, configuration.GetConfigurationParams{
		ServiceVersionSpecifiers: serviceVersionSpecifiers,
		ChangesetID:              ptr.To(uint(ptr.From(req.ChangesetId)), ptr.NilIfZero()),
		Mode:                     ptr.From(req.Mode),
		Variation:                variation,
	})
	if err != nil {
		return nil, ToGRPCError(err)
	}

	features := make([]*pb.Feature, len(configuration.Features))
	for i, feature := range configuration.Features {
		keys := make([]*pb.ConfigKey, len(feature.Keys))
		for j, key := range feature.Keys {

			values := make([]*pb.ConfigValue, len(key.Values))
			for k, value := range key.Values {
				values[k] = &pb.ConfigValue{
					Data:      value.Data,
					Variation: value.Variation,
					Rank:      int32(value.Rank),
				}
			}

			keys[j] = &pb.ConfigKey{
				Name:     key.Name,
				DataType: key.DataType,
				Values:   values,
			}
		}

		features[i] = &pb.Feature{
			Name: feature.Name,
			Keys: keys,
		}
	}

	response := &pb.GetConfigurationResponse{
		ChangesetId: uint32(configuration.ChangesetID),
		Features:    features,
	}

	if configuration.AppliedAt != nil {
		response.AppliedAt = timestamppb.New(*configuration.AppliedAt)
	}

	return response, nil
}
