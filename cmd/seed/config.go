package main

import (
	internalConfig "github.com/flash-go/files-service/internal/config"
	"github.com/flash-go/sdk/telemetry"
)

var envMap = map[string]string{
	"OTEL_COLLECTOR_GRPC":       telemetry.OtelCollectorGrpcOptKey,
	"OTEL_COLLECTOR_CA_CRT":     telemetry.OtelCollectorCaCrtOptKey,
	"OTEL_COLLECTOR_CLIENT_CRT": telemetry.OtelCollectorClientCrtOptKey,
	"OTEL_COLLECTOR_CLIENT_KEY": telemetry.OtelCollectorClientKeyOptKey,
	"USERS_SERVICE_NAME":        internalConfig.UsersServiceNameOptKey,
	"USERS_ADMIN_ROLE":          internalConfig.UsersAdminRoleOptKey,
	"STORE_LOCAL_ROOT_PATH":     internalConfig.StoreLocalRootPathOptKey,
}
