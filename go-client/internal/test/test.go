package test

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"slices"
	"testing"
	"time"

	grpcgen "github.com/necroskillz/config-service/go-client/grpc/gen"
	"go.nhat.io/grpcmock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/poll"
)

type TestConfigurationReponseBuilder struct {
	response *grpcgen.GetConfigurationResponse
	features map[string]*grpcgen.Feature
	keys     map[string]map[string]*grpcgen.ConfigKey
}

func NewTestConfigurationReponseBuilder() *TestConfigurationReponseBuilder {
	return &TestConfigurationReponseBuilder{
		response: &grpcgen.GetConfigurationResponse{},
		features: make(map[string]*grpcgen.Feature),
		keys:     make(map[string]map[string]*grpcgen.ConfigKey),
	}
}

func (b *TestConfigurationReponseBuilder) WithFeature(featureName string) *TestConfigurationReponseBuilder {
	if _, ok := b.features[featureName]; !ok {
		b.features[featureName] = &grpcgen.Feature{
			Name: featureName,
			Keys: make([]*grpcgen.ConfigKey, 0),
		}

		b.keys[featureName] = make(map[string]*grpcgen.ConfigKey)
		b.response.Features = append(b.response.Features, b.features[featureName])
	}

	return b
}

func (b *TestConfigurationReponseBuilder) WithoutFeature(featureName string) *TestConfigurationReponseBuilder {
	if _, ok := b.features[featureName]; ok {
		delete(b.features, featureName)
		delete(b.keys, featureName)

		b.response.Features = slices.DeleteFunc(b.response.Features, func(feature *grpcgen.Feature) bool {
			return feature.Name == featureName
		})
	}

	return b
}

func (b *TestConfigurationReponseBuilder) WithKey(featureName string, keyName string, dataType string) *TestConfigurationReponseBuilder {
	b.WithFeature(featureName)

	if _, ok := b.keys[featureName][keyName]; !ok {
		b.keys[featureName][keyName] = &grpcgen.ConfigKey{
			Name:     keyName,
			DataType: dataType,
			Values:   make([]*grpcgen.ConfigValue, 0),
		}

		b.features[featureName].Keys = append(b.features[featureName].Keys, b.keys[featureName][keyName])
	}

	return b
}

func (b *TestConfigurationReponseBuilder) WithoutKey(featureName string, keyName string) *TestConfigurationReponseBuilder {
	b.WithFeature(featureName)

	if _, ok := b.keys[featureName][keyName]; ok {
		delete(b.keys[featureName], keyName)
		b.features[featureName].Keys = slices.DeleteFunc(b.features[featureName].Keys, func(key *grpcgen.ConfigKey) bool {
			return key.Name == keyName
		})
	}

	return b
}

func (b *TestConfigurationReponseBuilder) WithDefaultValue(featureName string, keyName string, dataType string, data string) *TestConfigurationReponseBuilder {
	b.WithKey(featureName, keyName, dataType)

	b.keys[featureName][keyName].Values = append(b.keys[featureName][keyName].Values, &grpcgen.ConfigValue{
		Data: data,
	})

	return b
}

func (b *TestConfigurationReponseBuilder) WithDynamicVariationValue(featureName string, keyName string, dataType string, data string, variation map[string]string, rank int32) *TestConfigurationReponseBuilder {
	b.WithKey(featureName, keyName, dataType)

	b.keys[featureName][keyName].Values = append(b.keys[featureName][keyName].Values, &grpcgen.ConfigValue{
		Data:      data,
		Variation: variation,
		Rank:      rank,
	})

	return b
}

func (b *TestConfigurationReponseBuilder) WithChangesetId(changesetId uint32) *TestConfigurationReponseBuilder {
	b.response.ChangesetId = changesetId

	return b
}

func (b *TestConfigurationReponseBuilder) Response() *grpcgen.GetConfigurationResponse {
	return b.response
}

type TestVariationHierarchyResponseBuilder struct {
	response   *grpcgen.GetVariationHierarchyResponse
	properties map[string]*grpcgen.VariationHierarchyProperty
	values     map[string]map[string]*grpcgen.VariationHierarchyPropertyValue
}

func NewTestVariationHierarchyResponseBuilder() *TestVariationHierarchyResponseBuilder {
	return &TestVariationHierarchyResponseBuilder{
		response:   &grpcgen.GetVariationHierarchyResponse{},
		properties: make(map[string]*grpcgen.VariationHierarchyProperty),
		values:     make(map[string]map[string]*grpcgen.VariationHierarchyPropertyValue),
	}
}

func (b *TestVariationHierarchyResponseBuilder) WithProperty(propertyName string) *TestVariationHierarchyResponseBuilder {
	if _, ok := b.properties[propertyName]; !ok {
		b.properties[propertyName] = &grpcgen.VariationHierarchyProperty{
			Name:   propertyName,
			Values: make([]*grpcgen.VariationHierarchyPropertyValue, 0),
		}

		b.values[propertyName] = make(map[string]*grpcgen.VariationHierarchyPropertyValue)

		b.response.Properties = append(b.response.Properties, b.properties[propertyName])
	}

	return b
}

func (b *TestVariationHierarchyResponseBuilder) WithValue(propertyName string, value string) *TestVariationHierarchyResponseBuilder {
	b.WithProperty(propertyName)

	if _, ok := b.values[propertyName][value]; !ok {
		b.values[propertyName][value] = &grpcgen.VariationHierarchyPropertyValue{
			Value: value,
		}

		b.properties[propertyName].Values = append(b.properties[propertyName].Values, b.values[propertyName][value])
	}

	return b
}

func (b *TestVariationHierarchyResponseBuilder) WithChildValue(propertyName string, parent string, value string) *TestVariationHierarchyResponseBuilder {
	b.WithProperty(propertyName)

	if _, ok := b.values[propertyName][parent]; ok {
		b.values[propertyName][value] = &grpcgen.VariationHierarchyPropertyValue{
			Value: value,
		}

		b.values[propertyName][parent].Children = append(b.values[propertyName][parent].Children, b.values[propertyName][value])
	} else {
		panic(fmt.Sprintf("parent value %s not found for property %s", parent, propertyName))
	}

	return b
}

func (b *TestVariationHierarchyResponseBuilder) Response() *grpcgen.GetVariationHierarchyResponse {
	return b.response
}

type TestConfigClient struct {
	dialer grpcmock.ContextDialer
	server *grpcmock.Server
}

func NewTestConfigGRPC(t *testing.T) *TestConfigClient {
	server, dialer := grpcmock.MockServerWithBufConn(
		grpcmock.RegisterService(grpcgen.RegisterConfigServiceServer),
	)(t)

	return &TestConfigClient{
		dialer: dialer,
		server: server,
	}
}

func (c *TestConfigClient) Client() grpcgen.ConfigServiceClient {
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(c.dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	return grpcgen.NewConfigServiceClient(conn)
}

func (c *TestConfigClient) Server() *grpcmock.Server {
	return c.server
}

func RunCases[TC any](t *testing.T, run func(*testing.T, TC), testCases map[string]TC) {
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			run(t, tc)
		})
	}
}

func WaitForFile(t *testing.T, path string) {
	poll.WaitOn(t, poll.FileExists(path), poll.WithTimeout(5*time.Second), poll.WithDelay(5*time.Millisecond))
}

func WaitForFileUpdate(t *testing.T, path string) {
	stat, err := os.Stat(path)
	assert.NilError(t, err)

	modTime := stat.ModTime()

	poll.WaitOn(t, func(t poll.LogT) poll.Result {
		stat, err := os.Stat(path)
		if err != nil {
			return poll.Error(err)
		}

		if stat.ModTime().After(modTime) {
			return poll.Success()
		}

		return poll.Continue("waiting for file update")
	}, poll.WithTimeout(time.Second*5), poll.WithDelay(5*time.Millisecond))
}

type LogMessage struct {
	Level  slog.Level
	Msg    string
	Fields []any
}

type TestLogger struct {
	Logs  []LogMessage
	LogFn func(ctx context.Context, level slog.Level, msg string, fields ...any)
}

func NewTestLogger(t *testing.T) *TestLogger {
	logger := &TestLogger{
		Logs: make([]LogMessage, 0),
	}

	logFn := func(ctx context.Context, level slog.Level, msg string, fields ...any) {
		if level == slog.LevelDebug {
			t.Logf("DEBUG: %s %v", msg, fields)
			return
		}

		logger.Logs = append(logger.Logs, LogMessage{
			Level:  level,
			Msg:    msg,
			Fields: fields,
		})
	}

	logger.LogFn = logFn

	return logger
}

func WithLevel(level slog.Level) func(t *testing.T, log LogMessage) cmp.Comparison {
	return func(t *testing.T, log LogMessage) cmp.Comparison {
		return cmp.Equal(log.Level, level)
	}
}

func WithMessage(message string) func(t *testing.T, log LogMessage) cmp.Comparison {
	return func(t *testing.T, log LogMessage) cmp.Comparison {
		return cmp.Equal(log.Msg, message)
	}
}

func WithField(key string, compareFunc func(t *testing.T, value any) cmp.Comparison) func(t *testing.T, log LogMessage) cmp.Comparison {
	return func(t *testing.T, log LogMessage) cmp.Comparison {
		for i := 0; i < len(log.Fields); i += 2 {
			if log.Fields[i] == key {
				return compareFunc(t, log.Fields[i+1])
			}
		}

		panic(fmt.Sprintf("field %s not found", key))
	}
}

func (l *TestLogger) AssertLog(t *testing.T, options ...func(t *testing.T, log LogMessage) cmp.Comparison) {
	for _, logMsg := range l.Logs {
		for _, option := range options {
			if !option(t, logMsg)().Success() {
				t.Fatalf("log message does not match option: %v", logMsg)
			}
		}
	}
}
