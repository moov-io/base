package testdb

import (
	"context"
	"fmt"
	"os"

	spannerdb "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	"cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
	"github.com/moov-io/base"
	"github.com/moov-io/base/database"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Must be called if using the docker spanner emulator
func SetSpannerEmulator(hostOverride *string) {
	host := "localhost:9010"
	if hostOverride != nil {
		host = *hostOverride
	}

	os.Setenv("SPANNER_EMULATOR_HOST", host)
}

func NewSpannerDatabase(databaseName string, spannerCfg *database.SpannerConfig) (database.DatabaseConfig, error) {
	if spannerCfg == nil {
		spannerCfg = &database.SpannerConfig{
			Project:  "proj" + base.ID()[0:26],
			Instance: "test",
		}
	}

	cfg := database.DatabaseConfig{
		DatabaseName: databaseName,
		Spanner:      spannerCfg,
	}

	if err := createInstance(cfg.Spanner); err != nil {
		return cfg, err
	}

	if err := createDatabase(cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func createInstance(cfg *database.SpannerConfig) error {
	ctx := context.Background()
	instanceAdmin, err := instance.NewInstanceAdminClient(ctx)
	if err != nil {
		return err
	}
	defer instanceAdmin.Close()

	op, err := instanceAdmin.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
		Parent:     fmt.Sprintf("projects/%s", cfg.Project),
		InstanceId: cfg.Instance,
		Instance: &instancepb.Instance{
			Config:      fmt.Sprintf("projects/%s/instanceConfigs/%s", cfg.Project, "emulator-config"),
			DisplayName: cfg.Instance,
			NodeCount:   1,
		},
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return nil
		}
		return fmt.Errorf("could not create instance %s: %v", fmt.Sprintf("projects/%s/instances/%s", cfg.Project, cfg.Instance), err)
	}

	// Wait for the instance creation to finish.
	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("waiting for instance creation to finish failed: %v", err)
	}

	return nil
}

func createDatabase(cfg database.DatabaseConfig) error {
	ctx := context.Background()
	databaseAdminClient, err := spannerdb.NewDatabaseAdminClient(ctx)
	if err != nil {
		return err
	}
	defer databaseAdminClient.Close()

	opDB, err := databaseAdminClient.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
		Parent:          fmt.Sprintf("projects/%s/instances/%s", cfg.Spanner.Project, cfg.Spanner.Instance),
		CreateStatement: fmt.Sprintf("CREATE DATABASE `%s`", cfg.DatabaseName),
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return nil
		}
		return err
	}

	// Wait for the database creation to finish.
	if _, err := opDB.Wait(ctx); err != nil {
		return fmt.Errorf("waiting for database creation to finish failed: %v", err)
	}

	return nil
}
