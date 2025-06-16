package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	configserviceclient "github.com/necroskillz/config-service/go-client"
)

type JSONObj struct {
	Ltd string `json:"ltd"`
}

type CantoneseWagonsFeature struct {
	OverheadKantJSON             JSONObj
	AcidOnString                 string
	LustfulInstancesDecimal      float64
	MesopotamianMedicinesInteger int
	FortranBoltBoolean           bool
}

func (c *CantoneseWagonsFeature) FeatureName() string {
	return "StoicService.CantoneseWagonsFeature"
}

func main() {
	ctx := context.Background()

	slog.SetLogLoggerLevel(slog.LevelError)

	client := configserviceclient.New(configserviceclient.Config{
		Url: "localhost:50051",
		Services: map[string]int{
			"StoicService": 1,
		},
		PollingInterval:          15 * time.Second,
		SnapshotCleanupInterval:  1 * time.Minute,
		UnusedSnapshotExpiration: 1 * time.Minute,
	},
		configserviceclient.WithFeatures(
			&CantoneseWagonsFeature{},
		),
		configserviceclient.WithStaticVariation("env", "dev"),
		configserviceclient.WithStaticVariation("domain", "necroskillz.io"),
		configserviceclient.WithStaticVariation("product", "shop"),
		configserviceclient.WithChangesetOverrider(func(ctx context.Context) *uint32 {
			changesetId := uint32(1011)

			return &changesetId
		}),
		configserviceclient.WithLogging(func(ctx context.Context, level slog.Level, msg string, fields ...any) {
			slog.Log(ctx, level, msg, fields...)
		}),
		configserviceclient.WithProductionMode(false),
		configserviceclient.WithFallbackFileLocation("C:\\Devel\\config-service\\go-client\\example\\basic\\fallback"),
		configserviceclient.WithOverride("StoicService.CantoneseWagonsFeature", "LustfulInstancesDecimal", 5.5),
	)

	err := client.Start(ctx)
	if err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		err = client.Stop(ctx)
		if err != nil {
			log.Fatalf("Failed to stop client: %v", err)
		}
		os.Exit(0)
	}()

	for {
		feature := &CantoneseWagonsFeature{}
		err = client.BindFeature(ctx, feature)
		if err != nil {
			log.Fatalf("Failed to bind feature: %v", err)
		}

		fmt.Printf("%+v\n", feature)
		time.Sleep(5 * time.Second)
	}
}
